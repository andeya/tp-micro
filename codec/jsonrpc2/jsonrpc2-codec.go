package jsonmyrpc

import (
	"io"
	"net/rpc"
)

// NewJSONMyrpcServerCodec creates a RPC-JSON 2.0 ServerCodec
func NewJSONMyrpcServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return NewServerCodec(conn, nil)
}

// NewJSONMyrpcClientCodec creates a RPC-JSON 2.0 ClientCodec
func NewJSONMyrpcClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return NewClientCodec(conn)
}
