package main

import (
	"github.com/henrylee2cn/cfgo"
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/henrylee2cn/tp-micro"
)

func main() {
	cfg := micro.CliConfig{}

	// auto create and sync config/config.yaml
	cfgo.MustGet("config/config.yaml", true).MustReg("micro_cli", &cfg)

	cli := micro.NewClient(
		cfg,
		micro.NewStaticLinker(":9090"),
	)

	type Args struct {
		A int
		B int
	}

	var reply int
	rerr := cli.Pull("/static/divide", &Args{
		A: 10,
		B: 2,
	}, &reply).Rerror()
	if rerr != nil {
		tp.Fatalf("%v", rerr)
	}
	tp.Infof("%d/%d = %d", 10, 2, reply)
}
