package gencode

import (
	"testing"
	"time"

	"github.com/henrylee2cn/myrpc"
	"github.com/henrylee2cn/myrpc/codec"
)

func TestGencodeCodec(t *testing.T) {
	// server
	server := myrpc.NewServer(60e9, 0, 0, NewGencodeServerCodec)
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

	err = myrpc.
		NewDialer(codec.Network, codec.ServerAddr, NewGencodeClientCodec).
		Remote(func(client myrpc.IClient) error {
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
