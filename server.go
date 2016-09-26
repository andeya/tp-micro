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
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

type (
	// Server represents an RPC Server.
	Server struct {
		plugins         []Plugin
		groupMap        map[string]*Group
		timeout         time.Duration
		readTimeout     time.Duration
		writeTimeout    time.Duration
		serverCodecFunc ServerCodecFunc
		ipWhitelist     *IPWhitelist
		routers         []string
		printRouters    bool

		mu         sync.RWMutex // protects the serviceMap
		serviceMap map[string]*service
		reqLock    sync.Mutex // protects freeReq
		freeReq    *Request
		respLock   sync.Mutex // protects freeResp
		freeResp   *Response
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

	// ServerCodecFunc is used to create a rpc.ServerCodec from io.ReadWriteCloser.
	ServerCodecFunc func(io.ReadWriteCloser) rpc.ServerCodec

	// WARN:Request must be consistent with the standard package, can not be changed!
	// Request is a header written before every RPC call. It is used internally
	// but documented here as an aid to debugging, such as when analyzing
	// network traffic.
	Request struct {
		ServiceMethod string   // format: "Service.Method"
		Seq           uint64   // sequence number chosen by client
		next          *Request // for free list in Server
	}

	// WARN:Response must be consistent with the standard package, can not be changed!
	// Response is a header written before every RPC return. It is used internally
	// but documented here as an aid to debugging, such as when analyzing
	// network traffic.
	Response struct {
		ServiceMethod string    // echoes that of the Request
		Seq           uint64    // echoes that of the request
		Error         string    // error, if any.
		next          *Response // for free list in Server
	}

	methodType struct {
		sync.Mutex // protects counters
		method     reflect.Method
		ArgType    reflect.Type
		ReplyType  reflect.Type
		numCalls   uint
	}

	service struct {
		name   string                 // name of service
		rcvr   reflect.Value          // receiver of methods for the service
		typ    reflect.Type           // type of the receiver
		method map[string]*methodType // registered methods
	}
)

// NewServer returns a new Server.
func NewServer(
	timeout time.Duration,
	readTimeout time.Duration,
	writeTimeout time.Duration,
	serverCodecFunc ServerCodecFunc,
	printRouters ...bool,
) *Server {
	if serverCodecFunc == nil {
		serverCodecFunc = NewGobServerCodec
	}
	if len(printRouters) == 0 {
		printRouters = append(printRouters, false)
	}
	return &Server{
		routers:         []string{},
		printRouters:    printRouters[0],
		serviceMap:      make(map[string]*service),
		plugins:         []Plugin{},
		timeout:         timeout,
		readTimeout:     readTimeout,
		writeTimeout:    writeTimeout,
		serverCodecFunc: serverCodecFunc,
		ipWhitelist:     NewIPWhitelist(),
		groupMap:        map[string]*Group{},
	}
}

// NewDefaultServer returns a new default Server.
func NewDefaultServer(printRouters ...bool) *Server {
	return NewServer(time.Minute, 0, 0, nil, printRouters...)
}

func (server *Server) Plugin(plugins ...Plugin) error {
	for _, plugin := range plugins {
		if plugin == nil {
			return errors.New("rpc.Plugin: plugins can not contain nil")
		}
	}
	server.plugins = append(server.plugins, plugins...)
	return nil
}

func (server *Server) Group(typePrefix string, plugins ...Plugin) (*Group, error) {
	return (&Group{
		plugins: []Plugin{},
		Server:  server,
	}).Group(typePrefix, plugins...)
}

func (group *Group) Group(typePrefix string, plugins ...Plugin) (*Group, error) {
	if !nameRegexp.MatchString(typePrefix) {
		return nil, errors.New("rpc.Group: group's prefix ('" + typePrefix + "') must conform to the regular expression '/^[a-zA-Z0-9_\\.]+$/'.")
	}
	for _, plugin := range plugins {
		if plugin == nil {
			return nil, errors.New("rpc.Group: plugins can not contain nil")
		}
	}
	g := &Group{
		prefix:  path.Join(group.prefix, typePrefix) + "/",
		plugins: plugins,
		Server:  group.Server,
	}
	g.Server.groupMap[g.prefix] = g
	return g, nil
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
func (server *Server) Register(rcvr interface{}, plugins ...Plugin) error {
	return server.register(rcvr, "", false, plugins...)
}

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func (server *Server) RegisterName(name string, rcvr interface{}, plugins ...Plugin) error {
	if !nameRegexp.MatchString(name) {
		return errors.New("rpc.RegisterName name ('" + name + "') must conform to the regular expression '/^[a-zA-Z0-9_\\.]+$/'.")
	}
	return server.register(rcvr, name, true, plugins...)
}

func (group *Group) Register(rcvr interface{}, plugins ...Plugin) error {
	name := reflect.Indirect(reflect.ValueOf(rcvr)).Type().Name()
	return group.Server.register(rcvr, path.Join(group.prefix, name), true, plugins...)
}

func (group *Group) RegisterName(name string, rcvr interface{}, plugins ...Plugin) error {
	if !nameRegexp.MatchString(name) {
		return errors.New("rpc.Group.RegisterName name ('" + name + "') must conform to the regular expression '/^[a-zA-Z0-9_\\.]+$/'.")
	}
	return group.Server.register(rcvr, path.Join(group.prefix, name), true, plugins...)
}

// IP return ip whitelist object.
func (server *Server) IP() *IPWhitelist {
	return server.ipWhitelist
}

// Routers return registered routers.
func (server *Server) Routers() []string {
	return server.routers
}

// ListenAndServe open RPC service at the specified network address.
func (server *Server) ListenAndServe(network, address string) {
	lis, err := net.Listen(network, address)
	if err != nil {
		log.Fatal("[RPC] listen %s error:", address, err)
	}
	if server.printRouters {
		log.Printf("[RPC] listening and serving %s on %s", strings.ToUpper(network), address)
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
			log.Println("[RPC] accept:", err.Error())
			return
		}

		// filter ip
		ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
		if !server.ipWhitelist.IsAllowed(ip) {
			log.Println("[RPC] not allowed client ip:", ip)
			conn.Close()
			continue
		}

		go server.ServeConn(conn)
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

	var ip = RealRemoteAddr(req)

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		log.Print("[RPC] hijacking ", ip, ": ", err.Error())
		return
	}

	// filter ip
	if !server.ipWhitelist.IsAllowed(ip) {
		log.Println("[RPC] not allowed client ip:", ip)
		conn.Close()
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

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
// ServeConn uses the gob wire format (see package gob) on the
// connection. To use an alternate codec, use ServeCodec.
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	codec := server.wrapServerCodec(conn)
	sending := new(sync.Mutex)
	for {
		service, mtype, req, argv, replyv, keepReading, err := server.readRequest(codec)
		if err != nil {
			if debugLog && err != io.EOF {
				log.Println("[RPC]", err.Error())
			}
			if !keepReading {
				break
			}
			// send a response if we actually managed to read a header.
			if req != nil {
				server.sendResponse(sending, req, invalidRequest, codec, err.Error())
				server.freeRequest(req)
			}
			continue
		}
		go service.call(server, sending, mtype, req, argv, replyv, codec)
	}
	codec.Close()
}

// ServeRequest is like ServeConn but synchronously serves a single request.
// It does not close the codec upon completion.
func (server *Server) ServeRequest(conn io.ReadWriteCloser) error {
	codec := server.wrapServerCodec(conn)
	sending := new(sync.Mutex)
	service, mtype, req, argv, replyv, keepReading, err := server.readRequest(codec)
	if err != nil {
		if !keepReading {
			return err
		}
		// send a response if we actually managed to read a header.
		if req != nil {
			server.sendResponse(sending, req, invalidRequest, codec, err.Error())
			server.freeRequest(req)
		}
		return err
	}
	service.call(server, sending, mtype, req, argv, replyv, codec)
	return nil
}

func (server *Server) wrapServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return &serverCodecWrapper{
		ServerCodec:  server.serverCodecFunc(conn),
		conn:         conn,
		Server:       server,
		groupPlugins: make([]Plugin, 0, 10),
	}
}

func (server *Server) register(rcvr interface{}, name string, useName bool, plugins ...Plugin) error {
	for _, plugin := range plugins {
		if plugin == nil {
			str := "rpc.Register: plugins can not contain nil"
			// log.Println("[RPC]", str)
			return errors.New(str)
		}
	}
	server.mu.Lock()
	defer server.mu.Unlock()
	if server.serviceMap == nil {
		server.serviceMap = make(map[string]*service)
	}
	s := new(service)
	s.typ = reflect.TypeOf(rcvr)
	s.rcvr = reflect.ValueOf(rcvr)
	sname := reflect.Indirect(s.rcvr).Type().Name()
	if useName {
		sname = name
	}
	if sname == "" {
		s := "rpc.Register: no service name for type " + s.typ.String()
		// log.Println("[RPC]", s)
		return errors.New(s)
	}
	if !isExported(sname) && !useName {
		s := "rpc.Register: type " + sname + " is not exported"
		// log.Println("[RPC]", s)
		return errors.New(s)
	}
	if _, present := server.serviceMap[sname]; present {
		return errors.New("rpc: service already defined: " + sname)
	}
	s.name = sname

	// Install the methods
	s.method = suitableMethods(s.typ, true)

	if len(s.method) == 0 {
		str := ""

		// To help the user, see if a pointer receiver would work.
		method := suitableMethods(reflect.PtrTo(s.typ), false)
		if len(method) != 0 {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type (hint: pass a pointer to value of that type)"
		} else {
			str = "rpc.Register: type " + sname + " has no exported methods of suitable type"
		}
		// log.Println("[RPC]", str)
		return errors.New(str)
	}
	server.serviceMap[s.name] = s

	// register Plugin.
	g := &Group{
		prefix:  s.name,
		plugins: plugins,
		Server:  server,
	}
	// register itself as Plugin.
	if plugin, ok := rcvr.(Plugin); ok {
		g.plugins = append(g.plugins, plugin)
	}
	server.groupMap[g.prefix] = g

	// record routers and sort it.
	for m := range s.method {
		router := s.name + "." + m
		server.routers = append(server.routers, router)
		if server.printRouters {
			log.Printf("[RPC ROUTER] %s", router)
		}
	}
	sort.Strings(server.routers)
	return nil
}

func (server *Server) getRequest() *Request {
	server.reqLock.Lock()
	req := server.freeReq
	if req == nil {
		req = new(Request)
	} else {
		server.freeReq = req.next
		*req = Request{}
	}
	server.reqLock.Unlock()
	return req
}

func (server *Server) freeRequest(req *Request) {
	server.reqLock.Lock()
	req.next = server.freeReq
	server.freeReq = req
	server.reqLock.Unlock()
}

func (server *Server) getResponse() *Response {
	server.respLock.Lock()
	resp := server.freeResp
	if resp == nil {
		resp = new(Response)
	} else {
		server.freeResp = resp.next
		*resp = Response{}
	}
	server.respLock.Unlock()
	return resp
}

func (server *Server) freeResponse(resp *Response) {
	server.respLock.Lock()
	resp.next = server.freeResp
	server.freeResp = resp
	server.respLock.Unlock()
}

func (server *Server) readRequest(codec rpc.ServerCodec) (service *service, mtype *methodType, req *Request, argv, replyv reflect.Value, keepReading bool, err error) {
	service, mtype, req, keepReading, err = server.readRequestHeader(codec)
	if err != nil {
		if !keepReading {
			return
		}
		// discard body
		codec.ReadRequestBody(nil)
		return
	}

	// Decode the argument value.
	argIsValue := false // if true, need to indirect before calling.
	if mtype.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(mtype.ArgType.Elem())
	} else {
		argv = reflect.New(mtype.ArgType)
		argIsValue = true
	}
	// argv guaranteed to be a pointer now.
	if err = codec.ReadRequestBody(argv.Interface()); err != nil {
		return
	}
	if argIsValue {
		argv = argv.Elem()
	}

	replyv = reflect.New(mtype.ReplyType.Elem())
	return
}

