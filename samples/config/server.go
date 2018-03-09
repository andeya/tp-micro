package main

import (
	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
)

// Args args
type Args struct {
	A int
	B int `param:"<range:1:>"`
}

// Divide divide API
func Divide(ctx tp.PullCtx, args *Args) (int, *tp.Rerror) {
	return args.A / args.B, nil
}

func main() {
	cfg := ant.SrvConfig{
		ListenAddress: ":9090",
	}

	// auto create and sync config/config.yaml
	cfgo.MustGet("config/config.yaml", true).MustReg("ant_srv", &cfg)

	srv := ant.NewServer(cfg)

	group := srv.SubRoute("/static")
	group.RoutePullFunc(Divide)
	srv.Listen()
}
