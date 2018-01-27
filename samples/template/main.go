package main

import (
	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/samples/template/api"
)

func main() {
	srv := ant.NewServer(ant.SrvConfig{
		ListenAddress:   ":9090",
		EnableHeartbeat: true,
	})
	api.Route("/root", srv.Router())
	srv.Listen()
}
