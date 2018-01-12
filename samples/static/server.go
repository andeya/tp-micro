package main

import (
	"github.com/henrylee2cn/ants"
)

type Args struct {
	A int
	B int `param:"<range:1:>"`
}

type P struct{ ants.PullCtx }

func (p *P) Divide(args *Args) (int, *ants.Rerror) {
	return args.A / args.B, nil
}

func main() {
	srv := ants.NewServer(ants.SrvConfig{ListenAddress: ":9090"})
	srv.PullRouter.Reg(new(P))
	srv.Listen()
}
