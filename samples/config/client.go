package main

import (
	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/cfgo"
)

func main() {
	cfg := ant.CliConfig{}

	// auto create and sync config/config.yaml
	cfgo.MustGet("config/config.yaml", true).MustReg("ant_cli", &cfg)

	cli := ant.NewClient(
		cfg,
		ant.NewStaticLinker(":9090"),
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
		ant.Fatalf("%v", rerr)
	}
	ant.Infof("%d/%d = %d", 10, 2, reply)
}
