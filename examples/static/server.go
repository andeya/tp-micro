package main

import (
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/henrylee2cn/tp-micro"
)

// Args args
type Args struct {
	A int
	B int `param:"<range:1:100>"`
}

// P handler
type P struct {
	tp.PullCtx
}

// Divide divide API
func (p *P) Divide(args *Args) (int, *tp.Rerror) {
	return args.A / args.B, nil
}

func main() {
	srv := micro.NewServer(micro.SrvConfig{
		ListenAddress:   ":9090",
		EnableHeartbeat: true,
	})
	group := srv.SubRoute("/static")
	group.RoutePull(new(P))
	srv.Listen()
}
