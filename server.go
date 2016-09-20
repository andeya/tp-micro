package rpc2

import (
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"path"
	"reflect"
	"time"
)

type (
	Server struct {
		plugins         []Plugin
		groupMap        map[string]*Group
		timeout         time.Duration
		readTimeout     time.Duration
		writeTimeout    time.Duration
		serverCodecFunc ServerCodecFunc
		ipWhitelist     *IPWhitelist
		*rpc.Server
	}

	Group struct {
		prefix  string
		plugins []Plugin
		*Server
	}

	Plugin interface {
		PostReadRequestHeader(*rpc.Request) error
		PostReadRequestBody(interface{}) error
	}

	serverCodecWrapper struct {
		rpc.ServerCodec
		conn         io.ReadWriteCloser
		groupPlugins []Plugin
		*Server
	}

	// ServerCodecFunc is used to create a rpc.ServerCodec from io.ReadWriteCloser.
	ServerCodecFunc func(io.ReadWriteCloser) rpc.ServerCodec
)

func NewServer(
	timeout,
	readTimeout,
	writeTimeout time.Duration,
	serverCodecFunc ServerCodecFunc,
) *Server {
	if serverCodecFunc == nil {
		serverCodecFunc = NewGobServerCodec
	}

	return &Server{
		Server:          rpc.NewServer(),
		plugins:         []Plugin{},
		timeout:         timeout,
		readTimeout:     readTimeout,
		writeTimeout:    writeTimeout,
		serverCodecFunc: serverCodecFunc,
		ipWhitelist:     NewIPWhitelist(),
		groupMap:        map[string]*Group{},
	}
}

func NewDefaultServer() *Server {
	return NewServer(time.Minute, 0, 0, nil)
}

func (server *Server) Group(typePrefix string, plugins ...Plugin) *Group {
	if !nameRegexp.MatchString(typePrefix) {
		log.Fatal("Group's prefix ('" + typePrefix + "') must conform to the regular expression '/^[a-zA-Z0-9_\\.]+$/'.")
		return nil
	}
	return (&Group{
		plugins: []Plugin{},
		Server:  server,
	}).Group(typePrefix, plugins...)
}

func (server *Server) Plugin(plugins ...Plugin) {
	server.plugins = append(server.plugins, plugins...)
}

func (group *Group) Group(typePrefix string, plugins ...Plugin) *Group {
	if !nameRegexp.MatchString(typePrefix) {
		log.Fatal("Group's prefix ('" + typePrefix + "') must conform to the regular expression '/^[a-zA-Z0-9_\\.]+$/'.")
		return nil
	}
	g := &Group{
		prefix:  path.Join(group.prefix, typePrefix),
		plugins: plugins,
		Server:  group.Server,
	}
	g.Server.groupMap[g.prefix] = g
	return g
}

// RegisterName is like server.Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, rcvr interface{}) error {
	if !nameRegexp.MatchString(name) {
		log.Fatal("RegisterName ('" + name + "') must conform to the regular expression '/^[a-zA-Z0-9_\\.]+$/'.")
		return nil
	}
	return server.Server.RegisterName(name, rcvr)
}

func (group *Group) RegisterName(name string, rcvr interface{}) error {
	if !nameRegexp.MatchString(name) {
		log.Fatal("RegisterName ('" + name + "') must conform to the regular expression '/^[a-zA-Z0-9_\\.]+$/'.")
		return nil
	}
	return group.Server.Server.RegisterName(path.Join(group.prefix, name), rcvr)
}

func (group *Group) Register(rcvr interface{}) error {
	name := reflect.Indirect(reflect.ValueOf(rcvr)).Type().Name()
	return group.Server.Server.RegisterName(path.Join(group.prefix, name), rcvr)
}

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
// ServeConn uses the setted wire format on the connection.
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	server.Server.ServeCodec(server.wrapServerCodec(conn))
}

// ServeCodec is ServeConn's alias.
func (server *Server) ServeCodec(conn io.ReadWriteCloser) {
	server.ServeConn(conn)
}

// ServeRequest is like ServeConn but synchronously serves a single request.
// It does not close the codec upon completion.
func (server *Server) ServeRequest(conn io.ReadWriteCloser) error {
	return server.Server.ServeRequest(server.wrapServerCodec(conn))
}