func (server *Server) readRequestHeader(codec rpc.ServerCodec) (service *service, mtype *methodType, req *Request, keepReading bool, err error) {
	// Grab the request header.
	req = server.getRequest()
	err = codec.ReadRequestHeader((*rpc.Request)(unsafe.Pointer(req)))
	// if err != nil {
	// 	req = nil
	// 	if err == io.EOF || err == io.ErrUnexpectedEOF {
	// 		return
	// 	}
	// 	err = errors.New("rpc: server cannot decode request: " + err.Error())
	// 	return
	// }
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			req = nil
			return
		}
		keepReading = true
		err = errors.New("rpc: server cannot decode request: " + err.Error())
		return
	}

	// We read the header successfully. If we see an error now,
	// we can still recover and move on to the next request.
	keepReading = true

	dot := strings.LastIndex(req.ServiceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc: service/method request ill-formed: " + req.ServiceMethod)
		return
	}
	serviceName := req.ServiceMethod[:dot]
	methodName := req.ServiceMethod[dot+1:]

	// Look up the request.
	server.mu.RLock()
	service = server.serviceMap[serviceName]
	server.mu.RUnlock()
	if service == nil {
		err = errors.New("rpc: can't find service " + req.ServiceMethod)
		return
	}
	mtype = service.method[methodName]
	if mtype == nil {
		err = errors.New("rpc: can't find method " + req.ServiceMethod)
	}
	return
}

