# Ant [![GitHub release](https://img.shields.io/github/release/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/releases) [![report card](https://goreportcard.com/badge/github.com/henrylee2cn/ant?style=flat-square)](http://goreportcard.com/report/henrylee2cn/ant) [![github issues](https://img.shields.io/github/issues/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/issues?q=is%3Aissue+is%3Aclosed) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/henrylee2cn/ant) [![view examples](https://img.shields.io/badge/learn%20by-examples-00BCD4.svg?style=flat-square)](https://github.com/henrylee2cn/ant/tree/master/samples) [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg?style=flat-square)](http://jq.qq.com/?_wv=1027&k=fzi4p1)


Ant is a simple and flexible microservice framework based on [Teleport](https://github.com/henrylee2cn/teleport).

[简体中文](https://github.com/henrylee2cn/ant/blob/master/README_ZH.md)


## 1. Install


```
go version ≥ 1.7
```

```sh
go get -u github.com/henrylee2cn/ant
```

## 2. Feature

- Support auto service-discovery
- Supports custom service linker
- Support load balancing
- Support NIO and connection pool
- Support custom protocol
- Support custom body codec
- Support plug-in expansion
- Support heartbeat mechanism
- Detailed log information, support print input and output details
- Support for setting slow operation alarm thresholds
- Support for custom log
- Support smooth shutdown and update
- Support push handler
- Support network list: `tcp`, `tcp4`, `tcp6`, `unix`, `unixpacket` and so on
- Client support automatically redials after disconnection

## 3. Project Structure

(recommend)

```
├── README.md
├── main.go
├── api
│   ├── handlers.gen.go
│   ├── handlers.go
│   ├── router.gen.go
│   └── router.go
├── logic
│   └── xxx.go
├── sdk
│   ├── rpc.gen.go
│   ├── rpc.gen_test.go
│   ├── rpc.go
│   └── rpc_test.go
└── types
    ├── types.gen.go
    └── types.go
```

Desc:

- add `.gen` suffix to the file name of the automatically generated file

## 4. Demo

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
}
```

[More](https://github.com/henrylee2cn/ant/tree/master/samples)


## 5. Command

Command ant is deployment tools of ant microservice frameware.

- Quickly create a ant project
- Run ant project with hot compilation

### 1. Install

```sh
go install
```

### 2. Usage

- new project

```
NAME:
   ant new - Create a new ant project

USAGE:
   ant new [options] [arguments...]
 or
   ant new [options except -app_path] [arguments...] {app_path}

OPTIONS:
   --app_path value, -a value  Specifies the path(relative/absolute) of the project
```

example: `ant new -a myant` or `ant new myant`

- run project

```
NAME:
   ant run - Compile and run gracefully (monitor changes) an any existing go project

USAGE:
   ant run [options] [arguments...]
 or
   ant run [options except -app_path] [arguments...] {app_path}

OPTIONS:
   --watch_exts value, -x value  Specified to increase the listening file suffix (default: ".go", ".ini", ".yaml", ".toml", ".xml")
   --app_path value, -a value    Specifies the path(relative/absolute) of the project
```

example: `ant run -x .yaml -a myant` or `ant run -x .yaml myant`

[ant command](https://github.com/henrylee2cn/ant/tree/master/cmd/ant)

## 6. Platform

[Ants](https://github.com/xiaoenai/ants): a highly available microservice platform based on [Ant](https://github.com/henrylee2cn/ant) and [Teleport](https://github.com/henrylee2cn/teleport).

## 7. License

Ant is under Apache v2 License. See the [LICENSE](https://github.com/henrylee2cn/ant/raw/master/LICENSE) file for the full license text