func (server *Server) wrapServerCodec(conn io.ReadWriteCloser) *serverCodecWrapper {
	return &serverCodecWrapper{
		ServerCodec:  server.serverCodecFunc(conn),
		conn:         conn,
		Server:       server,
		groupPlugins: make([]Plugin, 0, 10),
	}
}

func (w *serverCodecWrapper) ReadRequestHeader(r *rpc.Request) error {
	var (
		conn net.Conn
		ok   bool
	)
	if w.Server.timeout > 0 {
		if conn, ok = w.conn.(net.Conn); ok {
			conn.SetDeadline(time.Now().Add(w.Server.timeout))
		}
	}
	if ok && w.Server.readTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(w.Server.readTimeout))
	}

	// decode
	err := w.ServerCodec.ReadRequestHeader(r)
	if err != nil {
		return err
	}

	for _, plugin := range w.Server.plugins {
		err = plugin.PostReadRequestHeader(r)
		if err != nil {
			return err
		}
	}

	serviceMethod := ParseServiceMethod(r.ServiceMethod)

	for _, groupPrefix := range serviceMethod.Groups() {
		group, ok := w.groupMap[groupPrefix]
		if !ok {
			return errors.New("rpc: can't find group " + groupPrefix)
		}
		for _, plugin := range group.plugins {
			err = plugin.PostReadRequestHeader(r)
			if err != nil {
				return err
			}
			w.groupPlugins = append(w.plugins, plugin)
		}
	}

	r.ServiceMethod = serviceMethod.Path

	return err
}

func (w *serverCodecWrapper) ReadRequestBody(body interface{}) error {
	err := w.ServerCodec.ReadRequestBody(body)
	if err != nil {
		return err
	}
	for _, plugin := range w.Server.plugins {
		err = plugin.PostReadRequestBody(body)
		if err != nil {
			return err
		}
	}
	for _, plugin := range w.groupPlugins {
		err = plugin.PostReadRequestBody(body)
		if err != nil {
			return err
		}
	}
	w.groupPlugins = nil
	return err
}

// WriteResponse must be safe for concurrent use by multiple goroutines.
func (w *serverCodecWrapper) WriteResponse(resp *rpc.Response, body interface{}) error {
	var (
		conn net.Conn
		ok   bool
	)
	if w.Server.timeout > 0 {
		if conn, ok = w.conn.(net.Conn); ok {
			conn.SetDeadline(time.Now().Add(w.Server.timeout))
		}
	}
	if ok && w.Server.writeTimeout > 0 {
		conn.SetWriteDeadline(time.Now().Add(w.Server.writeTimeout))
	}

	err := w.ServerCodec.WriteResponse(resp, body)

	return err
}

// Open Service
// @timeout, optional, setting server response timeout.
func (server *Server) ListenTCP(addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("Error: listen %s error:", addr, err)
	}
	server.Accept(lis)
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection. Accept blocks until the listener
// returns a non-nil error. The caller typically invokes Accept in a
// go statement.
func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Print("rpc.Serve: accept:", err.Error())
			return
		}

		// filter ip
		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		if !server.ipWhitelist.IsAllowed(ip) {
			log.Print("Not allowed client IP: ", ip)
			conn.Close()
			continue
		}

		go server.ServeConn(conn)
	}
}

func (server *Server) IP() *IPWhitelist {
	return server.ipWhitelist
}

// Can connect to RPC service using HTTP CONNECT to rpcPath.
var connected = "200 Connected to Go RPC"

// ServeHTTP implements an http.Handler that answers RPC requests.
func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
		return
	}

	var ip = RealRemoteAddr(req)

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ", ip, ": ", err.Error())
		return
	}

	// filter ip
	if !server.ipWhitelist.IsAllowed(ip) {
		log.Print("Not allowed client IP: ", ip)
		conn.Close()
		return
	}

	io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	server.ServeConn(conn)
}

// HandleHTTP registers an HTTP handler for RPC messages on rpcPath.
// It is still necessary to invoke http.Serve(), typically in a go statement.
func (server *Server) HandleHTTP(rpcPath string) {
	http.Handle(rpcPath, server)
}

func RealRemoteAddr(req *http.Request) string {
	var ip string
	if ip = req.Header.Get("X-Real-IP"); len(ip) == 0 {
		if ip = req.Header.Get("X-Forwarded-For"); len(ip) == 0 {
			ip, _, _ = net.SplitHostPort(req.RemoteAddr)
		}
	}
	return ip
}
