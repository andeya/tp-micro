package main

import (
	"github.com/henrylee2cn/ants"
	tp "github.com/henrylee2cn/teleport"
)

type Args struct {
	A int
	B int `param:"<range:1:>"`
}

type P struct{ tp.PullCtx }

func (p *P) Divide(args *Args) (int, *tp.Rerror) {
	return args.A / args.B, nil
}

func main() {
	srv := ants.NewServer(ants.SrvConfig{
		ListenAddress: ":9090",
		RouterRoot:    "/static",
	})
	srv.RoutePull(new(P))
	srv.Listen()
}
