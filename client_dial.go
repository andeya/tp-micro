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
	// Dialer dial rpc server.
	Dialer struct {
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

// exponentialBackoff for tial timeout.
var exponentialBackoff = func() []time.Duration {
	var ds []time.Duration
	for i := uint(0); i < 4; i++ {
		ds = append(ds, time.Duration(0.75e9*2<<i))
	}
	return ds
}()

func (dialer *Dialer) Init() *Dialer {
	if dialer.ClientCodecFunc == nil {
		dialer.ClientCodecFunc = codecGob.NewGobClientCodec
	}
	if len(dialer.DialTimeouts) == 0 {
		dialer.DialTimeouts = exponentialBackoff
	}
	if dialer.Log == nil {
		dialer.Log = newDefaultLogger()
	}
	if dialer.PluginContainer == nil {
		dialer.PluginContainer = new(ClientPluginContainer)
	}
	return dialer
}

// Remote is shortcut method of rpc calling remotely.
func Remote(callfunc ClientCallFunc) error {
	return new(Dialer).Init().Remote(callfunc)
}

// Remote is shortcut method of rpc calling remotely.
func (dialer *Dialer) Remote(callfunc ClientCallFunc) error {
	client, err := dialer.Dial()
	if err != nil {
		return err
	}
	defer client.Close()
	return callfunc(client)
}

// Dial connects to an RPC server at the specified network address.
func Dial(network, address string) (Client, error) {
	return new(Dialer).Init().Dial()
}

// Dial connects to an RPC server at the setted network address.
func (dialer *Dialer) Dial() (Client, error) {
	dialer.Init()
	var wrapper = &clientCodecWrapper{
		pluginContainer: dialer.PluginContainer,
		timeout:         dialer.Timeout,
		readTimeout:     dialer.ReadTimeout,
		writeTimeout:    dialer.WriteTimeout,
	}
	switch dialer.Network {
	case "http":
		return dialer.dialHTTP(wrapper)
	case "kcp":
		return dialer.dialKCP(wrapper)
	default:
		return dialer.dialXXX(wrapper)
	}
}

func (dialer *Dialer) dialXXX(wrapper *clientCodecWrapper) (Client, error) {
	var (
		err       error
		tlsConn   *tls.Conn
		netDialer = new(net.Dialer)
	)
	for _, d := range dialer.DialTimeouts {
		netDialer.Timeout = d
		if dialer.TLSConfig != nil {
			tlsConn, err = tls.DialWithDialer(netDialer, dialer.Network, dialer.Address, dialer.TLSConfig)
			wrapper.conn = net.Conn(tlsConn)
		} else {
			wrapper.conn, err = netDialer.Dial(dialer.Network, dialer.Address)
		}
		if err == nil {
			wrapper.conn, err = dialer.PluginContainer.doPostConnected(wrapper.conn)
			if err == nil {
				wrapper.codec = dialer.ClientCodecFunc(wrapper.conn)
				return NewClientWithCodec(wrapper), nil
			}
		}
	}
	return nil, NewRPCError("dial error: ", err.Error())
}

func (dialer *Dialer) dialHTTP(wrapper *clientCodecWrapper) (Client, error) {
	if dialer.HTTPPath == "" {
		dialer.HTTPPath = rpc.DefaultRPCPath
	}
	var (
		err       error
		resp      *http.Response
		tlsConn   *tls.Conn
		netDialer = new(net.Dialer)
	)
	for _, d := range dialer.DialTimeouts {
		netDialer.Timeout = d
		if dialer.TLSConfig != nil {
			tlsConn, err = tls.DialWithDialer(netDialer, dialer.Network, dialer.Address, dialer.TLSConfig)
			wrapper.conn = net.Conn(tlsConn)
		} else {
			wrapper.conn, err = netDialer.Dial(dialer.Network, dialer.Address)
		}
		if err == nil {
			wrapper.conn, err = dialer.PluginContainer.doPostConnected(wrapper.conn)
			if err == nil {
				wrapper.codec = dialer.ClientCodecFunc(wrapper.conn)
				io.WriteString(wrapper.conn, "CONNECT "+dialer.HTTPPath+" HTTP/1.0\n\n")
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
		Net:  dialer.Network + " " + dialer.Address,
		Addr: nil,
		Err:  err,
	}).Error())
}

func (dialer *Dialer) dialKCP(wrapper *clientCodecWrapper) (Client, error) {
	var err error
	for range dialer.DialTimeouts {
		wrapper.conn, err = kcp.DialWithOptions(dialer.Address, dialer.KCPBlock, 10, 3)
		if err == nil {
			wrapper.conn, err = dialer.PluginContainer.doPostConnected(wrapper.conn)
			if err == nil {
				wrapper.codec = dialer.ClientCodecFunc(wrapper.conn)
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
		return NewRPCError("PreReadResponseHeader: ", err.Error())
	}

	err = w.codec.ReadResponseHeader(r)
	if err != nil {
		return NewRPCError("ReadResponseHeader: ", err.Error())
	}

	//post
	err = w.pluginContainer.doPostReadResponseHeader(r)
	if err != nil {
		return NewRPCError("PostReadResponseHeader: ", err.Error())
	}
	return nil
}

func (w *clientCodecWrapper) ReadResponseBody(body interface{}) error {
	//pre
	err := w.pluginContainer.doPreReadResponseBody(body)
	if err != nil {
		return NewRPCError("PreReadResponseBody: ", err.Error())
	}

	err = w.codec.ReadResponseBody(body)
	if err != nil {
		return NewRPCError("ReadResponseBody: ", err.Error())
	}

	//post
	err = w.pluginContainer.doPostReadResponseBody(body)
	if err != nil {
		return NewRPCError("PostReadResponseBody: ", err.Error())
	}
	return nil
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
		return NewRPCError("PreWriteRequest: ", err.Error())
	}

	err = w.codec.WriteRequest(r, body)
	if err != nil {
		return NewRPCError("WriteRequest: ", err.Error())
	}

	//post
	err = w.pluginContainer.doPostWriteRequest(r, body)
	if err != nil {
		return NewRPCError("PostWriteRequest: ", err.Error())
	}
	return nil
}

func (w *clientCodecWrapper) Close() error {
	return w.codec.Close()
}
