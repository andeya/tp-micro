package server

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"net/url"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	codecGob "github.com/henrylee2cn/rpc2/codec/gob"
	"github.com/henrylee2cn/rpc2/common"
	"github.com/henrylee2cn/rpc2/log"
	"github.com/henrylee2cn/rpc2/plugin"
)

type (
	// Server represents an RPC Server.
	Server struct {
		PluginContainer IServerPluginContainer
		Timeout         time.Duration
		ReadTimeout     time.Duration
		WriteTimeout    time.Duration
		ServerCodecFunc ServerCodecFunc
		ServiceBuilder  IServiceBuilder

		serviceMap   map[string]IService
		mu           sync.RWMutex // protects the serviceMap
		routers      []string
		listener     net.Listener
		contextPool  sync.Pool
		baseMetadata string
	}

	// ServiceGroup is the group of service.
	ServiceGroup struct {
		prefixes        []string
		PluginContainer IServerPluginContainer
		server          *Server
	}
)

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer(Server{})

// NewServer returns a new Server.
func NewServer(srv Server) *Server {
	return srv.init()
}

// init initializes Server.
func (server *Server) init() *Server {
	server.routers = []string{}
	server.serviceMap = make(map[string]IService)
	server.contextPool.New = func() interface{} {
		return &Context{
			server: server,
			req:    new(rpc.Request),
			resp:   new(rpc.Response),
		}
	}
	if server.PluginContainer == nil {
		server.PluginContainer = new(ServerPluginContainer)
	}
	if server.ServerCodecFunc == nil {
		server.ServerCodecFunc = codecGob.NewGobServerCodec
	}
	if server.ServiceBuilder == nil {
		server.ServiceBuilder = NewNormServiceBuilder(new(URLFormat))
	}

	addServers(server)
	return server
}

// SetBaseMetadata sets default meta data.
// Must be called before the registration service.
// Its priority is lower than the register metadata parameter.
func (server *Server) SetBaseMetadata(metadata string) {
	server.baseMetadata = metadata
}

// Group add service group
func (server *Server) Group(prefix string, plugins ...plugin.IPlugin) *ServiceGroup {
	return (&ServiceGroup{
		server: server,
	}).Group(prefix, plugins...)
}

