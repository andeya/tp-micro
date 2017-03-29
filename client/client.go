package client

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"time"

	kcp "github.com/xtaci/kcp-go"

	codecGob "github.com/henrylee2cn/rpc2/codec/gob"
	"github.com/henrylee2cn/rpc2/common"
	"github.com/henrylee2cn/rpc2/log"
)

type (
	// Client rpc client.
	Client struct {
		ClientCodecFunc ClientCodecFunc
		// PluginContainer means as its name
		PluginContainer IClientPluginContainer
		// TLSConfig specifies the TLS configuration to use with tls.Config.
		TLSConfig *tls.Config
		// HTTPPath is only for HTTP network
		HTTPPath string
		// KCPBlock is only for KCP network
		KCPBlock kcp.BlockCrypt
		FailMode FailMode
		// The maximum number of attempts of the Call.
		MaxTry int
		//Timeout sets deadline for underlying net.Conns
		Timeout time.Duration
		//ReadTimeout sets readdeadline for underlying net.Conns
		ReadTimeout time.Duration
		//WriteTimeout sets writedeadline for underlying net.Conns
		WriteTimeout time.Duration
		selector     Selector
	}
)

//FailMode is a feature to decide client actions when clients fail to invoke services
type FailMode int

const (
	//Failover selects another server automaticaly
	Failover FailMode = iota
	//Failfast returns error immediately
	Failfast
	//Failtry use current client again
	Failtry
	//Broadcast sends requests to all servers and Success only when all servers return OK
	Broadcast
	//Forking sends requests to all servers and Success once one server returns OK
	Forking
)

// NewClient creates a new Client
func NewClient(client Client, selector Selector) *Client {
	client.selector = selector
	return client.init()
}

func (client *Client) init() *Client {
	if client.ClientCodecFunc == nil {
		client.ClientCodecFunc = codecGob.NewGobClientCodec
	}
	if client.PluginContainer == nil {
		client.PluginContainer = new(ClientPluginContainer)
	}
	if client.MaxTry <= 0 {
		client.MaxTry = 3
	}
	if client.selector == nil {
		log.Fatal("rpc: client do not have a 'Selector' field!")
	}
	client.selector.SetNewInvokerFunc(client.newInvoker)
	return client
}

var _ NewInvokerFunc = new(Client).newInvoker

// NewInvoker connects to an RPC server at the setted network address.
func (client *Client) newInvoker(network, address string, dialTimeout time.Duration) (Invoker, error) {
	var wrapper = &clientCodecWrapper{
		pluginContainer: client.PluginContainer,
		timeout:         client.Timeout,
		readTimeout:     client.ReadTimeout,
		writeTimeout:    client.WriteTimeout,
	}
	switch network {
	case "http":
		return client.newHTTPClient(network, address, dialTimeout, wrapper)
	case "kcp":
		return client.newKCPClient(address, wrapper)
	default:
		return client.newXXXClient(network, address, dialTimeout, wrapper)
	}
}

func (client *Client) newXXXClient(network, address string, dialTimeout time.Duration, wrapper *clientCodecWrapper) (Invoker, error) {
	var (
		err     error
		tlsConn *tls.Conn
		dialer  = &net.Dialer{Timeout: dialTimeout}
		conn    net.Conn
	)
	if client.TLSConfig != nil {
		tlsConn, err = tls.DialWithDialer(dialer, network, address, client.TLSConfig)
		conn = net.Conn(tlsConn)
	} else {
		conn, err = dialer.Dial(network, address)
	}
	if err == nil {
		wrapper.codecConn = NewClientCodecConn(conn)
		err = client.PluginContainer.doPostConnected(wrapper.codecConn)
		if err == nil {
			if wrapper.codecConn.GetClientCodec() == nil {
				wrapper.codecConn.SetClientCodec(client.ClientCodecFunc)
			}
			return NewInvokerWithCodec(wrapper), nil
		}
	}
	return nil, common.NewRPCError("dial error: " + err.Error())
}

func (client *Client) newHTTPClient(network, address string, dialTimeout time.Duration, wrapper *clientCodecWrapper) (Invoker, error) {
	if client.HTTPPath == "" {
		client.HTTPPath = rpc.DefaultRPCPath
	}
	var (
		err     error
		resp    *http.Response
		tlsConn *tls.Conn
		conn    net.Conn
		dialer  = &net.Dialer{Timeout: dialTimeout}
	)
	if client.TLSConfig != nil {
		tlsConn, err = tls.DialWithDialer(dialer, network, address, client.TLSConfig)
		conn = net.Conn(tlsConn)
	} else {
		conn, err = dialer.Dial(network, address)
	}
	if err == nil {
		wrapper.codecConn = NewClientCodecConn(conn)
		err = client.PluginContainer.doPostConnected(wrapper.codecConn)
		if err == nil {
			if wrapper.codecConn.GetClientCodec() == nil {
				wrapper.codecConn.SetClientCodec(client.ClientCodecFunc)
			}
			io.WriteString(wrapper.codecConn, "CONNECT "+client.HTTPPath+" HTTP/1.0\n\n")
			// Require successful HTTP response before switching to RPC protocol.
			resp, err = http.ReadResponse(bufio.NewReader(wrapper.codecConn), &http.Request{Method: "CONNECT"})
			if err == nil {
				if resp.Status == common.Connected {
					return NewInvokerWithCodec(wrapper), nil
				}
				err = common.NewRPCError("unexpected HTTP response: " + resp.Status)
			}
			wrapper.codecConn.Close()
		}
	}
	return nil, common.NewRPCError("dial error: " + (&net.OpError{
		Op:   "dial-http",
		Net:  network + " " + address,
		Addr: nil,
		Err:  err,
	}).Error())
}

