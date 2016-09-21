package rpc2

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"time"
)

type (
	Dialer struct {
		network            string
		address            string
		exponentialBackoff []time.Duration
		clientCodecFunc    ClientCodecFunc
	}

	IClient interface {
		Call(serviceMethod string, args interface{}, reply interface{}) error
		Go(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call
	}

	// ClientCodecFunc is used to create a rpc.ClientCodecFunc from net.Conn.
	ClientCodecFunc func(io.ReadWriteCloser) rpc.ClientCodec

	CallFunc func(IClient) error
)

var (
	defaultExponentialBackoff = func() []time.Duration {
		var ds []time.Duration
		for i := uint(0); i < 4; i++ {
			ds = append(ds, time.Duration(0.75e9*2<<i))
		}
		return ds
	}()
)

func NewDialer(network, address string, clientCodecFunc ClientCodecFunc) *Dialer {
	if clientCodecFunc == nil {
		clientCodecFunc = NewGobClientCodec
	}
	return &Dialer{
		network:            network,
		address:            address,
		exponentialBackoff: defaultExponentialBackoff,
		clientCodecFunc:    clientCodecFunc,
	}
}

// NewDefaultDialer returns a new default Dialer.
func NewDefaultDialer(network, address string) *Dialer {
	return NewDialer(network, address, nil)
}

// Dial connects to an RPC server at the setted network address.
func (dialer *Dialer) Dial() (*rpc.Client, error) {
	var (
		conn net.Conn
		err  error
	)
	for _, d := range dialer.exponentialBackoff {
		conn, err = net.DialTimeout(dialer.network, dialer.address, d)
		if err == nil {
			return rpc.NewClientWithCodec(dialer.clientCodecFunc(conn)), nil
		}
	}
	return nil, errors.New("dial error: " + err.Error())
}

// Remote is shortcut method of rpc calling remotely.
func (dialer *Dialer) Remote(callfunc CallFunc) error {
	client, err := dialer.Dial()
	if err != nil {
		return err
	}
	defer client.Close()
	return callfunc(client)
}

// Dial connects to an RPC server at the specified network address.
func Dial(network, address string) (*rpc.Client, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}

// DialHTTP connects to an HTTP RPC server at the specified network address
// listening on the default HTTP RPC path.
func DialHTTP(network, address string) (*rpc.Client, error) {
	return DialHTTPPath(network, address, rpc.DefaultRPCPath)
}

// DialHTTPPath connects to an HTTP RPC server
// at the specified network address and path.
func DialHTTPPath(network, address, path string) (*rpc.Client, error) {
	var err error
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	io.WriteString(conn, "CONNECT "+path+" HTTP/1.0\n\n")

	// Require successful HTTP response
	// before switching to RPC protocol.
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return NewClient(conn), nil
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	conn.Close()
	return nil, &net.OpError{
		Op:   "dial-http",
		Net:  network + " " + address,
		Addr: nil,
		Err:  err,
	}
}

// NewClient returns a new Client to handle requests to the
// set of services at the other end of the connection.
// It adds a buffer to the write side of the connection so
// the header and payload are sent as a unit.
func NewClient(conn io.ReadWriteCloser) *rpc.Client {
	return rpc.NewClientWithCodec(NewGobClientCodec(conn))
}

// NewClientWithCodec is like NewClient but uses the specified
// codec to encode requests and decode responses.
func NewClientWithCodec(codec rpc.ClientCodec) *rpc.Client {
	return rpc.NewClientWithCodec(codec)
}
