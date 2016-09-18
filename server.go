package rpc2

import (
	"bufio"
	"encoding/gob"
	"errors"
	"log"
	"net"
	"net/rpc"
	"path"
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

// Open Service
// @timeout, optional, setting server response timeout.
func (server *Server) ListenTCP() {
	l, e := net.Listen("tcp", server.addr)
	if e != nil {
		log.Fatal("Error: listen %s error:", server.addr, e)
	}
	for {
		conn, err := l.Accept()
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

func (server *Server) Group(typePrefix string, plugins ...Plugin) *Group {
	if strings.Contains(typePrefix, ".") {
		panic("Group's prefix ('" + typePrefix + "') can not contain '.'.")
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
	if strings.Contains(typePrefix, ".") {
		panic("Group's prefix ('" + typePrefix + "') can not contain '.'.")
		return nil
	}
	if strings.Contains(typePrefix, "/") {
		panic("Group's prefix ('" + typePrefix + "') can not contain '/'.")
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
	if strings.Contains(name, "/") {
		panic("RegisterName ('" + name + "') can not contain '/'.")
		return nil
	}
	return server.Server.RegisterName(name, rcvr)
}

func (group *Group) RegisterName(name string, rcvr interface{}) error {
	if strings.Contains(name, "/") {
		panic("RegisterName '" + name + "' can not contain '/'.")
		return nil
	}
	return group.Server.Server.RegisterName(path.Join(group.prefix, name), rcvr)
}

// ServeConnx runs the server on a single connection.
// ServeConnx blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
// ServeConnx uses the gob wire format (see package gob) on the
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

	if !strings.Contains(r.ServiceMethod, ".") {
		return errors.New("rpc: service/method request ill-formed: " + r.ServiceMethod)
	}

	for _, plugin := range w.Server.plugins {
		err = plugin.PostReadRequestHeader(r)
		if err != nil {
			return err
		}
	}

	var (
		p      = strings.Split(r.ServiceMethod, "/")
		prefix string
	)
	for i, count := 0, len(p)-1; i < count; i++ {
		prefix += p[i]
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