func (client *Client) newKCPClient(address string, wrapper *clientCodecWrapper) (Invoker, error) {
	conn, err := kcp.DialWithOptions(address, client.KCPBlock, 10, 3)
	if err == nil {
		wrapper.codecConn = NewClientCodecConn(conn)
		err = client.PluginContainer.doPostConnected(wrapper.codecConn)
		if err == nil {
			if wrapper.codecConn.GetClientCodec() == nil {
				wrapper.codecConn.SetClientCodec(client.ClientCodecFunc)
			}
			return NewInvokerWithCodec(wrapper), nil
		}
	}
	return nil, common.NewRPCError("dial error: " + err.Error())
}

//Call invokes the named function, waits for it to complete, and returns its error status.
func (client *Client) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	if client.FailMode == Broadcast {
		return client.invokerBroadCast(serviceMethod, args, &reply)
	}
	if client.FailMode == Forking {
		return client.invokerForking(serviceMethod, args, &reply)
	}

	var invoker Invoker

	if client.FailMode == Failover {
		for tries := client.MaxTry; tries > 0; tries-- {
			invoker, err = client.selector.Select(serviceMethod, args)
			if err != nil || invoker == nil {
				continue
			}

			err = invoker.Call(serviceMethod, args, reply)
			if err == nil {
				return nil
			}

			log.Errorf("rpc: failed to call: %v", err)
			client.selector.HandleFailed(invoker)
		}

	} else if client.FailMode == Failtry {
		for tries := client.MaxTry; tries > 0; tries-- {
			if invoker == nil {
				if invoker, err = client.selector.Select(serviceMethod, args); err != nil {
					log.Errorf("rpc: failed to select a invoker: %v", err)
				}
			}

			if invoker != nil {
				err = invoker.Call(serviceMethod, args, reply)
				if err == nil {
					return nil
				}

				log.Errorf("rpc: failed to call: %v", err)
				client.selector.HandleFailed(invoker)
			}
		}
	}

	return
}

func (client *Client) invokerBroadCast(serviceMethod string, args interface{}, reply *interface{}) (err error) {
	invokers := client.selector.List()

	if len(invokers) == 0 {
		log.Infof("rpc: no any invoker is available")
		return nil
	}

	l := len(invokers)
	done := make(chan *Call, l)
	for _, invoker := range invokers {
		invoker.Go(serviceMethod, args, reply, done)
	}

	for l > 0 {
		call := <-done
		if call == nil || call.Error != nil {
			if call != nil {
				log.Warnf("rpc: failed to call: %v", call.Error)
			}
			return common.NewRPCError("some invokers return Error")
		}
		*reply = call.Reply
		l--
	}

	return nil
}

func (client *Client) invokerForking(serviceMethod string, args interface{}, reply *interface{}) (err error) {
	invokers := client.selector.List()

	if len(invokers) == 0 {
		log.Infof("rpc: no any invoker is available")
		return nil
	}

	l := len(invokers)
	done := make(chan *Call, l)
	for _, invoker := range invokers {
		invoker.Go(serviceMethod, args, reply, done)
	}

	for l > 0 {
		call := <-done
		if call != nil && call.Error == nil {
			*reply = call.Reply
			return nil
		}
		if call == nil {
			break
		}
		if call.Error != nil {
			log.Warnf("rpc: failed to call: %v", call.Error)
		}
		l--
	}

	return common.NewRPCError("all invokers return Error")
}

// Go invokes the function asynchronously. It returns the Call structure representing the invocation.
// The done channel will signal when the call is complete by returning the same Call object.
// If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
func (client *Client) Go(serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call {
	invoker, err := client.selector.Select()
	if err != nil {
		call := new(Call)
		call.ServiceMethod = serviceMethod
		call.Args = args
		call.Reply = reply
		call.Error = err
		if done == nil {
			done = make(chan *Call, 1) // buffered.
		} else {
			// If caller passes done != nil, it must arrange that
			// done has enough buffer for the number of simultaneous
			// RPCs that will be using that channel. If the channel
			// is totally unbuffered, it's best not to run at all.
			if cap(done) == 0 {
				log.Panic("rpc: done channel is unbuffered")
			}
		}
		call.Done = done
		call.done()
		return call
	}
	return invoker.Go(serviceMethod, args, reply, done)
}

// Close closes the connection
func (client *Client) Close() {
	for _, invoker := range client.selector.List() {
		client.selector.HandleFailed(invoker)
		invoker.Close()
	}
}
