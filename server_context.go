package rpc2

import (
	"io"
	"net"
	"net/rpc"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"
)

type (
	// ServerCodecConn net.Conn with ServerCodec
	ServerCodecConn interface {
		net.Conn
		SetConn(net.Conn)
		GetConn() net.Conn

		// ServerCodec
		ReadRequestHeader(*rpc.Request) error
		ReadRequestBody(interface{}) error
		// WriteResponse must be safe for concurrent use by multiple goroutines.
		WriteResponse(*rpc.Response, interface{}) error

		GetServerCodec() rpc.ServerCodec

		//Must ensure that both Conn and ServerCodecFunc are not nil
		SetServerCodec(ServerCodecFunc)
	}
	serverCodecConn struct {
		net.Conn
		rpc.ServerCodec
	}
)

// NewServerCodecConn get a ServerCodecConn.
func NewServerCodecConn(conn net.Conn) ServerCodecConn {
	return &serverCodecConn{Conn: conn}
}

func (conn *serverCodecConn) SetConn(c net.Conn) {
	conn.Conn = c
}

func (conn *serverCodecConn) GetConn() net.Conn {
	return conn.Conn
}

// SetServerCodec must ensure that both Conn and ServerCodecFunc are not nil
func (conn *serverCodecConn) SetServerCodec(fn ServerCodecFunc) {
	if fn != nil && conn.Conn != nil {
		conn.ServerCodec = fn(conn.Conn)
	}
}

func (conn *serverCodecConn) GetServerCodec() rpc.ServerCodec {
	return conn.ServerCodec
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (conn *serverCodecConn) Close() error {
	var err error
	if conn.ServerCodec != nil {
		err = conn.ServerCodec.Close()
	} else {
		err = conn.Conn.Close()
	}
	return err
}

// Context is by name.
type Context struct {
	codecConn ServerCodecConn
	service   *service
	server    *Server
	req       *rpc.Request
	resp      *rpc.Response
	mtype     *methodType
	argv      reflect.Value
	replyv    reflect.Value
	Path      string
	Query     url.Values
	data      map[interface{}]interface{}
	Log       Logger
	sync.RWMutex
}

// RemoteAddr return remote address
func (ctx *Context) RemoteAddr() string {
	addr := ctx.codecConn.RemoteAddr()
	return addr.String()
}

func (ctx *Context) readRequestHeader() (keepReading bool, notSend bool, err error) {
	// set timeout
	if ctx.server.timeout > 0 {
		ctx.codecConn.SetDeadline(time.Now().Add(ctx.server.timeout))
	}
	if ctx.server.readTimeout > 0 {
		ctx.codecConn.SetReadDeadline(time.Now().Add(ctx.server.readTimeout))
	}

	// pre
	err = ctx.server.PluginContainer.doPreReadRequestHeader(ctx)
	if err != nil {
		keepReading = true // Added laster by henry
		return
	}

	// decode request header
	err = ctx.codecConn.ReadRequestHeader(ctx.req)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			notSend = true
			return
		}
		keepReading = true // Added laster by henry
		err = NewRPCError("ReadRequestHeader: ", err.Error())
		return
	}

	// We read the header successfully. If we see an error now,
	// we can still recover and move on to the next request.
	keepReading = true

	var serviceName, methodName string
	serviceName, methodName, err = ctx.parseServiceMethod()
	if err != nil {
		err = NewRPCError("parse ServiceMethod: ", err.Error())
		return
	}

	// Look up the request.
	ctx.server.mu.RLock()
	ctx.service = ctx.server.serviceMap[serviceName]
	ctx.server.mu.RUnlock()
	if ctx.service == nil {
		err = NewRPCError("can't find service " + ctx.req.ServiceMethod)
		return
	}
	ctx.mtype = ctx.service.method[methodName]
	if ctx.mtype == nil {
		err = NewRPCError("can't find method " + ctx.req.ServiceMethod)
	}

	// post
	err = ctx.service.pluginContainer.doPostReadRequestHeader(ctx)
	if err != nil {
		return
	}
	err = ctx.server.PluginContainer.doPostReadRequestHeader(ctx)
	return
}

func (ctx *Context) readRequestBody(body interface{}) error {
	var err error
	// pre
	err = ctx.server.PluginContainer.doPreReadRequestBody(ctx, body)
	if err != nil {
		return err
	}
	if ctx.service != nil {
		err = ctx.service.pluginContainer.doPreReadRequestBody(ctx, body)
		if err != nil {
			return err
		}
	}

	err = ctx.codecConn.ReadRequestBody(body)
	if err != nil {
		return NewRPCError("ReadRequestBody: ", err.Error())
	}

	// post
	if ctx.service != nil {
		err = ctx.service.pluginContainer.doPostReadRequestBody(ctx, body)
		if err != nil {
			return err
		}
	}
	err = ctx.server.PluginContainer.doPostReadRequestBody(ctx, body)
	return err
}

// writeResponse must be safe for concurrent use by multiple goroutines.
func (ctx *Context) writeResponse(body interface{}) error {
	// set timeout
	if ctx.server.timeout > 0 {
		ctx.codecConn.SetDeadline(time.Now().Add(ctx.server.timeout))
	}
	if ctx.server.writeTimeout > 0 {
		ctx.codecConn.SetWriteDeadline(time.Now().Add(ctx.server.writeTimeout))
	}

	var err error
	// pre
	err = ctx.server.PluginContainer.doPreWriteResponse(ctx, body)
	if err != nil {
		return err
	}
	if ctx.service != nil {
		err = ctx.service.pluginContainer.doPreWriteResponse(ctx, body)
		if err != nil {
			return err
		}
	}

	// decode request header
	err = ctx.codecConn.WriteResponse(ctx.resp, body)
	if err != nil {
		return NewRPCError("WriteResponse: ", err.Error())
	}

	// post
	if ctx.service != nil {
		err = ctx.service.pluginContainer.doPostWriteResponse(ctx, body)
		if err != nil {
			return err
		}
	}
	err = ctx.server.PluginContainer.doPostWriteResponse(ctx, body)
	return err
}

func (ctx *Context) parseServiceMethod() (service string, method string, err error) {
	boundary := strings.Index(ctx.req.ServiceMethod, "?")
	if boundary < 0 {
		ctx.Path = ctx.req.ServiceMethod
	} else {
		ctx.Path = ctx.req.ServiceMethod[:boundary]
		queryString := ctx.req.ServiceMethod[boundary+1:]
		ctx.Query, err = url.ParseQuery(queryString)
		if err != nil {
			err = ErrServiceMethod.Format(ctx.req.ServiceMethod)
			return
		}
	}
	return ctx.splitServiceMethod()
}

func (ctx *Context) splitServiceMethod() (service string, method string, err error) {
	dot := strings.LastIndex(ctx.Path, ".")
	if dot <= 0 || dot+1 == len(ctx.Path) {
		err = ErrServiceMethod.Format(ctx.req.ServiceMethod)
		return
	}
	service = ctx.Path[:dot]
	method = ctx.Path[dot+1:]
	return
}
