package rpc2

import (
	"bufio"
	"encoding/gob"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"path"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type (
	Server struct {
		addr string
		*rpc.Server
		plugins         []Plugin
		groupMap        map[string]*Group
		timeout         time.Duration
		readTimeout     time.Duration
		writeTimeout    time.Duration
		serverCodecFunc ServerCodecFunc
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
		conn         net.Conn
		groupPlugins []Plugin
		*Server
	}

	// ServerCodecFunc is used to create a rpc.ServerCodec from net.Conn.
	ServerCodecFunc func(net.Conn) rpc.ServerCodec

	debugHTTP struct {
		*Server
	}
)

func NewServer(
	addr string,
	timeout,
	readTimeout,
	writeTimeout time.Duration,
	serverCodecFunc ServerCodecFunc,
) *Server {
	if serverCodecFunc == nil {
		serverCodecFunc = func(conn net.Conn) rpc.ServerCodec {
			buf := bufio.NewWriter(conn)
			return &gobServerCodec{
				rwc:    conn,
				dec:    gob.NewDecoder(conn),
				enc:    gob.NewEncoder(buf),
				encBuf: buf,
			}
		}
	}

	return &Server{
		addr:            addr,
		Server:          rpc.NewServer(),
		plugins:         []Plugin{},
		timeout:         timeout,
		readTimeout:     readTimeout,
		writeTimeout:    writeTimeout,
		serverCodecFunc: serverCodecFunc,
		groupMap:        map[string]*Group{},
	}
}

func NewDefaultServer(addr string) *Server {
	return NewServer(addr, time.Minute, 0, 0, nil)
}

func (server *Server) Group(typePrefix string, plugins ...Plugin) *Group {
	if !nameRegexp.MatchString(typePrefix) {
		panic("Group's prefix ('" + typePrefix + "') must conform to the regular expression '/^[a-zA-Z0-9_]+$/'.")
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
		panic("Group's prefix ('" + typePrefix + "') must conform to the regular expression '/^[a-zA-Z0-9_]+$/'.")
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
		panic("RegisterName ('" + name + "') must conform to the regular expression '/^[a-zA-Z0-9_]+$/'.")
		return nil
	}
	return server.Server.RegisterName(name, rcvr)
}

func (group *Group) RegisterName(name string, rcvr interface{}) error {
	if !nameRegexp.MatchString(name) {
		panic("RegisterName ('" + name + "') must conform to the regular expression '/^[a-zA-Z0-9_]+$/'.")
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
// ServeConn uses the gob wire format (see package gob) on the
// connection.
func (server *Server) ServeConn(conn net.Conn) {
	srv := server.serverCodecFunc(conn)
	server.ServeCodec(conn, srv)
}

// ServeCodec is like ServeConn but uses the specified codec to
// decode requests and encode responses.
func (server *Server) ServeCodec(conn net.Conn, codec rpc.ServerCodec) {
	server.Server.ServeCodec(server.wrapServerCodec(conn, codec))
}

// ServeRequest is like ServeCodec but synchronously serves a single request.
// It does not close the codec upon completion.
func (server *Server) ServeRequest(conn net.Conn, codec rpc.ServerCodec) error {
	return server.Server.ServeRequest(server.wrapServerCodec(conn, codec))
}

func (server *Server) wrapServerCodec(conn net.Conn, codec rpc.ServerCodec) *serverCodecWrapper {
	return &serverCodecWrapper{
		ServerCodec:  codec,
		conn:         conn,
		Server:       server,
		groupPlugins: make([]Plugin, 0, 50),
	}
}

func (w *serverCodecWrapper) ReadRequestHeader(r *rpc.Request) error {
	if w.Server.timeout > 0 {
		w.conn.SetDeadline(time.Now().Add(w.Server.timeout))
	}
	if w.Server.readTimeout > 0 {
		w.conn.SetReadDeadline(time.Now().Add(w.Server.readTimeout))
	}

	err := w.ServerCodec.ReadRequestHeader(r)

	if err != nil {
		return err
	}

	var dot = strings.Index(r.ServiceMethod, ".")
	if dot < 0 || dot+1 == len(r.ServiceMethod) {
		return errors.New("rpc: service/method request ill-formed: " + r.ServiceMethod)
	}

	for _, plugin := range w.Server.plugins {
		err = plugin.PostReadRequestHeader(r)
		if err != nil {
			return err
		}
	}

	var serviceName = r.ServiceMethod[:dot]

	var p = strings.Split(serviceName, "/")
	var prefix string
	for i, count := 0, len(p)-1; i < count; i++ {
		if i == 0 {
			prefix = p[i]
		} else {
			prefix = prefix + "/" + p[i]
		}
		group, ok := w.groupMap[prefix]
		if !ok {
			return errors.New("rpc: can't find group " + prefix)
		}
		for _, plugin := range group.plugins {
			err = plugin.PostReadRequestHeader(r)
			if err != nil {
				return err
			}
			w.groupPlugins = append(w.plugins, plugin)
		}
	}

	var methodName = r.ServiceMethod[dot+1:]
	boundary := strings.IndexFunc(methodName, nameCharsFunc)
	if boundary == 0 {
		return errors.New("rpc: service/method request ill-formed: " + r.ServiceMethod)
	}
	if boundary > 0 {
		// methodName = methodName[:boundary]
		r.ServiceMethod = serviceName + "." + methodName[:boundary]
	}

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
	if w.Server.timeout > 0 {
		w.conn.SetDeadline(time.Now().Add(w.Server.timeout))
	}
	if w.Server.writeTimeout > 0 {
		w.conn.SetWriteDeadline(time.Now().Add(w.Server.writeTimeout))
	}

	err := w.ServerCodec.WriteResponse(resp, body)

	return err
}

var nameRegexp = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

func nameCharsFunc(r rune) bool {
	// A-Z
	if r >= 65 && r <= 90 {
		return false
	}
	// a-z
	if r >= 97 && r <= 122 {
		return false
	}
	// _
	if r == 95 {
		return false
	}
	// 0-9
	if r >= 48 && r <= 57 {
		return false
	}

	return true
}

// Open Service
// @timeout, optional, setting server response timeout.
func (server *Server) ListenTCP() {
	lis, err := net.Listen("tcp", server.addr)
	if err != nil {
		log.Fatal("Error: listen %s error:", server.addr, err)
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
			log.Println("Error: accept rpc connection", err.Error())
			continue
		}

		// filter ip
		if !ipWhitelist.allowAccess(conn.RemoteAddr().String()) {
			log.Println("Client Not allow:", conn.RemoteAddr().String())
			continue
		}

		go server.ServeCodec(conn, server.serverCodecFunc(conn))
	}
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
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("rpc hijacking ", req.RemoteAddr, ": ", err.Error())
		return
	}
	io.WriteString(conn, "HTTP/1.0 "+connected+"\n\n")
	server.ServeConn(conn)
}

// HandleHTTP registers an HTTP handler for RPC messages on rpcPath,
// and a debugging handler on debugPath.
// It is still necessary to invoke http.Serve(), typically in a go statement.
func (server *Server) HandleHTTP(rpcPath, debugPath string) {
	http.Handle(rpcPath, server)
	http.Handle(debugPath, debugHTTP{server})
}
