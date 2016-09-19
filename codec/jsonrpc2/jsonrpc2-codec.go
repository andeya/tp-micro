package jsonrpc2

import (
	"io"
	"net/rpc"
)

// NewJSONRPC2ServerCodec creates a RPC-JSON 2.0 ServerCodec
func NewJSONRPC2ServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return NewServerCodec(conn, nil)
}

// NewJSONRPC2ClientCodec creates a RPC-JSON 2.0 ClientCodec
func NewJSONRPC2ClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return NewClientCodec(conn)
}
