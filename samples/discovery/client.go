package main

import (
	"time"

	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/discovery"
)

func main() {
	cli := ant.NewClient(
		ant.CliConfig{Failover: 3},
		discovery.NewLinker([]string{"http://127.0.0.1:2379"}),
	)

	var arg = &struct {
		A int
		B int
	}{
		A: 10,
		B: 2,
	}

	var reply int

	rerr := cli.Pull("/static/p/divide", arg, &reply).Rerror()
	if rerr != nil {
		ant.Fatalf("%v", rerr)
	}
	ant.Infof("10/2=%d", reply)

	time.Sleep(time.Second * 15)

	arg.B = 5
	rerr = cli.Pull("/static/p/divide", arg, &reply).Rerror()
	if rerr != nil {
		ant.Fatalf("%v", rerr)
	}
	ant.Infof("10/5=%d", reply)

	time.Sleep(time.Second * 15)
}
