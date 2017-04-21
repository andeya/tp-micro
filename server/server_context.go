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
	"github.com/henrylee2cn/rpc2/log"
)

type (
	// Context means as its name.
	Context struct {
		codecConn    ServerCodecConn
		server       *Server
		req          *rpc.Request
		resp         *rpc.Response
		service      IService
		argv         reflect.Value
		replyv       reflect.Value
		path         string
		query        url.Values
		data         *Store
		rpcErrorType common.ErrorType
		sync.RWMutex
	}
	// Store concurrent secure data storage.
	Store struct {
		lock sync.RWMutex
		data map[interface{}]interface{}
	}
)

// Data returns the data store.
// The data are only available in this context.
func (ctx *Context) Data() *Store {
	ctx.RLock()
	if ctx.data == nil {
		ctx.RUnlock()

		ctx.Lock()
		if ctx.data == nil {
			ctx.data = &Store{
				data: make(map[interface{}]interface{}),
			}
		}
		defer ctx.Unlock()

		return ctx.data
	}

	defer ctx.RUnlock()
	return ctx.data
}

// Set stores data with given key in this context.
func (store *Store) Set(key, val interface{}) {
	store.lock.Lock()
	defer store.lock.Unlock()
	store.data[key] = val
}

// Get returns the stored data in this context.
func (store *Store) Get(key interface{}) interface{} {
	store.lock.RLock()
	defer store.lock.RUnlock()
	if v, ok := store.data[key]; ok {
		return v
	}
	return nil
}

// Has checks if the key exists in the context.
func (store *Store) Has(key interface{}) bool {
	store.lock.RLock()
	defer store.lock.RUnlock()
	_, ok := store.data[key]
	return ok
}

// Each performs the callback function on each key.
func (store *Store) Each(callback func(key interface{}, data map[interface{}]interface{}) (next bool)) {
	store.lock.Lock()
	defer store.lock.Unlock()
	d := store.data
	for k := range d {
		if !callback(k, d) {
			return
		}
	}
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
		ctx.rpcErrorType = common.ErrorTypeServerPreReadRequestHeader
		return
	}

	// decode request header
	err = ctx.codecConn.ReadRequestHeader(ctx.req)
	if err != nil {
		ctx.rpcErrorType = common.ErrorTypeServerReadRequestHeader
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			notSend = true
			return
		}
		err = common.NewError("ReadRequestHeader: " + err.Error())
		return
	}

	// We read the header successfully. If we see an error now,
	// we can still recover and move on to the next request.
	keepReading = true

	// parse serviceMethod
	ctx.path, ctx.query, err = ctx.server.ServiceBuilder.URIParse(ctx.req.ServiceMethod)
	if err != nil {
		ctx.rpcErrorType = common.ErrorTypeServerInvalidServiceMethod
		err = common.NewError(err.Error())
		return
	}

	// post
	err = ctx.server.PluginContainer.doPostReadRequestHeader(ctx)
	if err != nil {
		ctx.rpcErrorType = common.ErrorTypeServerPostReadRequestHeader
		return
	}

	// get service
	ctx.server.mu.RLock()
	ctx.service = ctx.server.serviceMap[ctx.path]
	ctx.server.mu.RUnlock()
	if ctx.service == nil {
		ctx.rpcErrorType = common.ErrorTypeServerNotFoundService
		err = common.NewError("can't find service '" + ctx.path + "'")
	}

	return
}

func (ctx *Context) readRequestBody(body interface{}) error {
	var err error
	// pre
	err = ctx.server.PluginContainer.doPreReadRequestBody(ctx, body)
	if err == nil && ctx.service != nil {
		err = ctx.service.GetPluginContainer().doPreReadRequestBody(ctx, body)
	}
	if err != nil {
		ctx.rpcErrorType = common.ErrorTypeServerPreReadRequestBody
		return err
	}

	err = ctx.codecConn.ReadRequestBody(body)
	if err != nil {
		ctx.rpcErrorType = common.ErrorTypeServerReadRequestBody
		return common.NewError("ReadRequestBody: " + err.Error())
	}

	// post
	if ctx.service != nil {
		err = ctx.service.GetPluginContainer().doPostReadRequestBody(ctx, body)
	}
	if err == nil {
		err = ctx.server.PluginContainer.doPostReadRequestBody(ctx, body)
	}
	if err != nil {
		ctx.rpcErrorType = common.ErrorTypeServerPostReadRequestBody
	}
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
	if err == nil && ctx.service != nil {
		err = ctx.service.GetPluginContainer().doPreWriteResponse(ctx, body)
	}
	if err != nil {
		log.Debug("rpc: PreWriteResponse: " + err.Error())
		ctx.rpcErrorType = common.ErrorTypeServerPreWriteResponse
		ctx.resp.Error = err.Error()
		body = nil
	}

	// decode request header
	if len(ctx.resp.Error) > 0 {
		ctx.resp.Error = string(ctx.rpcErrorType) + ctx.resp.Error
	}
	err = ctx.codecConn.WriteResponse(ctx.resp, body)
	if err != nil {
		ctx.rpcErrorType = common.ErrorTypeServerWriteResponse
		ctx.resp.Error = string(ctx.rpcErrorType) + err.Error()
		ctx.codecConn.WriteResponse(ctx.resp, invalidRequest)
		return common.NewError("WriteResponse: " + err.Error())
	}

	// post
	if ctx.service != nil {
		err = ctx.service.GetPluginContainer().doPostWriteResponse(ctx, body)
	}
	if err == nil {
		err = ctx.server.PluginContainer.doPostWriteResponse(ctx, body)
	}
	return err
}
