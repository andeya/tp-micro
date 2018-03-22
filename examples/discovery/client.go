package main

import (
	"time"

	tp "github.com/henrylee2cn/teleport"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
)

func main() {
	cli := micro.NewClient(
		micro.CliConfig{
			Failover: 3,
			// HeartbeatSecond: 30,
		},
		discovery.NewLinker(etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		}),
	)

	var arg = &struct {
		A int
		B int
	}{
		A: 10,
		B: 2,
	}

	var reply int

	rerr := cli.Pull("/p/divide", arg, &reply).Rerror()
	if rerr != nil {
		tp.Fatalf("%v", rerr)
	}
	tp.Infof("10/2=%d", reply)

	tp.Debugf("waiting for 10s...")
	time.Sleep(time.Second * 10)

	arg.B = 5
	rerr = cli.Pull("/p/divide", arg, &reply).Rerror()
	if rerr != nil {
		tp.Fatalf("%v", rerr)
	}
	tp.Infof("10/5=%d", reply)

	tp.Debugf("waiting for 10s...")
	time.Sleep(time.Second * 10)
}
