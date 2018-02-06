# Ant [![GitHub release](https://img.shields.io/github/release/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/releases) [![report card](https://goreportcard.com/badge/github.com/henrylee2cn/ant?style=flat-square)](http://goreportcard.com/report/henrylee2cn/ant) [![github issues](https://img.shields.io/github/issues/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/issues?q=is%3Aissue+is%3Aclosed) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/henrylee2cn/ant) [![view examples](https://img.shields.io/badge/learn%20by-examples-00BCD4.svg?style=flat-square)](https://github.com/henrylee2cn/ant/tree/master/samples) [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg?style=flat-square)](http://jq.qq.com/?_wv=1027&k=fzi4p1)


Ant 是一套简单、灵活的基于 [Teleport](https://github.com/henrylee2cn/teleport) 的微服务框架。


## 1. 安装

```
go version ≥ 1.7
```

```sh
go get -u github.com/henrylee2cn/ant
```

## 2. 特性

- 支持服务自动发现
- 支持自定义服务链接选择器
- 支持负载均衡
- 支持多路复用IO及其连接池
- 支持自定义协议
- 支持自定义Body的编解码类型
- 支持插件扩展
- 支持心跳机制
- 日志信息详尽，支持打印输入、输出消息的详细信息（状态码、消息头、消息体）
- 支持设置慢操作报警阈值
- 支持自定义日志
- 支持平滑关闭与更新
- 支持推送
- 支持的网络类型：`tcp`、`tcp4`、`tcp6`、`unix`、`unixpacket`等
- 客户端支持断线后自动重连

## 3. 项目结构

（推荐）

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

说明：

- 在自动生成的文件的文件名中增加 `.gen` 后缀进行标记

## 4. 示例

- 服务端

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

- 客户端

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

[更多](https://github.com/henrylee2cn/ant/tree/master/samples)


## 5. 命令行工具

- 快速创建ant项目
- 热编译模式运行ant项目

### 1. 安装

	```sh
	go install
	```

### 2. 用法

- 新建项目

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

- 热编译运行

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

[ant command](https://github.com/henrylee2cn/ant/tree/master/cmd/ant)

## 6. 平台架构

[Ants](https://github.com/xiaoenai/ants): 一套基于 [Ant](https://github.com/henrylee2cn/ant) 和 [Teleport](https://github.com/henrylee2cn/teleport) 的高可用的微服务平台解决方案。

## 7. 开源协议

Ant 项目采用商业应用友好的 [Apache2.0](https://github.com/henrylee2cn/ant/raw/master/LICENSE) 协议发布
