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

	type Arg struct {
		A int
		B int
	}

	var result int
	rerr := cli.Pull("/static/divide", &Arg{
		A: 10,
		B: 2,
	}, &result).Rerror()
	if rerr != nil {
		tp.Fatalf("%v", rerr)
	}
	tp.Infof("%d/%d = %d", 10, 2, result)
}
