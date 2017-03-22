package plugin

import (
	"errors"
	"github.com/henrylee2cn/rpc2"
	"log"
	"testing"
	"time"
)

type worker struct{}

func (*worker) Todo1(task string, reply *string) error {
	log.Println("Worker.Todo1: do job", task)
	// time.Sleep(time.Second * 3)
	*reply = "OK"
	return nil
}

func (*worker) Todo2(task string, reply *string) error {
	log.Println("Worker.Todo2: do job", task)
	// time.Sleep(time.Second * 3)
	*reply = "OK"
	return nil
}

func TestAuthorization(t *testing.T) {
	const (
		__token__ = "1234567890"
		__tag__   = "basic"
	)

	var checkAuthorization = func(token string, tag string, serviceMethod string) error {
		if serviceMethod != "test/1.0.work.Todo1" {
			return nil
		}
		if __token__ == token && __tag__ == tag {
			return nil
		}
		return errors.New("Illegal request!")
	}

	// server
	server := rpc2.NewDefaultServer(true)

	// authorization
	group, err := server.Group("test", NewServerAuthorization(checkAuthorization))
	if err != nil {
		panic(err)
	}

	err = group.RegisterName("1.0.work", new(worker))
	if err != nil {
		panic(err)
	}

	go server.Serve("tcp", "0.0.0.0:8080")
	time.Sleep(2e9)

	// client
	factory := &rpc2.ClientFactory{
		Network:         "tcp",
		Address:         "127.0.0.1:8080",
		PluginContainer: new(rpc2.ClientPluginContainer),
	}
	factory.PluginContainer.Add(NewClientAuthorization(__token__, __tag__))
	client, _ := factory.NewClient()
	var reply = new(string)
	e := client.Call("test/1.0.work.Todo1", "test_request1", reply)
	t.Log(*reply, e)
	e = client.Call("test/1.0.work.Todo2", "test_request2", reply)
	t.Log(*reply, e)
	client.Close()
	server.Close()
}
