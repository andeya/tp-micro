package main

import (
	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/henrylee2cn/tp-micro"
)

// Arg arg
type Arg struct {
	A int
	B int `param:"<range:1:>"`
}

// Divide divide API
func Divide(ctx tp.PullCtx, arg *Arg) (int, *tp.Rerror) {
	return arg.A / arg.B, nil
}

func main() {
	cfg := micro.SrvConfig{}
	// auto create and sync config/config.yaml
	cfgo.MustGet("config/config.yaml", true).MustReg("micro_srv", &cfg)

	srv := micro.NewServer(cfg)

	group := srv.SubRoute("/static")
	group.RoutePullFunc(Divide)
	srv.ListenAndServe()
}
