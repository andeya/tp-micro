package main

import (
	"github.com/henrylee2cn/ant"
	"{{PROJ_PATH}}/api"
)

func main() {
	srv := ant.NewServer(ant.SrvConfig{
		ListenAddress:   ":9090",
		EnableHeartbeat: true,
	})
	api.Route("/root", srv.Router())
	srv.Listen()
}
