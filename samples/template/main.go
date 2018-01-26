package main

import (
	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/samples/template/controller"
)

func main() {
	srv := ant.NewServer(ant.SrvConfig{
		ListenAddress:   ":9090",
		EnableHeartbeat: true,
	})
	controller.Route("/root", srv.Router())
	srv.Listen()
}
