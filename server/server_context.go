package server

import (
	"io"
	"net/rpc"
	"net/url"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/henrylee2cn/rpc2/common"
)

// Context means as its name.
type Context struct {
	codecConn ServerCodecConn
	server    *Server
	req       *rpc.Request
	resp      *rpc.Response
	service   IService
	argv      reflect.Value
	replyv    reflect.Value
	path      string
	query     url.Values
	data      map[interface{}]interface{}
	sync.RWMutex
}

// RemoteAddr returns remote address
func (ctx *Context) RemoteAddr() string {
	addr := ctx.codecConn.RemoteAddr()
	return addr.String()
}

// Seq returns request sequence number chosen by client.
func (ctx *Context) Seq() uint64 {
	return ctx.req.Seq
}

// ID returns request unique identifier.
// Node: Called before 'ReadRequestHeader' is invalid!
func (ctx *Context) ID() string {
	return ctx.RemoteAddr() + "-" + strconv.FormatUint(ctx.req.Seq, 10)
}

// ServiceMethod returns request raw serviceMethod.
func (ctx *Context) ServiceMethod() string {
	// return ctx.req.ServiceMethod
	return ctx.server.ServiceBuilder.URIEncode(ctx.query, ctx.path)
}

// Path returns request serviceMethod path.
func (ctx *Context) Path() string {
	return ctx.path
}

// SetPath sets request serviceMethod path.
func (ctx *Context) SetPath(p string) {
	ctx.path = p
}

// Query returns request query params.
func (ctx *Context) Query() url.Values {
	return ctx.query
}

// Data returns the stored data in this context.
func (ctx *Context) Data(key interface{}) interface{} {
	if v, ok := ctx.data[key]; ok {
		return v
	}
	return nil
}

// HasData checks if the key exists in the context.
func (ctx *Context) HasData(key interface{}) bool {
	_, ok := ctx.data[key]
	return ok
}

// DataAll return the implicit data in the context
func (ctx *Context) DataAll() map[interface{}]interface{} {
	if ctx.data == nil {
		ctx.data = make(map[interface{}]interface{})
	}
	return ctx.data
}

// SetData stores data with given key in this context.
// This data are only available in this context.
func (ctx *Context) SetData(key, val interface{}) {
	if ctx.data == nil {
		ctx.data = make(map[interface{}]interface{})
	}
	ctx.data[key] = val
}

func (ctx *Context) readRequestHeader() (keepReading bool, notSend bool, err error) {
	// set timeout
	if ctx.server.Timeout > 0 {
		ctx.codecConn.SetDeadline(time.Now().Add(ctx.server.Timeout))
	}
	if ctx.server.ReadTimeout > 0 {
		ctx.codecConn.SetReadDeadline(time.Now().Add(ctx.server.ReadTimeout))
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
		err = common.NewRPCError("ReadRequestHeader: " + err.Error())
		return
	}

	// We read the header successfully. If we see an error now,
	// we can still recover and move on to the next request.
	keepReading = true

	// parse serviceMethod
	ctx.path, ctx.query, err = ctx.server.ServiceBuilder.URIParse(ctx.req.ServiceMethod)
	if err != nil {
		err = common.NewRPCError(err.Error())
		return
	}

	// post
	err = ctx.server.PluginContainer.doPostReadRequestHeader(ctx)
	if err != nil {
		return
	}

	// get service
	ctx.server.mu.RLock()
	ctx.service = ctx.server.serviceMap[ctx.path]
	ctx.server.mu.RUnlock()
	if ctx.service == nil {
		err = common.NewRPCError("can't find service '" + ctx.path + "'")
	}

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
		err = ctx.service.GetPluginContainer().doPreReadRequestBody(ctx, body)
		if err != nil {
			return err
		}
	}

	err = ctx.codecConn.ReadRequestBody(body)
	if err != nil {
		return common.NewRPCError("ReadRequestBody: " + err.Error())
	}

	// post
	if ctx.service != nil {
		err = ctx.service.GetPluginContainer().doPostReadRequestBody(ctx, body)
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
	if ctx.server.Timeout > 0 {
		ctx.codecConn.SetDeadline(time.Now().Add(ctx.server.Timeout))
	}
	if ctx.server.WriteTimeout > 0 {
		ctx.codecConn.SetWriteDeadline(time.Now().Add(ctx.server.WriteTimeout))
	}

	var err error
	// pre
	err = ctx.server.PluginContainer.doPreWriteResponse(ctx, body)
	if err != nil {
		return err
	}
	if ctx.service != nil {
		err = ctx.service.GetPluginContainer().doPreWriteResponse(ctx, body)
		if err != nil {
			return err
		}
	}

	// decode request header
	err = ctx.codecConn.WriteResponse(ctx.resp, body)
	if err != nil {
		return common.NewRPCError("WriteResponse: " + err.Error())
	}

	// post
	if ctx.service != nil {
		err = ctx.service.GetPluginContainer().doPostWriteResponse(ctx, body)
		if err != nil {
			return err
		}
	}
	err = ctx.server.PluginContainer.doPostWriteResponse(ctx, body)
	return err
}
