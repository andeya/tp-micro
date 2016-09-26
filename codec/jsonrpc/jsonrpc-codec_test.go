package jsonrpc

import (
	"testing"
	"time"

	"github.com/henrylee2cn/rpc2"
	"github.com/henrylee2cn/rpc2/codec"
)

func TestJSONRPCCodec(t *testing.T) {
	// server
	server := rpc2.NewServer(60e9, 0, 0, NewJSONRPCServerCodec)
	group, _ := server.Group(codec.ServiceGroup)
	err := group.RegisterName(codec.ServiceName, codec.Service)
	if err != nil {
		t.Fatal(err)
	}
	go server.ListenAndServe(codec.Network, codec.ServerAddr)
	time.Sleep(2e9)

	// client
	var args = &codec.Args{7, 8}
	var reply codec.Reply

	err = rpc2.
		NewDialer(codec.Network, codec.ServerAddr, NewJSONRPCClientCodec).
		Remote(func(client rpc2.IClient) error {
			return client.Call(codec.ServiceMethodName, args, &reply)
		})

	if err != nil {
		t.Errorf("error for Arith: %d*%d, %v \n", args.A, args.B, err)
	} else {
		t.Logf("Arith: %d*%d=%d \n", args.A, args.B, reply.C)
	}
}
