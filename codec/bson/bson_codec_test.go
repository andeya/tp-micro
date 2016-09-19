package bson

import (
	"testing"
	"time"

	"github.com/henrylee2cn/rpc2"
	"github.com/henrylee2cn/rpc2/codec"
)

func TestBsonCodec(t *testing.T) {
	// server
	server := rpc2.NewServer(60e9, 0, 0, NewBsonServerCodec)
	group := server.Group(codec.ServiceGroup)
	err := group.RegisterName(codec.ServiceName, codec.Service)
	if err != nil {
		panic(err)
	}
	go server.ListenTCP(codec.ServerAddr)
	time.Sleep(2e9)

	// client
	client := rpc2.NewClient(codec.ServerAddr, NewBsonClientCodec)

	args := &codec.Args{7, 8}
	var reply codec.Reply
	err = client.Call(codec.ServiceMethodName, args, &reply)
	if err != nil {
		t.Errorf("error for Arith: %d*%d, %v \n", args.A, args.B, err)
	} else {
		t.Logf("Arith: %d*%d=%d \n", args.A, args.B, reply.C)
	}

	client.Close()
}
