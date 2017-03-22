package rpc2

import (
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"net/rpc"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	codecGob "github.com/henrylee2cn/rpc2/codec/gob"
)

type (
	// Server represents an RPC Server.
	Server struct {
		Log             Logger
		serviceMap      map[string]*service
		mu              sync.RWMutex // protects the serviceMap
		PluginContainer IServerPluginContainer
		timeout         time.Duration
		readTimeout     time.Duration
		writeTimeout    time.Duration
		serverCodecFunc ServerCodecFunc
		routers         []string
		printRouters    bool
		listener        net.Listener
		contextPool     sync.Pool
	}

	// ServiceGroup is the group of service.
	ServiceGroup struct {
		prefix          string
		PluginContainer IServerPluginContainer
		server          *Server
	}

	// ServerCodecFunc is used to create a ServerCodec from io.ReadWriteCloser.
	ServerCodecFunc func(io.ReadWriteCloser) rpc.ServerCodec
	service         struct {
		name            string                 // name of service
		rcvr            reflect.Value          // receiver of methods for the service
		typ             reflect.Type           // type of the receiver
		method          map[string]*methodType // registered methods
		pluginContainer IServerPluginContainer
	}
	methodType struct {
		sync.Mutex // protects counters
		method     reflect.Method
		ArgType    reflect.Type
		ReplyType  reflect.Type
		numCalls   uint
	}
)

// DefaultServer is the default instance of *Server.
var DefaultServer = NewDefaultServer()

// NewDefaultServer returns a new default Server.
func NewDefaultServer(printRouters ...bool) *Server {
	return NewServer(time.Minute, 0, 0, nil, printRouters...)
}

// NewServer returns a new Server.
func NewServer(
	timeout time.Duration,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	serverCodecFunc ServerCodecFunc,
	printRouters ...bool,
) *Server {
	if serverCodecFunc == nil {
		serverCodecFunc = codecGob.NewGobServerCodec
	}
	if len(printRouters) == 0 {
		printRouters = append(printRouters, false)
	}
	server := &Server{
		Log:             newDefaultLogger(),
		routers:         []string{},
		printRouters:    printRouters[0],
		serviceMap:      make(map[string]*service),
		PluginContainer: new(ServerPluginContainer),
		timeout:         timeout,
		readTimeout:     readTimeout,
		writeTimeout:    writeTimeout,
		serverCodecFunc: serverCodecFunc,
	}
	server.contextPool.New = func() interface{} {
		return &Context{
			server: server,
			Log:    server.Log,
			req:    new(rpc.Request),
			resp:   new(rpc.Response),
		}
	}
	addServers(server)
	return server
}

// Group add service group
func (server *Server) Group(prefix string, plugins ...IPlugin) (*ServiceGroup, error) {
	return (&ServiceGroup{
		server: server,
	}).Group(prefix, plugins...)
}

