package rpc2

import (
	"bufio"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"time"

	kcp "github.com/xtaci/kcp-go"

	codecGob "github.com/henrylee2cn/rpc2/codec/gob"
)

type (
	// ClientFactory dials rpc server and returns Client.
	ClientFactory struct {
		Network         string
		Address         string
		ClientCodecFunc ClientCodecFunc
		// PluginContainer is by name
		PluginContainer IClientPluginContainer
		// TLSConfig specifies the TLS configuration to use with tls.Config.
		TLSConfig *tls.Config
		// HTTPPath is only for HTTP network
		HTTPPath string
		// KCPBlock is only for KCP network
		KCPBlock kcp.BlockCrypt
		// DialTimeout specify the number of redials, and the timeout for each redial.
		// Default use exponentialBackoff.
		DialTimeouts []time.Duration
		//Timeout sets deadline for underlying net.Conns
		Timeout time.Duration
		//ReadTimeout sets readdeadline for underlying net.Conns
		ReadTimeout time.Duration
		//WriteTimeout sets writedeadline for underlying net.Conns
		WriteTimeout time.Duration
		Log          Logger
	}
	// ClientCallFunc callback function for remote calls.
	ClientCallFunc func(Client) error
	// ClientCodecFunc is used to create a rpc.ClientCodec from net.Conn.
	ClientCodecFunc func(io.ReadWriteCloser) rpc.ClientCodec
)

// NewClientFactory creates a new ClientFactory
func NewClientFactory(factory ClientFactory) *ClientFactory {
	return factory.init()
}

func (factory *ClientFactory) init() *ClientFactory {
	if factory.ClientCodecFunc == nil {
		factory.ClientCodecFunc = codecGob.NewGobClientCodec
	}
	if len(factory.DialTimeouts) == 0 {
		// exponentialBackoff for tial timeout.
		factory.DialTimeouts = func() []time.Duration {
			var ds []time.Duration
			for i := uint(0); i < 4; i++ {
				ds = append(ds, time.Duration(0.75e9*2<<i))
			}
			return ds
		}()
	}
	if factory.Log == nil {
		factory.Log = newDefaultLogger()
	}
	if factory.PluginContainer == nil {
		factory.PluginContainer = new(ClientPluginContainer)
	}
	return factory
}

// NewClient connects to an RPC server at the specified network address.
func NewClient(network, address string) (Client, error) {
	return (&ClientFactory{
		Network: network,
		Address: address,
	}).NewClient()
}

// RemoteCall is shortcut method of rpc calling remotely.
func RemoteCall(network, address string, callfunc ClientCallFunc) error {
	client, err := NewClient(network, address)
	if err != nil {
		return err
	}
	defer client.Close()
	return callfunc(client)
}

// RemoteCall create a client and make remote calls.
func (factory *ClientFactory) RemoteCall(callfunc ClientCallFunc) error {
	client, err := factory.NewClient()
	if err != nil {
		return err
	}
	defer client.Close()
	return callfunc(client)
}

// NewClient connects to an RPC server at the setted network address.
func (factory *ClientFactory) NewClient() (Client, error) {
	var wrapper = &clientCodecWrapper{
		pluginContainer: factory.PluginContainer,
		timeout:         factory.Timeout,
		readTimeout:     factory.ReadTimeout,
		writeTimeout:    factory.WriteTimeout,
	}
	switch factory.Network {
	case "http":
		return factory.newHTTPClient(wrapper)
	case "kcp":
		return factory.newKCPClient(wrapper)
	default:
		return factory.newXXXClient(wrapper)
	}
}

func (factory *ClientFactory) newXXXClient(wrapper *clientCodecWrapper) (Client, error) {
	var (
		err     error
		tlsConn *tls.Conn
		dialer  = new(net.Dialer)
	)
	for _, d := range factory.DialTimeouts {
		dialer.Timeout = d
		if factory.TLSConfig != nil {
			tlsConn, err = tls.DialWithDialer(dialer, factory.Network, factory.Address, factory.TLSConfig)
			wrapper.conn = net.Conn(tlsConn)
		} else {
			wrapper.conn, err = dialer.Dial(factory.Network, factory.Address)
		}
		if err == nil {
			wrapper.conn, err = factory.PluginContainer.doPostConnected(wrapper.conn)
			if err == nil {
				wrapper.codec = factory.ClientCodecFunc(wrapper.conn)
				return NewClientWithCodec(wrapper), nil
			}
		}
	}
	return nil, NewRPCError("dial error: ", err.Error())
}

