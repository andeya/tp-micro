package main

import (
	"time"

	tp "github.com/henrylee2cn/teleport"
	micro "github.com/henrylee2cn/tp-micro"
)

func main() {
	cli := micro.NewClient(
		micro.CliConfig{
			Failover:        3,
			HeartbeatSecond: 4,
		},
		micro.NewStaticLinker(":9090"),
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
	if tp.IsConnRerror(rerr) {
		tp.Fatalf("has conn rerror: %v", rerr)
	}
	if rerr != nil {
		tp.Fatalf("%v", rerr)
	}
	tp.Infof("10/2=%d", reply)
	rerr = cli.Pull("/static/p/divide", &Args{
		A: 10,
		B: 0,
	}, &reply).Rerror()
	if tp.IsConnRerror(rerr) {
		tp.Fatalf("has conn rerror: %v", rerr)
	}
	if rerr == nil {
		tp.Fatalf("%v", rerr)
	}
	tp.Infof("test binding error: ok: %v", rerr)

	time.Sleep(time.Second * 5)

	cli.Close()
	rerr = cli.Pull("/static/p/divide", &Args{
		A: 10,
		B: 5,
	}, &reply).Rerror()
	if rerr == nil {
		tp.Fatalf("test closing client: fail")
	}
	tp.Infof("test closing client: ok: %v", rerr)
}