// Group add service group
func (group *ServiceGroup) Group(prefix string, plugins ...IPlugin) (*ServiceGroup, error) {
	if !nameRegexp.MatchString(prefix) {
		return nil, ErrInvalidPath.Format(prefix)
	}
	p := new(ServerPluginContainer)
	if group.PluginContainer != nil {
		p.Add(group.PluginContainer.GetAll()...)
	}
	if err := p.Add(plugins...); err != nil {
		return nil, err
	}
	return &ServiceGroup{
		prefix:          path.Join(group.prefix, prefix) + "/",
		PluginContainer: p,
		server:          group.server,
	}, nil
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
func (server *Server) Register(rcvr interface{}, metadata ...string) error {
	name := SnakeString(ObjectName(rcvr))
	return server.RegisterName(name, rcvr, metadata...)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, rcvr interface{}, metadata ...string) error {
	if err := CheckSname(name); err != nil {
		return err
	}
	p := new(ServerPluginContainer)
	return server.register(name, rcvr, p, metadata...)
}

// Register register service based on group
func (group *ServiceGroup) Register(rcvr interface{}, metadata ...string) error {
	name := SnakeString(ObjectName(rcvr))
	return group.RegisterName(name, rcvr, metadata...)
}

// RegisterName register service based on group
func (group *ServiceGroup) RegisterName(name string, rcvr interface{}, metadata ...string) error {
	if err := CheckSname(name); err != nil {
		return err
	}
	name = path.Join(group.prefix, name)
	var all []IPlugin
	if group.PluginContainer != nil {
		_plugins := group.PluginContainer.GetAll()
		all = make([]IPlugin, len(_plugins))
		copy(all, _plugins)
	}
	p := &ServerPluginContainer{
		PluginContainer: PluginContainer{
			plugins: all,
		},
	}
	return group.server.register(name, rcvr, p, metadata...)
}

func (server *Server) register(spath string, rcvr interface{}, p IServerPluginContainer, metadata ...string) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	for _, plugin := range p.GetAll() {
		if _, ok := plugin.(IPostConnAcceptPlugin); ok {
			server.Log.Warnf("The method 'PostConnAccept()' of the plugin '%s' in the service '%s' will not be executed!", plugin.Name(), spath)
		}
		if _, ok := plugin.(IPreReadRequestHeaderPlugin); ok {
			server.Log.Warnf("The method 'PreReadRequestHeader()' of the plugin '%s' in the service '%s' will not be executed!", plugin.Name(), spath)
		}
	}

	if server.serviceMap == nil {
		server.serviceMap = make(map[string]*service)
	} else if _, present := server.serviceMap[spath]; present {
		return ErrServiceAlreadyExists.Format(spath)
	}

	var err error
	err = server.PluginContainer.doRegister(spath, rcvr, metadata...)
	if err != nil {
		return err
	}
	err = p.doRegister(spath, rcvr, metadata...)
	if err != nil {
		return err
	}

	s := &service{
		pluginContainer: p,
		name:            spath,
		typ:             reflect.TypeOf(rcvr),
		rcvr:            reflect.ValueOf(rcvr),
	}
	s.method = suitableMethods(s.typ, true)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PtrTo(s.typ), false)
		if len(method) != 0 {
			str = "rpc.Register: type " + spath + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + spath + " has no exported methods of suitable type"
		}
		return errors.New(str)
	}
	server.serviceMap[s.name] = s

	// record routers and sort it.
	for m := range s.method {
		router := s.name + "." + m
		server.routers = append(server.routers, router)
		if server.printRouters {
			server.Log.Infof("[RPC ROUTER] %s", router)
		}
	}
	sort.Strings(server.routers)
	return nil
}

// Routers return registered routers.
func (server *Server) Routers() []string {
	return server.routers
}

// Serve open RPC service at the specified network address.
func (server *Server) Serve(network, address string) {
	lis, err := makeListener(network, address)
	if err != nil {
		server.Log.Fatalf("[RPC] %v", err)
	}
	if server.printRouters {
		server.Log.Infof("[RPC] listening and serving %s on %s", strings.ToUpper(network), address)
	}
	server.ServeListener(lis)
}

// ServeTLS open secure RPC service at the specified network address.
func (server *Server) ServeTLS(network, address string, config *tls.Config) {
	lis, err := tls.Listen(network, address, config)
	if err != nil {
		server.Log.Fatalf("[RPC] %v", err)
	}
	if server.printRouters {
		server.Log.Infof("[RPC] listening and serving %s on %s", strings.ToUpper(network), address)
	}
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
				server.Log.Infof("[RPC] accept: %v", err)
			}
			return
		}
		conn := NewServerCodecConn(c)
		if err = server.PluginContainer.doPostConnAccept(conn); err != nil {
			server.Log.Infof("[RPC] PostConnAccept: %s", err.Error())
			continue
		}
		go server.ServeConn(conn)
	}
}

// ServeByHTTP serves
func (server *Server) ServeByHTTP(lis net.Listener, rpcPath string) {
	http.Handle(rpcPath, server)
	srv := &http.Server{Handler: nil}
	srv.Serve(lis)
}

// ServeByMux serves
func (server *Server) ServeByMux(lis net.Listener, rpcPath string, mux *http.ServeMux) {
	mux.Handle(rpcPath, server)
	srv := &http.Server{Handler: mux}
	srv.Serve(lis)
}