// Group add service group
func (group *ServiceGroup) Group(prefix string, plugins ...plugin.IPlugin) *ServiceGroup {
	if err := common.CheckSname(prefix); err != nil {
		log.Fatal("rpc: " + err.Error())
	}
	p := new(ServerPluginContainer)
	if group.PluginContainer != nil {
		p.Add(group.PluginContainer.GetAll()...)
	}
	if err := p.Add(plugins...); err != nil {
		log.Fatal("rpc: " + err.Error())
	}
	prefixes := append(group.prefixes, prefix)
	groupPath := group.server.ServiceBuilder.URIEncode(nil, prefixes...)
	for _, plugin := range plugins {
		if _, ok := plugin.(IPostConnAcceptPlugin); ok {
			log.Noticef("rpc: 'PostConnAccept()' of '%s' plugin in '%s' group is invalid", plugin.Name(), groupPath)
		}
		if _, ok := plugin.(IPreReadRequestHeaderPlugin); ok {
			log.Noticef("rpc: 'PreReadRequestHeader()' of '%s' plugin in '%s' group is invalid", plugin.Name(), groupPath)
		}
		if _, ok := plugin.(IPostReadRequestHeaderPlugin); ok {
			log.Noticef("rpc: 'PostReadRequestHeader()' of '%s' plugin in '%s' group is invalid", plugin.Name(), groupPath)
		}
	}
	return &ServiceGroup{
		prefixes:        prefixes,
		PluginContainer: p,
		server:          group.server,
	}
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
func (server *Server) Register(rcvr interface{}, metadata ...string) {
	name := common.ObjectName(rcvr)
	server.RegisterName(name, rcvr, metadata...)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, rcvr interface{}, metadata ...string) {
	if err := common.CheckSname(name); err != nil {
		log.Fatal("rpc: " + err.Error())
	}
	p := new(ServerPluginContainer)
	server.register([]string{name}, rcvr, p, metadata...)
}

// Register register service based on group
func (group *ServiceGroup) Register(rcvr interface{}, metadata ...string) {
	name := common.ObjectName(rcvr)
	group.RegisterName(name, rcvr, metadata...)
}

// RegisterName register service based on group
func (group *ServiceGroup) RegisterName(name string, rcvr interface{}, metadata ...string) {
	if err := common.CheckSname(name); err != nil {
		log.Fatal("rpc: " + err.Error())
	}
	var all []plugin.IPlugin
	if group.PluginContainer != nil {
		_plugins := group.PluginContainer.GetAll()
		all = make([]plugin.IPlugin, len(_plugins))
		copy(all, _plugins)
	}
	p := &ServerPluginContainer{
		PluginContainer: plugin.PluginContainer{
			Plugins: all,
		},
	}
	group.server.register(append(group.prefixes, name), rcvr, p, metadata...)
}

func (server *Server) register(pathSegments []string, rcvr interface{}, p IServerPluginContainer, metadata ...string) {
	server.mu.Lock()
	defer server.mu.Unlock()
	services, err := server.ServiceBuilder.NewServices(rcvr, pathSegments...)
	if err != nil {
		log.Fatal("rpc: " + err.Error())
	}
	if len(services) == 0 {
		log.Fatal("rpc: can not register invalid service: '" + reflect.ValueOf(rcvr).String() + "'")
	}
	var errs []error
	for _, service := range services {
		spath := service.GetPath()

		if _, present := server.serviceMap[spath]; present {
			errs = append(errs, common.ErrServiceAlreadyExists.Format(spath))
		}

		metadata = append(metadata, server.baseMetadata)

		var err error
		err = server.PluginContainer.doRegister(spath, rcvr, metadata...)
		if err != nil {
			errs = append(errs, common.NewError(err.Error()))
		}
		err = p.doRegister(spath, rcvr, metadata...)
		if err != nil {
			errs = append(errs, common.NewError(err.Error()))
		}

		service.SetPluginContainer(p)

		// print routers.
		server.routers = append(server.routers, spath)
		log.Infof("rpc: route ->	%s", spath)

		server.serviceMap[spath] = service
	}
	if len(errs) > 0 {
		log.Fatal("rpc: " + common.NewMultiError(errs).Error())
	}
	// sort router
	sort.Strings(server.routers)
}

// Routers return registered routers.
func (server *Server) Routers() []string {
	return server.routers
}

// Serve open RPC service at the specified network address.
func (server *Server) Serve(network, address string) {
	lis, err := makeListener(network, address)
	if err != nil {
		log.Fatal("rpc: " + err.Error())
	}
	log.Infof("rpc: listening and serving %s on %s", strings.ToUpper(network), address)
	server.ServeListener(lis)
}

// ServeTLS open secure RPC service at the specified network address.
func (server *Server) ServeTLS(network, address string, config *tls.Config) {
	lis, err := tls.Listen(network, address, config)
	if err != nil {
		log.Fatal("rpc: " + err.Error())
	}
	log.Infof("rpc: listening and serving %s on %s", strings.ToUpper(network), address)
	server.ServeListener(lis)
}

func validIP4(ipAddress string) bool {
	ipAddress = strings.Trim(ipAddress, " ")
	i := strings.LastIndex(ipAddress, ":")
	ipAddress = ipAddress[:i] //remove port

	re, _ := regexp.Compile(`^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`)
	return re.MatchString(ipAddress)
}

// ServeListener accepts connection on the listener and serves requests.
// ServeListener blocks until the listener returns a non-nil error.
// The caller typically invokes ServeListener in a go statement.
func (server *Server) ServeListener(lis net.Listener) {
	server.mu.Lock()
	server.listener = lis
	server.mu.Unlock()
	for {
		c, err := lis.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Debugf("rpc: accept: %s", err.Error())
			}
			return
		}
		conn := NewServerCodecConn(c)
		if err = server.PluginContainer.doPostConnAccept(conn); err != nil {
			log.Debugf("rpc: PostConnAccept: %s", err.Error())
			continue
		}
		go server.ServeConn(conn)
	}
}

// ServeByHTTP serves
func (server *Server) ServeByHTTP(lis net.Listener, rpcPath ...string) {
	var p = rpc.DefaultRPCPath
	if len(rpcPath) > 0 && len(rpcPath[0]) > 0 {
		p = rpcPath[0]
	}
	http.Handle(p, server)
	srv := &http.Server{Handler: nil}
	srv.Serve(lis)
}

// ServeByMux serves
func (server *Server) ServeByMux(lis net.Listener, mux *http.ServeMux, rpcPath ...string) {
	var p = rpc.DefaultRPCPath
	if len(rpcPath) > 0 && len(rpcPath[0]) > 0 {
		p = rpcPath[0]
	}
	mux.Handle(p, server)
	srv := &http.Server{Handler: mux}
	srv.Serve(lis)
}

// ServeHTTP implements an http.Handler that answers RPC requests.
func (server *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "CONNECT" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, "405 must CONNECT\n")
		return
	}

	c, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Debugf("rpc: hijacking %s: %s", req.RemoteAddr, err.Error())
		return
	}

	conn := NewServerCodecConn(c)
	if err = server.PluginContainer.doPostConnAccept(conn); err != nil {
		log.Debugf("rpc: PostConnAccept: %s", err.Error())
		return
	}

	io.WriteString(conn, "HTTP/1.0 "+common.Connected+"\n\n")
	server.ServeConn(conn)
}

// HandleHTTP registers an HTTP handler for RPC messages on rpcPath,
// and a debugging handler on debugPath.
// It is still necessary to invoke http.Serve(), typically in a go statement.
func (server *Server) HandleHTTP(rpcPath string) {
	http.Handle(rpcPath, server)
}

