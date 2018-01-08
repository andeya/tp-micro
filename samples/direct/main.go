package main

import (
	"time"

	"github.com/henrylee2cn/ants"
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
	srv := ants.NewServer(ants.SrvConfig{ListenAddress: ":9090"})
	srv.PullRouter.Reg(new(P))
	go srv.Listen()
	time.Sleep(time.Second)

	cli := ants.NewClient(ants.CliConfig{})
	cli.SetLinker(ants.NewDirectLinker(":9090"))
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
