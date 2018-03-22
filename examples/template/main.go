package main

import (
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/examples/template/api"
)

func main() {
	srv := micro.NewServer(micro.SrvConfig{
		ListenAddress:   ":9090",
		EnableHeartbeat: true,
	})
	api.Route("/template", srv.Router())
	srv.Listen()
}
