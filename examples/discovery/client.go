package main

import (
	"time"

	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket/example/pb"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
)

func main() {
	tp.SetSocketNoDelay(false)
	tp.SetShutdown(time.Second*20, nil, nil)

	cli := micro.NewClient(
		micro.CliConfig{
			DefaultBodyCodec:   "protobuf",
			DefaultDialTimeout: time.Second * 5,
			Failover:           3,
			CircuitBreaker: micro.CircuitBreakerConfig{
				Enable:          true,
				ErrorPercentage: 50,
			},
			HeartbeatSecond: 3,
		},
		discovery.NewLinker(etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		}),
	)
	defer cli.Close()
	var reply = new(pb.PbTest)
	rerr := cli.Pull(
		"/group/home/test",
		&pb.PbTest{A: 10, B: 2},
		reply,
	).Rerror()
	if rerr != nil {
		tp.Errorf("pull error: %v", rerr)
	} else {
		tp.Infof("pull reply: %v", reply)
	}

	// test heartbeat
	time.Sleep(10e9)
	cli.UsePullHeartbeat()
	time.Sleep(10e9)
}
