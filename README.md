# rpc2    [![GoDoc](https://godoc.org/github.com/tsuna/gohbase?status.png)](https://godoc.org/github.com/henrylee2cn/rpc2)

RPC2 is modified version that based on the standard package, 42% performance increase.

Added router group, middleware and header information in an form of 'URL Query'. 

![rpc2_server](https://github.com/henrylee2cn/rpc2/raw/master/doc/rpc2_server.png)

# usage

```
package main

import (
    "errors"
    "log"
    "net/rpc"
    "strconv"
    "time"

    "github.com/henrylee2cn/rpc2"
    "github.com/henrylee2cn/rpc2/plugin"
)

type Worker struct{}

func (w *Worker) Todo1(task string, reply *string) error {
    *reply = "Worker.Todo1 OK: " + task
    return nil
}

func (w *Worker) Todo2(task string, reply *string) error {
    *reply = "Worker.Todo2 OK: " + task
    return nil
}

type serverPlugin struct{}

func (t *serverPlugin) Name() string {
    return "server_plugin"
}

func (t *serverPlugin) PostReadRequestHeader(ctx *rpc2.Context) error {
    ctx.Log.Infof("serverPlugin.PostReadRequestHeader -> ctx: %v", ctx)
    return nil
}

type clientPlugin struct{}

func (t *clientPlugin) Name() string {
    return "client_plugin"
}

func (t *clientPlugin) PostReadResponseHeader(resp *rpc.Response) error {
    log.Printf("clientPlugin.PostReadResponseHeader -> resp: %v", resp)
    return nil
}

const (
    __token__ = "1234567890"
    __tag__   = "basic"
)

func checkAuthorization(token string, tag string, serviceMethod string) error {
    if serviceMethod != "test/1.0.work.Todo1" {
        return nil
    }
    if __token__ == token && __tag__ == tag {
        return nil
    }
    return errors.New("Illegal request!")
}

// rpc2
func main() {
    // server
    server := rpc2.NewDefaultServer(true)

    // ip filter
    ipwl := plugin.NewIPWhitelist()
    ipwl.Allow("127.0.0.1")
    server.PluginContainer.Add(ipwl)

    // authorization

    group, err := server.Group("test", plugin.NewServerAuthorization(checkAuthorization), new(serverPlugin))
    if err != nil {
        panic(err)
    }
    err = group.RegisterName("1.0.work", new(Worker))
    if err != nil {
        panic(err)
    }
    go server.Serve("tcp", "0.0.0.0:8080")
    time.Sleep(2e9)

    // client
    dialer := &rpc2.Dialer{
        Network:         "tcp",
        Address:         "127.0.0.1:8080",
        PluginContainer: new(rpc2.ClientPluginContainer),
    }
    dialer.PluginContainer.Add(plugin.NewClientAuthorization(__token__, __tag__), new(clientPlugin))
   
    client, _ := dialer.Dial()
    
    var reply = new(string)
    e := client.Call("test/1.0.work.Todo1", "test_request1", reply)
    log.Println(*reply, e)
    e = client.Call("test/1.0.work.Todo2", "test_request2", reply)
    log.Println(*reply, e)
    client.Close()
    server.Close()
}

```
