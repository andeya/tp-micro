package bson

import (
	"testing"
	"time"

	"github.com/henrylee2cn/myrpc"
	"github.com/henrylee2cn/myrpc/codec"
)

func TestBsonCodec(t *testing.T) {
	// server
	server := myrpc.NewServer(60e9, 0, 0, NewBsonServerCodec)
	group, _ := server.Group(codec.ServiceGroup)
	err := group.NamedRegister(codec.ServiceName, codec.Service)
	if err != nil {
		t.Fatal(err)
	}
	go server.ListenAndServe(codec.Network, codec.ServerAddr)
	time.Sleep(2e9)

	// client
	var args = &codec.Args{7, 8}
	var reply codec.Reply

	err = myrpc.
		NewDialer(codec.Network, codec.ServerAddr, NewBsonClientCodec).
		Remote(func(client myrpc.IClient) error {
			return client.Call(codec.ServiceMethodName, args, &reply)
		})

	if err != nil {
		t.Errorf("error for Arith: %d*%d, %v \n", args.A, args.B, err)
	} else {
		t.Logf("Arith: %d*%d=%d \n", args.A, args.B, reply.C)
	}
}
