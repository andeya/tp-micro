package codec

import (
	"sync"
)

var (
	Network           = "tcp"
	ServerAddr        = "127.0.0.1:8080"
	ServiceGroup      = "Arith"
	ServiceName       = "1.0"
	ServiceMethodName = "Arith/1.0.Mul"
	Service           = new(Arith)
	once              sync.Once
)

type Args struct {
	A int `msg:"a"`
	B int `msg:"b"`
}

type Reply struct {
	C int `msg:"c"`
}

type Arith int

func (t *Arith) Mul(args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func (t *Arith) Error(args *Args, reply *Reply) error {
	panic("ERROR")
}