// Can connect to RPC service using HTTP CONNECT to rpcPath.
const connected = "200 Connected to Go RPC"

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
		server.Log.Infof("[RPC] hijacking %s: %v", req.RemoteAddr, err)
		return
	}

	conn := NewServerCodecConn(c)
	if err = server.PluginContainer.doPostConnAccept(conn); err != nil {
		server.Log.Infof("[RPC] PostConnAccept: %s", err.Error())
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

// Address return the listening address.
func (server *Server) Address() string {
	return server.listener.Addr().String()
}

// Close listening and serveing.
func (server *Server) Close() {
	server.mu.Lock()
	defer server.mu.Unlock()
	server.listener.Close()
	server.Log.Infof("[RPC] stopped listening and serveing %s", server.Address())
}

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
// ServeConn uses the gob wire format (see package gob) on the
// connection. To use an alternate codec, use ServeCodec.
func (server *Server) ServeConn(conn ServerCodecConn) {
	if conn.GetServerCodec() == nil {
		conn.SetServerCodec(server.serverCodecFunc)
	}
	sending := new(sync.Mutex)
	for {
		ctx := server.getContext(conn)
		keepReading, notSend, err := server.readRequest(ctx)
		if err != nil {
			if debugLog && err != io.EOF {
				server.Log.Infof("[RPC] %s", err.Error())
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
		conn.SetServerCodec(server.serverCodecFunc)
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

	var mtype = ctx.mtype
	// Decode the argument value.
	argIsValue := false // if true, need to indirect before calling.
	if mtype.ArgType.Kind() == reflect.Ptr {
		ctx.argv = reflect.New(mtype.ArgType.Elem())
	} else {
		ctx.argv = reflect.New(mtype.ArgType)
		argIsValue = true
	}
	// argv guaranteed to be a pointer now.
	if err = ctx.readRequestBody(ctx.argv.Interface()); err != nil {
		return
	}
	if argIsValue {
		ctx.argv = ctx.argv.Elem()
	}

	ctx.replyv = reflect.New(mtype.ReplyType.Elem())
	return
}

func (server *Server) call(sending *sync.Mutex, ctx *Context) {
	ctx.mtype.Lock()
	ctx.mtype.numCalls++
	ctx.mtype.Unlock()
	function := ctx.mtype.method.Func
	// Invoke the method, providing a new value for the reply.
	returnValues := function.Call([]reflect.Value{ctx.service.rcvr, ctx.argv, ctx.replyv})
	// The return value for the method is an error.
	errInter := returnValues[0].Interface()
	errmsg := ""
	if errInter != nil {
		errmsg = errInter.(error).Error()
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
	if debugLog && err != nil {
		server.Log.Infof("[RPC] writing response: %s", err.Error())
	}
	sending.Unlock()
	server.putContext(ctx)
}

func (m *methodType) NumCalls() (n uint) {
	m.Lock()
	n = m.numCalls
	m.Unlock()
	return n
}

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

// Precompute the reflect type for error. Can't use error directly
// because Typeof takes an empty interface value. This is annoying.
var typeOfError = reflect.TypeOf((*error)(nil)).Elem()

// suitableMethods returns suitable Rpc methods of typ, it will report
// error using log if reportErr is true.
func suitableMethods(typ reflect.Type, reportErr bool) map[string]*methodType {
	methods := make(map[string]*methodType)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		// Method must be exported.
		if method.PkgPath != "" {
			continue
		}
		// Method needs three ins: receiver, *args, *reply.
		if mtype.NumIn() != 3 {
			if reportErr {
				// log.Println("[RPC] method", mname, "has wrong number of ins:", mtype.NumIn())
			}
			continue
		}
		// First arg need not be a pointer.
		argType := mtype.In(1)
		if !isExportedOrBuiltinType(argType) {
			if reportErr {
				// log.Println("[RPC]", mname, "argument type not exported:", argType)
			}
			continue
		}
		// Second arg must be a pointer.
		replyType := mtype.In(2)
		if replyType.Kind() != reflect.Ptr {
			if reportErr {
				// log.Println("[RPC] method", mname, "reply type not a pointer:", replyType)
			}
			continue
		}
		// Reply type must be exported.
		if !isExportedOrBuiltinType(replyType) {
			if reportErr {
				// log.Println("[RPC] method", mname, "reply type not exported:", replyType)
			}
			continue
		}
		// Method needs one out.
		if mtype.NumOut() != 1 {
			if reportErr {
				// log.Println("[RPC] method", mname, "has wrong number of outs:", mtype.NumOut())
			}
			continue
		}
		// The return type of the method must be error.
		if returnType := mtype.Out(0); returnType != typeOfError {
			if reportErr {
				// log.Println("[RPC] method", mname, "returns", returnType.String(), "not error")
			}
			continue
		}
		methods[mname] = &methodType{method: method, ArgType: argType, ReplyType: replyType}
	}
	return methods
}

func (server *Server) getContext(conn ServerCodecConn) *Context {
	ctx := server.contextPool.Get().(*Context)
	ctx.Lock()
	ctx.req.ServiceMethod = ""
	ctx.req.Seq = 0
	ctx.resp.Error = ""
	ctx.resp.Seq = 0
	ctx.resp.ServiceMethod = ""
	ctx.data = make(map[interface{}]interface{})
	ctx.Path = ""
	ctx.Query = nil
	ctx.codecConn = conn
	ctx.service = nil
	ctx.mtype = nil
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
