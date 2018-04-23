package main

import (
	"github.com/henrylee2cn/cfgo"
	"github.com/henrylee2cn/goutil"
	tp "github.com/henrylee2cn/teleport"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/discovery/etcd"
	"github.com/xiaoenai/ants/model"
	"github.com/xiaoenai/ants/model/redis"

	mod "github.com/henrylee2cn/tp-micro/examples/sample/logic/model"
)

type config struct {
	Srv      micro.SrvConfig `yaml:"srv"`
	Etcd     etcd.EasyConfig `yaml:"etcd"`
	DB       model.Config    `yaml:"db"`
	Redis    redis.Config    `yaml:"redis"`
	LogLevel string          `yaml:"log_level"`
}

func (c *config) Reload(bind cfgo.BindFunc) error {
	err := bind()
	if err != nil {
		return err
	}
	if len(c.LogLevel) == 0 {
		c.LogLevel = "TRACE"
	}
	tp.SetLoggerLevel(c.LogLevel)
	err = mod.Init(c.DB, c.Redis)
	if err != nil {
		tp.Errorf("%v", err)
	}
	return nil
}

var cfg = &config{
	Srv: micro.SrvConfig{
		ListenAddress:   ":9090",
		EnableHeartbeat: true,
		PrintDetail:     true,
		CountTime:       true,
	},
	Etcd: etcd.EasyConfig{
		Endpoints: []string{"http://127.0.0.1:2379"},
	},
	DB: model.Config{
		Port: 3306,
	},
	Redis:    *redis.NewConfig(),
	LogLevel: "TRACE",
}

func init() {
	goutil.WritePidFile()
	cfgo.MustReg("sample", cfg)
}
