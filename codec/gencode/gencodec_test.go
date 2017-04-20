package gencode

import (
	"testing"
	"time"

	"github.com/henrylee2cn/rpc2"
	"github.com/henrylee2cn/rpc2/codec"
)

func TestGencodeCodec(t *testing.T) {
	// server
	server := rpc2.NewServer(60e9, 0, 0, NewGencodeServerCodec)
	group, _ := server.Group(codec.ServiceGroup)
	err := group.NamedRegister(codec.ServiceName, new(GencodeArith))
	if err != nil {
		t.Fatal(err)
	}
	go server.ListenAndServe(codec.Network, codec.ServerAddr)
	time.Sleep(2e9)

	// client
	var args = &GencodeArgs{7, 8}
	var reply GencodeReply

	err = rpc2.
		NewDialer(codec.Network, codec.ServerAddr, NewGencodeClientCodec).
		Remote(func(client rpc2.IClient) error {
			return client.Call(codec.ServiceMethodName, args, &reply)
		})

	if err != nil {
		t.Errorf("error for Arith: %d*%d, %v \n", args.A, args.B, err)
	} else {
		t.Logf("Arith: %d*%d=%d \n", args.A, args.B, reply.C)
	}
}

type GencodeArith int

func (t *GencodeArith) Mul(args *GencodeArgs, reply *GencodeReply) error {
	reply.C = args.A * args.B
	return nil
}

func (t *GencodeArith) Error(args *GencodeArgs, reply *GencodeReply) error {
	panic("ERROR")
}
