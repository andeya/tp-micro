package gencodec_test

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"time"
	"github.com/smallnest/net-rpc-gencode"
)

type MyArith int

func (t *MyArith) Mul(args *gencodec.Args, reply *gencodec.Reply) error {
	reply.C = args.A * args.B
	return nil
}

func Example() {
	rpc.Register(new(MyArith))
	ln, e := net.Listen("tcp", "127.0.0.1:0") // any available address
	if e != nil {
		log.Fatalf("net.Listen tcp :0: %v", e)
	}
	address := ln.Addr().String()
	defer ln.Close()

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				continue
			}
			go gencodec.ServeConn(c)

		}
	}()

	client, err := gencodec.DialTimeout("tcp", address, time.Minute)
	if err != nil {
		fmt.Println("dialing:", err)
	}

	defer client.Close()

	// Synchronous call
	args := &gencodec.Args{7, 8}
	var reply gencodec.Reply
	err = client.Call("MyArith.Mul", args, &reply)
	if err != nil {
		fmt.Println("arith error:", err)
	} else {
		fmt.Printf("Arith: %d*%d=%d\n", args.A, args.B, reply.C)
	}
}
