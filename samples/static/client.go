package main

import (
	"github.com/henrylee2cn/ants"
)

func main() {
	cli := ants.NewClient(
		ants.CliConfig{Failover: 3},
		ants.NewStaticLinker(":9090"),
	)

	type Args struct {
		A int
		B int
	}

	var reply int
	rerr := cli.Pull("/p/divide", &Args{
		A: 10,
		B: 2,
	}, &reply).Rerror()
	if ants.IsConnRerror(rerr) {
		ants.Fatalf("has conn rerror: %v", rerr)
	}
	if rerr != nil {
		ants.Fatalf("%v", rerr)
	}
	ants.Infof("10/2=%d", reply)
	rerr = cli.Pull("/p/divide", &Args{
		A: 10,
		B: 0,
	}, &reply).Rerror()
	if ants.IsConnRerror(rerr) {
		ants.Fatalf("has conn rerror: %v", rerr)
	}
	if rerr == nil {
		ants.Fatalf("%v", rerr)
	}
	ants.Infof("test binding error: ok: %v", rerr)

	cli.Close()
	rerr = cli.Pull("/p/divide", &Args{
		A: 10,
		B: 5,
	}, &reply).Rerror()
	if rerr == nil {
		ants.Fatalf("test closing client: fail")
	}
	ants.Infof("test closing client: ok: %v", rerr)
}
