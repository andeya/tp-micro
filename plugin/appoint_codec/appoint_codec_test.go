package appoint_codec

import (
	"testing"
	"time"

	"github.com/henrylee2cn/myrpc/client"
	"github.com/henrylee2cn/myrpc/client/selector"
	"github.com/henrylee2cn/myrpc/log"
	"github.com/henrylee2cn/myrpc/server"
)

type worker struct{}

func (*worker) Todo1(arg string, reply *string) error {
	log.Info("[server] Worker.Todo1: do job:", arg)
	*reply = "OK: " + arg
	return nil
}

func (*worker) Todo2(arg string, reply *string) error {
	log.Info("[server] Worker.Todo2: do job:", arg)
	*reply = "OK: " + arg
	return nil
}

func TestAppointCodecPlugin(t *testing.T) {
	// server
	srv := server.NewServer(server.Server{})
	srv.PluginContainer.Add(NewServerAppointCodecPlugin())
	srv.NamedRegister("work", new(worker))

	go srv.Serve("tcp", "0.0.0.0:8080")
	time.Sleep(2e9)

	// client
	c := client.NewClient(
		client.Client{
			FailMode: client.Failtry,
		},
		&selector.DirectSelector{
			Network: "tcp",
			Address: "127.0.0.1:8080",
		},
	)

	c.PluginContainer.Add(NewClientAppointCodecPlugin(CODEC_TYPE_JSON))

	var reply = new(string)
	e := c.Call("/work/todo1?key=value", "test_request1", reply)
	t.Log(*reply, e)
	e = c.Call("/work/todo2", "test_request2", reply)
	t.Log(*reply, e)
	call := <-c.Go("/work/todo2", "test_request2", reply, nil).Done
	t.Log(*reply, call.Error)
	c.Close()
	srv.Close()
}
