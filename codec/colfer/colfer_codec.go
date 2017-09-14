package colfer

import "github.com/pascaldekloe/colfer/rpc"

// NewColferClientCodec returns a new Colfer implementation for the core library's RPC.
var NewColferClientCodec = rpc.NewClientCodec

// NewColferServerCodec returns a new Colfer implementation for the core library's RPC.
var NewColferServerCodec = rpc.NewServerCodec
