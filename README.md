# Ants [![report card](https://goreportcard.com/badge/github.com/henrylee2cn/ants?style=flat-square)](http://goreportcard.com/report/henrylee2cn/ants) [![github issues](https://img.shields.io/github/issues/henrylee2cn/ants.svg?style=flat-square)](https://github.com/henrylee2cn/ants/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/henrylee2cn/ants.svg?style=flat-square)](https://github.com/henrylee2cn/ants/issues?q=is%3Aissue+is%3Aclosed) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/henrylee2cn/ants) [![view examples](https://img.shields.io/badge/learn%20by-examples-00BCD4.svg?style=flat-square)](https://github.com/henrylee2cn/ants/tree/master/samples) [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg?style=flat-square)](http://jq.qq.com/?_wv=1027&k=fzi4p1)


Ants is a set of microservices-system based on [Teleport](https://github.com/henrylee2cn/teleport) framework and similar to lightweight service mesh.

[简体中文](https://github.com/henrylee2cn/ants/blob/master/README_ZH.md)


## Install


```
go version ≥ 1.9
```

```sh
go get -u github.com/henrylee2cn/ants/...
```

## Demo

- server

```go
package main

import (
	"github.com/henrylee2cn/ants"
)

type Args struct {
	A int
	B int `param:"<range:1:>"`
}

type P struct{ ants.PullCtx }

func (p *P) Divide(args *Args) (int, *ants.Rerror) {
	return args.A / args.B, nil
}

func main() {
	srv := ants.NewServer(ants.SrvConfig{ListenAddress: ":9090"})
	srv.PullRouter.Reg(new(P))
	srv.Listen()
}
```

- client

```go
package main

import (
	"github.com/henrylee2cn/ants"
)

func main() {
	cli := ants.NewClient(
		ants.CliConfig{},
		ants.NewStaticLinker(":9090"),
	)

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
		ants.Fatalf("%v", rerr)
	}
	ants.Infof("10/2=%d", reply)
	rerr = cli.Pull("/p/divide", &Args{
		A: 10,
		B: 0,
	}, &reply).Rerror()
	if rerr == nil {
		ants.Fatalf("%v", rerr)
	}
	ants.Infof("test binding error: ok: %v", rerr)

	cli.Close()
	rerr = cli.Pull("/p/divide", &Args{
		A: 10,
		B: 5,
	}, &reply).Rerror()
	if rerr == nil {
		ants.Fatalf("test closing client: fail")
	}
	ants.Infof("test closing client: ok: %v", rerr)
}
```

[More](https://github.com/henrylee2cn/ants/tree/master/samples)

## License

Ants is under Apache v2 License. See the [LICENSE](https://github.com/henrylee2cn/ants/raw/master/LICENSE) file for the full license text
