package main

import (
	"time"

	"github.com/henrylee2cn/ant"
)

func main() {
	cli := ant.NewClient(
		ant.CliConfig{
			Failover:  3,
			Heartbeat: time.Second,
		},
		ant.NewStaticLinker(":9090"),
	)

	type Args struct {
		A int
		B int
	}

	var reply int
	rerr := cli.Pull("/static/p/divide", &Args{
		A: 10,
		B: 2,
	}, &reply).Rerror()
	if ant.IsConnRerror(rerr) {
		ant.Fatalf("has conn rerror: %v", rerr)
	}
	if rerr != nil {
		ant.Fatalf("%v", rerr)
	}
	ant.Infof("10/2=%d", reply)
	rerr = cli.Pull("/static/p/divide", &Args{
		A: 10,
		B: 0,
	}, &reply).Rerror()
	if ant.IsConnRerror(rerr) {
		ant.Fatalf("has conn rerror: %v", rerr)
	}
	if rerr == nil {
		ant.Fatalf("%v", rerr)
	}
	ant.Infof("test binding error: ok: %v", rerr)

	time.Sleep(time.Second * 5)

	cli.Close()
	rerr = cli.Pull("/static/p/divide", &Args{
		A: 10,
		B: 5,
	}, &reply).Rerror()
	if rerr == nil {
		ant.Fatalf("test closing client: fail")
	}
	ant.Infof("test closing client: ok: %v", rerr)
}
