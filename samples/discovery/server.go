package main

import (
	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/discovery"
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
	srv := ant.NewServer(ant.SrvConfig{
		ListenAddress: ":9090",
		RouterRoot:    "/static",
	},
		discovery.ServicePlugin(":9090", []string{"http://127.0.0.1:2379"}),
	)
	srv.RoutePull(new(P))
	srv.Listen()
}
