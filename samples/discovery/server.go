package main

import (
	"github.com/henrylee2cn/ants"
	"github.com/henrylee2cn/ants/discovery"
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
	srv := ants.NewServer(ants.SrvConfig{
		ListenAddress: ":9090",
		RouterRoot:    "/static",
	},
		discovery.ServicePlugin(":9090", []string{"http://127.0.0.1:2379"}),
	)
	srv.RoutePull(new(P))
	srv.Listen()
}
