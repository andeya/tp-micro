package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/henrylee2cn/rpc2/client"
	"github.com/henrylee2cn/rpc2/client/selector"
	"github.com/henrylee2cn/rpc2/log"
	"github.com/henrylee2cn/rpc2/server"
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

func TestAuthorizationPlugin(t *testing.T) {
	const (
		__token__ = "1234567890"
		__tag__   = "basic"
	)

	var checkAuthorization = func(serviceMethod, tag, token string) error {
		if serviceMethod != "/test/1.0.work/todo1" {
			return nil
		}
		if __token__ == token && __tag__ == tag {
			return nil
		}
		return errors.New("Illegal request!")
	}

	// server
	srv := server.NewServer(server.Server{
		RouterPrintable: true,
	})

	// authorization
	group, err := srv.Group("test", NewServerAuthorizationPlugin(checkAuthorization))
	if err != nil {
		panic(err)
	}

	err = group.RegisterName("1.0.work", new(worker))
	if err != nil {
		panic(err)
	}

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

	c.PluginContainer.Add(NewClientAuthorizationPlugin(new(server.URLFormat), __tag__, __token__))

	var reply = new(string)
	e := c.Call("/test/1.0.work/todo1?key=value", "test_request1", reply)
	t.Log(*reply, e)
	e = c.Call("/test/1.0.work/todo2", "test_request2", reply)
	t.Log(*reply, e)
	call := <-c.Go("/test/1.0.work/todo2", "test_request2", reply, nil).Done
	t.Log(*reply, call.Error)
	c.Close()
	srv.Close()
}
