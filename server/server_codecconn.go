package server

import (
	"io"
	"net"
	"net/rpc"
)

type (
	// ServerCodecConn net.Conn with ServerCodec
	ServerCodecConn interface {
		// Conn
		net.Conn
		SetConn(net.Conn)
		GetConn() net.Conn

		// ServerCodec
		ReadRequestHeader(*rpc.Request) error
		ReadRequestBody(interface{}) error
		// WriteResponse must be safe for concurrent use by multiple goroutines.
		WriteResponse(*rpc.Response, interface{}) error

		GetServerCodec() rpc.ServerCodec

		//Must ensure that both Conn and ServerCodecFunc are not nil
		SetServerCodec(ServerCodecFunc)
	}

	// ServerCodecFunc is used to create a ServerCodec from io.ReadWriteCloser.
	ServerCodecFunc func(io.ReadWriteCloser) rpc.ServerCodec

	serverCodecConn struct {
		net.Conn
		rpc.ServerCodec
	}
)

// NewServerCodecConn get a ServerCodecConn.
func NewServerCodecConn(conn net.Conn) ServerCodecConn {
	return &serverCodecConn{Conn: conn}
}

func (conn *serverCodecConn) SetConn(c net.Conn) {
	conn.Conn = c
}

func (conn *serverCodecConn) GetConn() net.Conn {
	return conn.Conn
}

// SetServerCodec must ensure that both Conn and ServerCodecFunc are not nil
func (conn *serverCodecConn) SetServerCodec(fn ServerCodecFunc) {
	if fn != nil && conn.Conn != nil {
		conn.ServerCodec = fn(conn.Conn)
	}
}

func (conn *serverCodecConn) GetServerCodec() rpc.ServerCodec {
	return conn.ServerCodec
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (conn *serverCodecConn) Close() error {
	var err error
	if conn.ServerCodec != nil {
		err = conn.ServerCodec.Close()
	} else {
		err = conn.Conn.Close()
	}
	return err
}
