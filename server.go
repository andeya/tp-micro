package rpc2

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"net/rpc"
	"strings"
	"sync"
	"time"
)

type ClientIPs struct {
	ips       map[string]bool
	ipPrefixs []string
	sync.RWMutex
}

var clientIPs = &ClientIPs{
	ips: map[string]bool{},

	// lan is allowed
	ipPrefixs: []string{
		"[",
		"127.",
		"192.168.",
		"",
		"10.",
	},
}

func (this *ClientIPs) notAllow(addr string) bool {
	this.RLock()
	defer this.RUnlock()
	ip := strings.Split(addr, ":")[0]
	if this.ips[ip] {
		return false
	}
	for i, count := 0, len(this.ipPrefixs); i < count; i++ {
		if strings.HasPrefix(ip, this.ipPrefixs[i]) {
			return false
		}
	}
	return true
}

// Add the client ip that is allowed to connect,
// LAN ips are always allowed.
func AllowClients(ips ...string) {
	clientIPs.Lock()
	defer clientIPs.Unlock()
	for _, ip := range ips {
		clientIPs.ips[ip] = true
	}
}

func timeoutCoder(f func(interface{}) error, e interface{}, msg string) error {
	echan := make(chan error, 1)
	go func() { echan <- f(e) }()
	select {
	case e := <-echan:
		return e
	case <-time.After(time.Minute):
		return fmt.Errorf("Timeout %s", msg)
	}
}

type gobServerCodec struct {
	rwc    io.ReadWriteCloser
	dec    *gob.Decoder
	enc    *gob.Encoder
	encBuf *bufio.Writer
	closed bool
}

func (c *gobServerCodec) ReadRequestHeader(r *rpc.Request) error {
	return timeoutCoder(c.dec.Decode, r, "server read request header")
}

func (c *gobServerCodec) ReadRequestBody(body interface{}) error {
	return timeoutCoder(c.dec.Decode, body, "server read request body")
}

func (c *gobServerCodec) WriteResponse(r *rpc.Response, body interface{}) (err error) {
	if err = timeoutCoder(c.enc.Encode, r, "server write response"); err != nil {
		if c.encBuf.Flush() == nil {
			log.Println("rpc: gob error encoding response:", err)
			c.Close()
		}
		return
	}
	if err = timeoutCoder(c.enc.Encode, body, "server write response body"); err != nil {
		if c.encBuf.Flush() == nil {
			log.Println("rpc: gob error encoding body:", err)
			c.Close()
		}
		return
	}
	return c.encBuf.Flush()
}

func (c *gobServerCodec) Close() error {
	if c.closed {
		// Only call c.rwc.Close once; otherwise the semantics are undefined.
		return nil
	}
	c.closed = true
	return c.rwc.Close()
}

// Register publishes in the server the set of methods of the
// receiver value that satisfy the following conditions:
//	- exported method of exported type
//	- two arguments, both of exported type
//	- the second argument is a pointer
//	- one return value, of type error
// It returns an error if the receiver is not an exported type or has
// no suitable methods. It also logs the error using package log.
// The client accesses each method using a string of the form "Type.Method",
// where Type is the receiver's concrete type.
// make sure called before 'ListenRPC'.
func Register(rcvrs ...interface{}) {
	for _, rcvr := range rcvrs {
		rpc.Register(rcvr)
	}
}

// Open Service
func ListenRPC(addr string) {
	l, e := net.Listen("tcp", addr)
	if e != nil {
		log.Fatal("Error: listen %s error:", addr, e)
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Println("Error: accept rpc connection", err.Error())
				continue
			}

			// filter ip
			if clientIPs.notAllow(conn.RemoteAddr().String()) {
				log.Println("Client Not allow:", conn.RemoteAddr().String())
				continue
			}

			go func(conn net.Conn) {
				buf := bufio.NewWriter(conn)
				srv := &gobServerCodec{
					rwc:    conn,
					dec:    gob.NewDecoder(conn),
					enc:    gob.NewEncoder(buf),
					encBuf: buf,
				}
				err = rpc.ServeRequest(srv)
				if err != nil {
					log.Print("Error: server rpc request", err.Error())
				}
				srv.Close()
			}(conn)
		}
	}()
}
