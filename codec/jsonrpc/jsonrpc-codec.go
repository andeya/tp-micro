package jsonrpc

import (
	"net/rpc/jsonrpc"
)

// NewJSONRPCServerCodec creates a RPC-JSON 2.0 ServerCodec
var NewJSONRPCServerCodec = jsonrpc.NewServerCodec

// NewJSONRPCClientCodec creates a RPC-JSON 2.0 ClientCodec
var NewJSONRPCClientCodec = jsonrpc.NewClientCodec
