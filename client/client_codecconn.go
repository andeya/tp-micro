package client

import (
	"io"
	"net"
	"net/rpc"
)

type (
	// ClientCodecConn net.Conn with ClientCodec
	ClientCodecConn interface {
		// Conn
		net.Conn
		SetConn(net.Conn)
		GetConn() net.Conn

		// ClientCodec

		// WriteRequest must be safe for concurrent use by multiple goroutines.
		WriteRequest(*rpc.Request, interface{}) error
		ReadResponseHeader(*rpc.Response) error
		ReadResponseBody(interface{}) error

		GetClientCodec() rpc.ClientCodec

		//Must ensure that both Conn and ClientCodecFunc are not nil
		SetClientCodec(ClientCodecFunc)
	}

	// ClientCodecFunc is used to create a rpc.ClientCodec from net.Conn.
	ClientCodecFunc func(io.ReadWriteCloser) rpc.ClientCodec

	clientCodecConn struct {
		net.Conn
		rpc.ClientCodec
	}
)

// NewClientCodecConn get a ClientCodecConn.
func NewClientCodecConn(conn net.Conn) ClientCodecConn {
	return &clientCodecConn{Conn: conn}
}

func (conn *clientCodecConn) SetConn(c net.Conn) {
	conn.Conn = c
}

func (conn *clientCodecConn) GetConn() net.Conn {
	return conn.Conn
}

// SetClientCodec must ensure that both Conn and ClientCodecFunc are not nil
func (conn *clientCodecConn) SetClientCodec(fn ClientCodecFunc) {
	if fn != nil && conn.Conn != nil {
		conn.ClientCodec = fn(conn.Conn)
	}
}

func (conn *clientCodecConn) GetClientCodec() rpc.ClientCodec {
	return conn.ClientCodec
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (conn *clientCodecConn) Close() error {
	var err error
	if conn.ClientCodec != nil {
		err = conn.ClientCodec.Close()
	} else {
		err = conn.Conn.Close()
	}
	return err
}
