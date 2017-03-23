package plugin

import (
	"errors"
	"testing"
	"time"

	"github.com/henrylee2cn/rpc2"
	"github.com/henrylee2cn/rpc2/invokerselector"
)

type worker struct{}

func (*worker) Todo1(arg string, reply *string) error {
	rpc2.Log.Info("[server] Worker.Todo1: do job:", arg)
	*reply = "OK: " + arg
	return nil
}

func (*worker) Todo2(arg string, reply *string) error {
	rpc2.Log.Info("[server] Worker.Todo2: do job:", arg)
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
	server := rpc2.NewServer(rpc2.Server{
		RouterPrintable:   true,
		ServiceMethodFunc: rpc2.NewURLServiceMethod,
	})

	// authorization
	group, err := server.Group("/test", NewServerAuthorizationPlugin(checkAuthorization))
	if err != nil {
		panic(err)
	}

	err = group.RegisterName("/1.0.work", new(worker))
	if err != nil {
		panic(err)
	}

	go server.Serve("tcp", "0.0.0.0:8080")
	time.Sleep(2e9)

	// client
	client := rpc2.NewClient(
		rpc2.Client{
			FailMode: rpc2.Failtry,
		},
		&invokerselector.DirectInvokerSelector{
			Network: "tcp",
			Address: "127.0.0.1:8080",
		},
	)

	client.PluginContainer.Add(NewClientAuthorizationPlugin(rpc2.NewURLServiceMethod, __tag__, __token__))

	var reply = new(string)
	e := client.Call("/test/1.0.work/todo1?key=value", "test_request1", reply)
	t.Log(*reply, e)
	e = client.Call("/test/1.0.work/todo2", "test_request2", reply)
	t.Log(*reply, e)
	call := <-client.Go("/test/1.0.work/todo2", "test_request2", reply, nil).Done
	t.Log(*reply, call.Error)
	client.Close()
	server.Close()
}