// A value sent as a placeholder for the server's response value when the server
// receives an invalid request. It is never decoded by the client since the Response
// contains an error when it is used.
var invalidRequest = struct{}{}

func (server *Server) sendResponse(sending *sync.Mutex, req *Request, reply interface{}, codec rpc.ServerCodec, errmsg string) {
	resp := server.getResponse()
	// Encode the response header
	resp.ServiceMethod = req.ServiceMethod
	if errmsg != "" {
		resp.Error = errmsg
		reply = invalidRequest
	}
	resp.Seq = req.Seq
	sending.Lock()
	err := codec.WriteResponse((*rpc.Response)(unsafe.Pointer(resp)), reply)
	if debugLog && err != nil {
		log.Println("[RPC] writing response:", err.Error())
	}
	sending.Unlock()
	server.freeResponse(resp)
}

func (m *methodType) NumCalls() (n uint) {
	m.Lock()
	n = m.numCalls
	m.Unlock()
	return n
}

func (s *service) call(server *Server, sending *sync.Mutex, mtype *methodType, req *Request, argv, replyv reflect.Value, codec rpc.ServerCodec) {
	mtype.Lock()
	mtype.numCalls++
	mtype.Unlock()
	function := mtype.method.Func
	// Invoke the method, providing a new value for the reply.
	returnValues := function.Call([]reflect.Value{s.rcvr, argv, replyv})
	// The return value for the method is an error.
	errInter := returnValues[0].Interface()
	errmsg := ""
	if errInter != nil {
		errmsg = errInter.(error).Error()
	}
	server.sendResponse(sending, req, replyv.Interface(), codec, errmsg)
	server.freeRequest(req)
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

// DefaultServer is the default instance of *Server.
var DefaultServer = NewDefaultServer()

// Register publishes the receiver's methods in the DefaultServer.
func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

// RegisterName is like Register but uses the provided name for the type
// instead of the receiver's concrete type.
func RegisterName(name string, rcvr interface{}) error {
	return DefaultServer.RegisterName(name, rcvr)
}

// ServeConn runs the DefaultServer on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
// The caller typically invokes ServeConn in a go statement.
// ServeConn uses the gob wire format (see package gob) on the
// connection. To use an alternate codec, use ServeCodec.
func ServeConn(conn io.ReadWriteCloser) {
	DefaultServer.ServeConn(conn)
}

// ServeRequest is like ServeConn but synchronously serves a single request.
// It does not close the codec upon completion.
func ServeRequest(conn io.ReadWriteCloser) error {
	return DefaultServer.ServeRequest(conn)
}

// Accept accepts connections on the listener and serves requests
// to DefaultServer for each incoming connection.
// Accept blocks; the caller typically invokes it in a go statement.
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

// HandleHTTP registers an HTTP handler for RPC messages to DefaultServer
// on rpc.DefaultRPCPath and a debugging handler on rpc.DefaultDebugPath.
// It is still necessary to invoke http.Serve(), typically in a go statement.
func HandleHTTP() {
	DefaultServer.HandleHTTP(rpc.DefaultRPCPath, rpc.DefaultDebugPath)
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
