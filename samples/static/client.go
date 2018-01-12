package main

import (
	"github.com/henrylee2cn/ants"
)

type Args struct {
	A int
	B int `param:"<range:1:>"`
}

func main() {
	cli := ants.NewClient(
		ants.CliConfig{},
		ants.NewStaticLinker(":9090"),
	)
	var reply int
	rerr := cli.Pull("/p/divide", &Args{
		A: 10,
		B: 2,
	}, &reply).Rerror()
	if rerr != nil {
		ants.Fatalf("%v", rerr)
	}
	ants.Infof("10/2=%d", reply)
	rerr = cli.Pull("/p/divide", &Args{
		A: 10,
		B: 0,
	}, &reply).Rerror()
	if rerr == nil {
		ants.Fatalf("%v", rerr)
	}
	ants.Errorf("10/0 error:%v", rerr)
}