func (factory *ClientFactory) newHTTPClient(wrapper *clientCodecWrapper) (Client, error) {
	if factory.HTTPPath == "" {
		factory.HTTPPath = rpc.DefaultRPCPath
	}
	var (
		err     error
		resp    *http.Response
		tlsConn *tls.Conn
		dialer  = new(net.Dialer)
	)
	for _, d := range factory.DialTimeouts {
		dialer.Timeout = d
		if factory.TLSConfig != nil {
			tlsConn, err = tls.DialWithDialer(dialer, factory.Network, factory.Address, factory.TLSConfig)
			wrapper.conn = net.Conn(tlsConn)
		} else {
			wrapper.conn, err = dialer.Dial(factory.Network, factory.Address)
		}
		if err == nil {
			wrapper.conn, err = factory.PluginContainer.doPostConnected(wrapper.conn)
			if err == nil {
				wrapper.codec = factory.ClientCodecFunc(wrapper.conn)
				io.WriteString(wrapper.conn, "CONNECT "+factory.HTTPPath+" HTTP/1.0\n\n")
				// Require successful HTTP response
				// before switching to RPC protocol.
				resp, err = http.ReadResponse(bufio.NewReader(wrapper.conn), &http.Request{Method: "CONNECT"})
				if err == nil {
					if resp.Status == connected {
						return NewClientWithCodec(wrapper), nil
					}
					err = errors.New("unexpected HTTP response: " + resp.Status)
				}
				wrapper.conn.Close()
			}
		}
	}
	return nil, NewRPCError("dial error: " + (&net.OpError{
		Op:   "dial-http",
		Net:  factory.Network + " " + factory.Address,
		Addr: nil,
		Err:  err,
	}).Error())
}

func (factory *ClientFactory) newKCPClient(wrapper *clientCodecWrapper) (Client, error) {
	var err error
	for range factory.DialTimeouts {
		wrapper.conn, err = kcp.DialWithOptions(factory.Address, factory.KCPBlock, 10, 3)
		if err == nil {
			wrapper.conn, err = factory.PluginContainer.doPostConnected(wrapper.conn)
			if err == nil {
				wrapper.codec = factory.ClientCodecFunc(wrapper.conn)
				return NewClientWithCodec(wrapper), nil
			}
		}
	}
	return nil, NewRPCError("dial error: ", err.Error())
}

type clientCodecWrapper struct {
	pluginContainer IClientPluginContainer
	codec           rpc.ClientCodec
	conn            net.Conn
	timeout         time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
}

func (w *clientCodecWrapper) WriteRequest(r *rpc.Request, body interface{}) error {
	if w.timeout > 0 {
		w.conn.SetDeadline(time.Now().Add(w.timeout))
	}
	if w.writeTimeout > 0 {
		w.conn.SetWriteDeadline(time.Now().Add(w.writeTimeout))
	}

	//pre
	err := w.pluginContainer.doPreWriteRequest(r, body)
	if err != nil {
		return err
	}

	err = w.codec.WriteRequest(r, body)
	if err != nil {
		return NewRPCError("WriteRequest: ", err.Error())
	}

	//post
	err = w.pluginContainer.doPostWriteRequest(r, body)
	return err
}

func (w *clientCodecWrapper) ReadResponseHeader(r *rpc.Response) error {
	if w.timeout > 0 {
		w.conn.SetDeadline(time.Now().Add(w.timeout))
	}
	if w.readTimeout > 0 {
		w.conn.SetReadDeadline(time.Now().Add(w.readTimeout))
	}

	//pre
	err := w.pluginContainer.doPreReadResponseHeader(r)
	if err != nil {
		return err
	}

	err = w.codec.ReadResponseHeader(r)
	if err != nil {
		return NewRPCError("ReadResponseHeader: ", err.Error())
	}

	//post
	err = w.pluginContainer.doPostReadResponseHeader(r)
	return err
}

func (w *clientCodecWrapper) ReadResponseBody(body interface{}) error {
	//pre
	err := w.pluginContainer.doPreReadResponseBody(body)
	if err != nil {
		return err
	}

	err = w.codec.ReadResponseBody(body)
	if err != nil {
		return NewRPCError("ReadResponseBody: ", err.Error())
	}

	//post
	err = w.pluginContainer.doPostReadResponseBody(body)
	return err
}

func (w *clientCodecWrapper) Close() error {
	return w.codec.Close()
}