// Address return the listening address.
func (server *Server) Address() string {
	return server.listener.Addr().String()
}

// Close listening and serveing.
func (server *Server) Close() {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.listener.Close()
	log.Infof("rpc: stopped listening and serveing %s", server.Address())
}

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
// ServeConn uses the gob wire format (see package gob) on the
// connection. To use an alternate codec, use ServeCodec.
func (server *Server) ServeConn(conn ServerCodecConn) {
	if conn.GetServerCodec() == nil {
		conn.SetServerCodec(server.ServerCodecFunc)
	}
	sending := new(sync.Mutex)
	for {
		ctx := server.getContext(conn)
		keepReading, notSend, err := server.readRequest(ctx)
		if err != nil {
			if err != io.EOF {
				log.Debugf("rpc: %s", err.Error())
			}
			if !keepReading {
				server.putContext(ctx)
				break
			}
			// send a response if we actually managed to read a header.
			if !notSend {
				server.sendResponse(sending, ctx, err.Error())
			} else {
				server.putContext(ctx)
			}
			continue
		}
		go server.call(sending, ctx)
	}
	conn.Close()
}

// ServeRequest is like ServeConn but synchronously serves a single request.
// It does not close the codec upon completion.
func (server *Server) ServeRequest(conn ServerCodecConn) error {
	if conn.GetServerCodec() == nil {
		conn.SetServerCodec(server.ServerCodecFunc)
	}
	sending := new(sync.Mutex)
	ctx := server.getContext(conn)
	keepReading, notSend, err := server.readRequest(ctx)
	if err != nil {
		if !keepReading {
			return err
		}
		// send a response if we actually managed to read a header.
		if !notSend {
			server.sendResponse(sending, ctx, err.Error())
		} else {
			server.putContext(ctx)
		}
		return err
	}
	server.call(sending, ctx)
	return nil
}

func (server *Server) readRequest(ctx *Context) (keepReading bool, notSend bool, err error) {
	keepReading, notSend, err = ctx.readRequestHeader()
	if err != nil {
		if !keepReading {
			return
		}
		// discard body
		ctx.readRequestBody(nil)
		return
	}

	// get arg value
	argType := ctx.service.GetArgType()
	argIsValue := false // if true, need to indirect before calling.
	var argv reflect.Value
	if argType.Kind() == reflect.Ptr {
		argv = reflect.New(argType.Elem())
	} else {
		argv = reflect.New(argType)
		argIsValue = true
	}

	if argIsValue {
		ctx.argv = argv.Elem()
	} else {
		ctx.argv = argv
	}

	// Decode the argument value.
	err = ctx.readRequestBody(argv.Interface())
	return
}

func (server *Server) call(sending *sync.Mutex, ctx *Context) {
	defer func() {
		if p := recover(); p != nil {
			log.Criticalf("rpc: (%s) %v\n[PANIC]\n%s\n", ctx.Path(), p, common.PanicTrace(4))
			ctx.rpcErrorType = common.ErrorTypeServerServicePanic
			server.sendResponse(sending, ctx, "Service Panic!")
		}
	}()
	var err error
	ctx.replyv, err = ctx.service.Call(ctx.argv, ctx)
	errmsg := ""
	if err != nil {
		errmsg = err.Error()
		ctx.rpcErrorType = common.ErrorTypeServerService
	}
	server.sendResponse(sending, ctx, errmsg)
}

// A value sent as a placeholder for the server's response value when the server
// receives an invalid request. It is never decoded by the client since the Response
// contains an error when it is used.
var invalidRequest = struct{}{}

func (server *Server) sendResponse(sending *sync.Mutex, ctx *Context, errmsg string) {
	var reply interface{}
	// Encode the response header
	ctx.resp.ServiceMethod = ctx.req.ServiceMethod
	if errmsg != "" {
		ctx.resp.Error = errmsg
		reply = invalidRequest
	} else {
		reply = ctx.replyv.Interface()
	}
	ctx.resp.Seq = ctx.req.Seq
	sending.Lock()
	err := ctx.writeResponse(reply)
	if err != nil {
		log.Debugf("rpc: writing response: %s", err.Error())
	}
	sending.Unlock()
	server.putContext(ctx)
}

func (server *Server) getContext(conn ServerCodecConn) *Context {
	ctx := server.contextPool.Get().(*Context)
	ctx.Lock()
	ctx.req.ServiceMethod = ""
	ctx.req.Seq = 0
	ctx.resp.Error = ""
	ctx.resp.Seq = 0
	ctx.resp.ServiceMethod = ""
	ctx.service = nil
	ctx.codecConn = conn
	ctx.service = nil
	ctx.query = url.Values{}
	ctx.argv = reflect.Value{}
	ctx.replyv = reflect.Value{}
	ctx.Unlock()
	return ctx
}

func (server *Server) putContext(ctx *Context) {
	ctx.Lock()
	ctx.data = nil
	ctx.codecConn = nil
	ctx.Unlock()
	server.contextPool.Put(ctx)
}
