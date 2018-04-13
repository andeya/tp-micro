package main

import (
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery"

	"github.com/henrylee2cn/tp-micro/examples/sample/api"
)

func main() {
	srv := micro.NewServer(
		cfg.Srv,
		discovery.ServicePlugin(cfg.Srv.InnerIpPort(), cfg.Etcd),
	)
	api.Route("/sample", srv.Router())
	srv.ListenAndServe()
}
