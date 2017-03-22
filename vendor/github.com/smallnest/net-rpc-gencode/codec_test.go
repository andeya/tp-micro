package gencodec

import (
	"log"
	"net"
	"net/rpc"
	"testing"
	"time"
)

type Arith int

func (t *Arith) Mul(args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func listenTCP() (net.Listener, string) {
	l, e := net.Listen("tcp", "127.0.0.1:0") // any available address
	if e != nil {
		log.Fatalf("net.Listen tcp :0: %v", e)
	}
	return l, l.Addr().String()
}

func TestGencodec(t *testing.T) {
	rpc.Register(new(Arith))
	ln, address := listenTCP()
	defer ln.Close()

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				continue
			}
			go ServeConn(c)

		}
	}()

	client, err := DialTimeout("tcp", address, time.Minute)
	if err != nil {
		t.Error("dialing:", err)
	}

	defer client.Close()

	// Synchronous call
	args := &Args{7, 8}
	var reply Reply
	err = client.Call("Arith.Mul", args, &reply)
	if err != nil {
		t.Error("arith error:", err)
	} else {
		t.Logf("Arith: %d*%d=%d", args.A, args.B, reply.C)
	}
}
