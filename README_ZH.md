# Ant [![GitHub release](https://img.shields.io/github/release/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/releases) [![report card](https://goreportcard.com/badge/github.com/henrylee2cn/ant?style=flat-square)](http://goreportcard.com/report/henrylee2cn/ant) [![github issues](https://img.shields.io/github/issues/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/issues?q=is%3Aopen+is%3Aissue) [![github closed issues](https://img.shields.io/github/issues-closed-raw/henrylee2cn/ant.svg?style=flat-square)](https://github.com/henrylee2cn/ant/issues?q=is%3Aissue+is%3Aclosed) [![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](http://godoc.org/github.com/henrylee2cn/ant) [![view examples](https://img.shields.io/badge/learn%20by-examples-00BCD4.svg?style=flat-square)](https://github.com/henrylee2cn/ant/tree/master/samples)
<!-- [![view Go网络编程群](https://img.shields.io/badge/官方QQ群-Go网络编程(42730308)-27a5ea.svg?style=flat-square)](http://jq.qq.com/?_wv=1027&k=fzi4p1) -->


Ant 是一套简单、灵活的基于 [Teleport](https://github.com/henrylee2cn/teleport) 的微服务框架。


## 1. 安装

```
go version ≥ 1.9
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

## 4. 用法

### 4.1 Peer端点（服务端或客户端）示例

```go
// Start a server
var peer1 = tp.NewPeer(tp.PeerConfig{
    ListenAddress: "0.0.0.0:9090", // for server role
})
peer1.Listen()

...

// Start a client
var peer2 = tp.NewPeer(tp.PeerConfig{})
var sess, err = peer2.Dial("127.0.0.1:8080")
```


### 4.2 Pull-Controller-Struct 接口模板

```go
type Aaa struct {
    tp.PullCtx
}
func (x *Aaa) XxZz(args *<T>) (<T>, *tp.Rerror) {
    ...
    return r, nil
}
```

- 注册到根路由：

```go
// register the pull route: /aaa/xx_zz
peer.RoutePull(new(Aaa))

// or register the pull route: /xx_zz
peer.RoutePullFunc((*Aaa).XxZz)
```

### 4.3 Pull-Handler-Function 接口模板

```go
func XxZz(ctx tp.PullCtx, args *<T>) (<T>, *tp.Rerror) {
    ...
    return r, nil
}
```

- 注册到根路由：

```go
// register the pull route: /xx_zz
peer.RoutePullFunc(XxZz)
```

### 4.4 Push-Controller-Struct 接口模板

```go
type Bbb struct {
    tp.PushCtx
}
func (b *Bbb) YyZz(args *<T>) *tp.Rerror {
    ...
    return nil
}
```

- 注册到根路由：

```go
// register the push route: /bbb/yy_zz
peer.RoutePush(new(Bbb))

// or register the push route: /yy_zz
peer.RoutePushFunc((*Bbb).YyZz)
```

### 4.5 Push-Handler-Function 接口模板

```go
// YyZz register the route: /yy_zz
func YyZz(ctx tp.PushCtx, args *<T>) *tp.Rerror {
    ...
    return nil
}
```

- 注册到根路由：

```go
// register the push route: /yy_zz
peer.RoutePushFunc(YyZz)
```

### 4.6 Unknown-Pull-Handler-Function 接口模板

```go
func XxxUnknownPull (ctx tp.UnknownPullCtx) (interface{}, *tp.Rerror) {
    ...
    return r, nil
}
```

- 注册到根路由：

```go
// register the unknown pull route: /*
peer.SetUnknownPull(XxxUnknownPull)
```

### 4.7 Unknown-Push-Handler-Function 接口模板

```go
func XxxUnknownPush(ctx tp.UnknownPushCtx) *tp.Rerror {
    ...
    return nil
}
```

- 注册到根路由：

```go
// register the unknown push route: /*
peer.SetUnknownPush(XxxUnknownPush)
```

### 4.8 插件示例

```go
// NewIgnoreCase Returns a ignoreCase plugin.
func NewIgnoreCase() *ignoreCase {
    return &ignoreCase{}
}

type ignoreCase struct{}

var (
    _ tp.PostReadPullHeaderPlugin = new(ignoreCase)
    _ tp.PostReadPushHeaderPlugin = new(ignoreCase)
)

func (i *ignoreCase) Name() string {
    return "ignoreCase"
}

func (i *ignoreCase) PostReadPullHeader(ctx tp.ReadCtx) *tp.Rerror {
    // Dynamic transformation path is lowercase
    ctx.Url().Path = strings.ToLower(ctx.Url().Path)
    return nil
}

func (i *ignoreCase) PostReadPushHeader(ctx tp.ReadCtx) *tp.Rerror {
    // Dynamic transformation path is lowercase
    ctx.Url().Path = strings.ToLower(ctx.Url().Path)
    return nil
}
```

### 4.9 注册以上操作和插件示例到路由

```go
// add router group
group := peer.SubRoute("test")
// register to test group
group.RoutePull(new(Aaa), NewIgnoreCase())
peer.RoutePullFunc(XxZz, NewIgnoreCase())
group.RoutePush(new(Bbb))
peer.RoutePushFunc(YyZz)
peer.SetUnknownPull(XxxUnknownPull)
peer.SetUnknownPush(XxxUnknownPush)
```

### 4.10 配置信息

```go
// SrvConfig server config
type SrvConfig struct {
	TlsCertFile       string        `yaml:"tls_cert_file"        ini:"tls_cert_file"        comment:"TLS certificate file path"`
	TlsKeyFile        string        `yaml:"tls_key_file"         ini:"tls_key_file"         comment:"TLS key file path"`
	DefaultSessionAge time.Duration `yaml:"default_session_age"  ini:"default_session_age"  comment:"Default session max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
	DefaultContextAge time.Duration `yaml:"default_context_age"  ini:"default_context_age"  comment:"Default PULL or PUSH context max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
	SlowCometDuration time.Duration `yaml:"slow_comet_duration"  ini:"slow_comet_duration"  comment:"Slow operation alarm threshold; ns,µs,ms,s ..."`
	DefaultBodyCodec  string        `yaml:"default_body_codec"   ini:"default_body_codec"   comment:"Default body codec type id"`
	PrintBody         bool          `yaml:"print_body"           ini:"print_body"           comment:"Is print body or not"`
	CountTime         bool          `yaml:"count_time"           ini:"count_time"           comment:"Is count cost time or not"`
	Network           string        `yaml:"network"              ini:"network"              comment:"Network; tcp, tcp4, tcp6, unix or unixpacket"`
	ListenAddress     string        `yaml:"listen_address"       ini:"listen_address"       comment:"Listen address; for server role"`
	EnableHeartbeat   bool          `yaml:"enable_heartbeat"     ini:"enable_heartbeat"     comment:"enable heartbeat"`
}

// CliConfig client config
type CliConfig struct {
	TlsCertFile         string        `yaml:"tls_cert_file"          ini:"tls_cert_file"          comment:"TLS certificate file path"`
	TlsKeyFile          string        `yaml:"tls_key_file"           ini:"tls_key_file"           comment:"TLS key file path"`
	DefaultSessionAge   time.Duration `yaml:"default_session_age"    ini:"default_session_age"    comment:"Default session max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
	DefaultContextAge   time.Duration `yaml:"default_context_age"    ini:"default_context_age"    comment:"Default PULL or PUSH context max age, if less than or equal to 0, no time limit; ns,µs,ms,s,m,h"`
	DefaultDialTimeout  time.Duration `yaml:"default_dial_timeout"   ini:"default_dial_timeout"   comment:"Default maximum duration for dialing; for client role; ns,µs,ms,s,m,h"`
	RedialTimes         int           `yaml:"redial_times"           ini:"redial_times"           comment:"The maximum times of attempts to redial, after the connection has been unexpectedly broken; for client role"`
	Failover            int           `yaml:"failover"               ini:"failover"               comment:"The maximum times of failover"`
	SlowCometDuration   time.Duration `yaml:"slow_comet_duration"    ini:"slow_comet_duration"    comment:"Slow operation alarm threshold; ns,µs,ms,s ..."`
	DefaultBodyCodec    string        `yaml:"default_body_codec"     ini:"default_body_codec"     comment:"Default body codec type id"`
	PrintBody           bool          `yaml:"print_body"             ini:"print_body"             comment:"Is print body or not"`
	CountTime           bool          `yaml:"count_time"             ini:"count_time"             comment:"Is count cost time or not"`
	Network             string        `yaml:"network"                ini:"network"                comment:"Network; tcp, tcp4, tcp6, unix or unixpacket"`
	HeartbeatSecond     int           `yaml:"heartbeat_second"       ini:"heartbeat_second"       comment:"When the heartbeat interval(second) is greater than 0, heartbeat is enabled; if it's smaller than 3, change to 3 default"`
	SessMaxQuota        int           `yaml:"sess_max_quota"         ini:"sess_max_quota"         comment:"The maximum number of sessions in the connection pool"`
	SessMaxIdleDuration time.Duration `yaml:"sess_max_idle_duration" ini:"sess_max_idle_duration" comment:"The maximum time period for the idle session in the connection pool; ns,µs,ms,s,m,h"`
}
```

### 4.11 通信优化

- SetPacketSizeLimit 设置包大小的上限，
  如果 maxSize<=0，上限默认为最大 uint32

    ```go
    func SetPacketSizeLimit(maxPacketSize uint32)
    ```

- SetSocketKeepAlive 是否允许操作系统的发送TCP的keepalive探测包

    ```go
    func SetSocketKeepAlive(keepalive bool)
    ```


- SetSocketKeepAlivePeriod 设置操作系统的TCP发送keepalive探测包的频度

    ```go
    func SetSocketKeepAlivePeriod(d time.Duration)
    ```

- SetSocketNoDelay 是否禁用Nagle算法，禁用后将不在合并较小数据包进行批量发送，默认为禁用

    ```go
    func SetSocketNoDelay(_noDelay bool)
    ```

- SetSocketReadBuffer 设置操作系统的TCP读缓存区的大小

    ```go
    func SetSocketReadBuffer(bytes int)
    ```

- SetSocketWriteBuffer 设置操作系统的TCP写缓存区的大小

    ```go
    func SetSocketWriteBuffer(bytes int)
    ```

[More Usage](https://github.com/henrylee2cn/teleport)

## 5. 示例

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

[更多示例](https://github.com/henrylee2cn/ant/tree/master/samples)


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
   ant new [command options] [arguments...]

OPTIONS:
   --script value, -s value    The script for code generation(relative/absolute)
   --app_path value, -p value  The path(relative/absolute) of the project
```

示例：`ant new -p ./myant -s ./test.ant`

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
   --app_path value, -p value    Specifies the path(relative/absolute) of the project
```

示例：`ant run -x .yaml -a myant` or `ant run -x .yaml myant`

[更多 ant command](https://github.com/henrylee2cn/ant/tree/master/cmd/ant)

## 6. 平台架构

[Ants](https://github.com/xiaoenai/ants): 一套基于 [Ant](https://github.com/henrylee2cn/ant) 和 [Teleport](https://github.com/henrylee2cn/teleport) 的高可用的微服务平台解决方案。

## 7. 开源协议

Ant 项目采用商业应用友好的 [Apache2.0](https://github.com/henrylee2cn/ant/raw/master/LICENSE) 协议发布
