# net-rpc-gencode
[![GoDoc](https://godoc.org/github.com/smallnest/net-rpc-gencode?status.png)](http://godoc.org/github.com/smallnest/net-rpc-gencode) [![Drone Build Status](https://drone.io/github.com/smallnest/net-rpc-gencode/status.png)](https://drone.io/github.com/smallnest/net-rpc-gencode/latest) [![Go Report Card](http://goreportcard.com/badge/smallnest/net-rpc-gencode)](http://goreportcard.com/report/smallnest/net-rpc-gencode)

This library provides the same functions as net/rpc/jsonrpc but for communicating with [gencode](https://github.com/andyleap/gencode) instead. The library is modeled directly after the Go standard library so it should be easy to use and obvious.

See the [GoDoc](https://godoc.org/github.com/smallnest/net-rpc-gencode) for API documentation.

> according to my test: [Golang Serializer Benchmark Comparison](https://github.com/smallnest/gosercomp), gencode is very faster than other serializers.


## Example

### gencode Schema 

**arith_test.schema**:

```go 
struct Args {
	A int32
	B int32
}

struct Reply {
	C int32
}
```

run `gencode go -schema arith_test.schema -package gencodec` to generate data

### RPC Server


```go
package gencodec

import (
	"log"
	"net"
	"net/rpc"
	"github.com/smallnest/net-rpc-gencode"
)

type Arith int

func (t *Arith) Mul(args *Args, reply *Reply) error {
	reply.C = args.A * args.B
	return nil
}

func main() {
	rpc.Register(new(Arith))
	ln, e := net.Listen("tcp", "127.0.0.1:5432") // any available address
	if e != nil {
		log.Fatalf("net.Listen tcp :0: %v", e)
	}
	defer ln.Close()

	for {
		c, err := ln.Accept()
		if err != nil {
			continue
		}
		go gencodec.ServeConn(c)

	}
}
```

### Client

```go
package gencodec

import (
	"log"
	"net"
	"time"
	"github.com/smallnest/net-rpc-gencode"
)

func main() {
	client, err := gencodec.DialTimeout("tcp", "127.0.0.1:5432", time.Minute)
	if err != nil {
		fmt.Println("dialing:", err)
	}

	defer client.Close()

	// Synchronous call
	args := &gencodec.Args{7, 8}
	var reply Reply
	err = client.Call("Arith.Mul", args, &reply)
	if err != nil {
		fmt.Println("arith error:", err)
	} else {
		fmt.Printf("Arith: %d*%d=%d\n", args.A, args.B, reply.C)
	}
}
```