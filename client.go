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

type gobClientCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
}

func (c *gobClientCodec) WriteRequest(r *rpc.Request, body interface{}) (err error) {
	if err = timeoutCoder(c.enc.Encode, r, "client write request"); err != nil {
		return
	}
	if err = timeoutCoder(c.enc.Encode, body, "client write request body"); err != nil {
		return
	}
	return c.encBuf.Flush()
}

func (c *gobClientCodec) ReadResponseHeader(r *rpc.Response) error {
	return c.dec.Decode(r)
}

func (c *gobClientCodec) ReadResponseBody(body interface{}) error {
	return c.dec.Decode(body)
}

func (c *gobClientCodec) Close() error {
	return c.rwc.Close()
}

var exponentialBackoff = func() []time.Duration {
	var ds []time.Duration
	for i := uint(0); i < 4; i++ {
		ds = append(ds, time.Duration(0.75e9*2<<i))
	}
	return ds
}()

// Call invokes the named function, waits for it to complete, and returns its error status.
func Call(srvAddr string, rpcname string, args interface{}, reply interface{}) error {
	var (
		conn net.Conn
		err  error
	)
	for _, d := range exponentialBackoff {
		conn, err = net.DialTimeout("tcp", srvAddr, d)
		// fmt.Println(d)
		if err == nil {
			break
		}
	}
	if err != nil {
		return fmt.Errorf("ConnectError: %s", err.Error())
	}
	encBuf := bufio.NewWriter(conn)
	codec := &gobClientCodec{conn, gob.NewDecoder(conn), gob.NewEncoder(encBuf), encBuf}
	c := rpc.NewClientWithCodec(codec)
	err = c.Call(rpcname, args, reply)
	errc := c.Close()
	if err != nil && errc != nil {
		return fmt.Errorf("%s %s", err, errc)
	}
	if err != nil {
		return err
	} else {
		return errc
	}
}
