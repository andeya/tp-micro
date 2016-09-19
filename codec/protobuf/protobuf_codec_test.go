package protobuf

import (
	"testing"
	"time"

	"github.com/henrylee2cn/rpc2"
	"github.com/henrylee2cn/rpc2/codec"
)

type ProtoArith int

func (t *ProtoArith) Mul(args *ProtoArgs, reply *ProtoReply) error {
	reply.C = args.A * args.B
	return nil
}

func (t *ProtoArith) Error(args *ProtoArgs, reply *ProtoReply) error {
	panic("ERROR")
}

func TestProtobufCodec(t *testing.T) {
	// server
	server := rpc2.NewServer(60e9, 0, 0, NewProtobufServerCodec)
	group := server.Group(codec.ServiceGroup)
	err := group.RegisterName(codec.ServiceName, new(ProtoArith))
	if err != nil {
		panic(err)
	}
	go server.ListenTCP(codec.ServerAddr)
	time.Sleep(2e9)

	// client
	client := rpc2.NewClient(codec.ServerAddr, NewProtobufClientCodec)

	args := &ProtoArgs{7, 8}
	var reply ProtoReply
	err = client.Call(codec.ServiceMethodName, args, &reply)
	if err != nil {
		t.Errorf("error for Arith: %d*%d, %v \n", args.A, args.B, err)
	} else {
		t.Logf("Arith: %d*%d=%d \n", args.A, args.B, reply.C)
	}

	client.Close()
}
