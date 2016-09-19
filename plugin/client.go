package rpc2

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"time"
)

type (
	Client struct {
		*rpc.Client
		srvAddr            string
		exponentialBackoff []time.Duration
		clientCodecFunc    ClientCodecFunc
	}

	// ClientCodecFunc is used to create a rpc.ClientCodecFunc from net.Conn.
	ClientCodecFunc func(io.ReadWriteCloser) rpc.ClientCodec
)

var defaultExponentialBackoff = func() []time.Duration {
	var ds []time.Duration
	for i := uint(0); i < 4; i++ {
		ds = append(ds, time.Duration(0.75e9*2<<i))
	}
	return ds
}()

func NewClient(srvAddr string, clientCodecFunc ClientCodecFunc) *Client {
	if clientCodecFunc == nil {
		clientCodecFunc = func(conn io.ReadWriteCloser) rpc.ClientCodec {
			encBuf := bufio.NewWriter(conn)
			return &gobClientCodec{
				rwc:    conn,
				dec:    gob.NewDecoder(conn),
				enc:    gob.NewEncoder(encBuf),
				encBuf: encBuf,
			}
		}
	}
	return &Client{
		srvAddr:            srvAddr,
		exponentialBackoff: defaultExponentialBackoff,
		clientCodecFunc:    clientCodecFunc,
	}
}

func (client *Client) dialTCP() (net.Conn, error) {
	var (
		conn net.Conn
		err  error
	)
	for _, d := range client.exponentialBackoff {
		conn, err = net.DialTimeout("tcp", client.srvAddr, d)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("ConnectError: %s", err.Error())
	}
	return conn, nil
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) Call(rpcname string, args interface{}, reply interface{}) error {
	if client.Client == nil {
		conn, err := client.dialTCP()
		if err != nil {
			return err
		}

		codec := client.clientCodecFunc(conn)

		client.Client = rpc.NewClientWithCodec(codec)
	}

	return client.Client.Call(rpcname, args, reply)
}

// Close closes the connection
func (client *Client) Close() error {
	if client.Client != nil {
		err := client.Client.Close()
		client.Client = nil
		return err
	}
	return nil
}
