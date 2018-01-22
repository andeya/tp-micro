# Ant [![GitHub release](https://img.shields.io/github/release/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/releases) [![report card](https://goreportcard.com/badge/github.com/henrylee2cn/ant?style=flat-square)](http://goreportcard.com/report/henrylee2cn/ant) [![github issues](https://img.shields.io/github/issues/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/issues?q=is%3Aissue+is%3Aclosed) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/henrylee2cn/ant) [![view examples](https://img.shields.io/badge/learn%20by-examples-00BCD4.svg?style=flat-square)](https://github.com/henrylee2cn/ant/tree/master/samples) [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg?style=flat-square)](http://jq.qq.com/?_wv=1027&k=fzi4p1)


Ant is a simple and flexible microservice framework based on [Teleport](https://github.com/henrylee2cn/teleport).

[简体中文](https://github.com/henrylee2cn/ant/blob/master/README_ZH.md)


## Install


```
go version ≥ 1.7
```

```sh
go get -u github.com/henrylee2cn/ant
```

## Demo

- server

```go
package main

import (
	"github.com/henrylee2cn/ant"
	tp "github.com/henrylee2cn/teleport"
)

// Args args
type Args struct {
	A int
	B int `param:"<range:1:>"`
}

// P handler
type P struct {
	tp.PullCtx
}

// Divide divide API
func (p *P) Divide(args *Args) (int, *tp.Rerror) {
	return args.A / args.B, nil
}

func main() {
	srv := ant.NewServer(ant.SrvConfig{
		ListenAddress: ":9090",
	})
	srv.RoutePull(new(P))
	srv.Listen()
}
```

- client

```go
package main

import (
	"github.com/henrylee2cn/ant"
)

func main() {
	cli := ant.NewClient(
		ant.CliConfig{},
		ant.NewStaticLinker(":9090"),
	)
	defer	cli.Close()

	type Args struct {
		A int
		B int
	}

	var reply int
	rerr := cli.Pull("/p/divide", &Args{
		A: 10,
		B: 2,
	}, &reply).Rerror()
	if rerr != nil {
		ant.Fatalf("%v", rerr)
	}
	ant.Infof("10/2=%d", reply)
	rerr = cli.Pull("/p/divide", &Args{
		A: 10,
		B: 0,
	}, &reply).Rerror()
	if rerr == nil {
		ant.Fatalf("%v", rerr)
	}
	ant.Infof("test binding error: ok: %v", rerr)
```

[More](https://github.com/henrylee2cn/ant/tree/master/samples)

## License

Ant is under Apache v2 License. See the [LICENSE](https://github.com/henrylee2cn/ant/raw/master/LICENSE) file for the full license text
