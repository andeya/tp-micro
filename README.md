# rpc2    [![GoDoc](https://godoc.org/github.com/tsuna/gohbase?status.png)](https://godoc.org/github.com/henrylee2cn/rpc2)

RPC2 is upgraded version RPC that based on the standard package.

It's added router Group, Plugin, Selector and customized ServiceMethod. 

![rpc2_server](https://github.com/henrylee2cn/rpc2/raw/master/doc/rpc2_server.png)

# Demo

```
package main

import (
    "errors"
    "net/rpc"
    "strconv"
    "time"

    "github.com/henrylee2cn/rpc2/client"
    "github.com/henrylee2cn/rpc2/client/selector"
    "github.com/henrylee2cn/rpc2/log"
    "github.com/henrylee2cn/rpc2/plugin/auth"
    "github.com/henrylee2cn/rpc2/plugin/ip_whitelist"
    "github.com/henrylee2cn/rpc2/server"
)

type Worker struct{}

func (*Worker) Todo1(arg string, reply *string) error {
    log.Info("[server] Worker.Todo1: do job:", arg)
    *reply = "OK: " + arg
    return nil
}

func (*Worker) Todo2(arg string, reply *string) error {
    log.Info("[server] Worker.Todo2: do job:", arg)
    *reply = "OK: " + arg
    return nil
}

type serverRedirectPlugin struct{}

func (t *serverRedirectPlugin) Name() string {
    return "server_plugin"
}

func (t *serverRedirectPlugin) PostReadRequestHeader(ctx *server.Context) error {
    if ctx.Path() == "/test/1.0.work/todo1" {
        ctx.SetPath("/test/1.0.work/todo2")
        log.Info("Redirect to todo2")
    }
    return nil
}

type clientPlugin struct{}

func (t *clientPlugin) Name() string {
    return "client_plugin"
}

func (t *clientPlugin) PostReadResponseHeader(resp *rpc.Response) error {
    log.Infof("clientPlugin.PostReadResponseHeader -> resp: %v", resp)
    return nil
}

const (
    __token__ = "1234567890"
    __tag__   = "basic"
)

func checkAuthorization(serviceMethod, tag, token string) error {
    if serviceMethod != "/test/1.0.work/todo1" {
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
    srv := server.NewServer(server.Server{
        RouterPrintable: true,
    })

    // ip filter
    ipwl := ip_whitelist.NewIPWhitelistPlugin()
    ipwl.Allow("127.0.0.1")
    srv.PluginContainer.Add(ipwl)

    // redirect
    srv.PluginContainer.Add(new(serverRedirectPlugin))

    // authorization
    group, err := srv.Group(
        "test",
        auth.NewServerAuthorizationPlugin(checkAuthorization),
    )
    if err != nil {
        panic(err)
    }

    err = group.RegisterName("1.0.work", new(Worker))
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

    c.PluginContainer.Add(
        auth.NewClientAuthorizationPlugin(new(server.URLFormat), __tag__, __token__),
        new(clientPlugin),
    )

    N := 1
    bad := 0
    good := 0
    mapChan := make(chan int, N)
    t1 := time.Now()
    for i := 0; i < N; i++ {
        go func(i int) {
            var reply = new(string)
            e := c.Call("/test/1.0.work/todo1?key=henrylee2cn", strconv.Itoa(i), reply)
            log.Info(i, *reply, e)
            if e != nil {
                mapChan <- 0
            } else {
                mapChan <- 1
            }
        }(i)
    }
    for i := 0; i < N; i++ {
        if r := <-mapChan; r == 0 {
            bad++
        } else {
            good++
        }
    }
    c.Close()
    srv.Close()
    log.Info("cost time:", time.Now().Sub(t1))
    log.Info("success rate:", float64(good)/float64(good+bad)*100, "%")
}
```
