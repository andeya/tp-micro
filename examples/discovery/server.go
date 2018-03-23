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

	cfg := micro.SrvConfig{
		DefaultBodyCodec: "protobuf",
		ListenAddress:    ":9090",
		EnableHeartbeat:  true,
	}
	srv := micro.NewServer(cfg, discovery.ServicePlugin(
		cfg.InnerIpPort(),
		etcd.EasyConfig{
			Endpoints: []string{"http://127.0.0.1:2379"},
		},
	))
	{
		group := srv.SubRoute("group")
		group.RoutePull(new(Home))
	}
	srv.Listen()
}

// Home controller
type Home struct {
	tp.PullCtx
}

// Test handler
func (h *Home) Test(args *pb.PbTest) (*pb.PbTest, *tp.Rerror) {
	return &pb.PbTest{
		A: args.A + args.B,
		B: args.A - args.B,
	}, nil
}
