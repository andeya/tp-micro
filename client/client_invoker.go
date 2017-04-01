package client

import (
	"errors"
	"io"
	"net/rpc"
	"strconv"
	"sync"

	"github.com/henrylee2cn/rpc2/common"
	"github.com/henrylee2cn/rpc2/log"
)

var _ Invoker = new(invoker)

type (
	// Invoker provides remote call function.
	Invoker interface {
		Call(serviceMethod string, args interface{}, reply interface{}) *common.RPCError
		Go(serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call
		Close() error
	}

	// Client represents an RPC Client.
	// There may be multiple outstanding Calls associated
	// with a single Client, and a Client may be used by
	// multiple goroutines simultaneously.
	invoker struct {
		codec *clientCodecWrapper

		reqMutex sync.Mutex // protects following
		request  rpc.Request

		mutex    sync.Mutex // protects following
		seq      uint64
		pending  map[uint64]*Call
		closing  bool // user has called Close
		shutdown bool // server has told us to stop
	}

	// Call represents an active RPC.
	Call struct {
		ServiceMethod string           // The name of the service and method to call.
		Args          interface{}      // The argument to the function (*struct).
		Reply         interface{}      // The reply from the function (*struct).
		Error         *common.RPCError // After completion, the error status.
		Done          chan *Call       // Strobes when call is complete.
	}
)

// newInvoker is like NewClientWithConn but uses the specified
// codec to encode requests and decode responses.
func newInvoker(codec *clientCodecWrapper) Invoker {
	invoker := &invoker{
		codec:   codec,
		pending: make(map[uint64]*Call),
	}
	go invoker.input()
	return invoker
}

// Go invokes the function asynchronously. It returns the Call structure representing
// the invocation. The done channel will signal when the call is complete by returning
// the same Call object. If done is nil, Go will allocate a new channel.
// If non-nil, done must be buffered or Go will deliberately crash.
func (invoker *invoker) Go(serviceMethod string, args interface{}, reply interface{}, done chan *Call) *Call {
	call := new(Call)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Reply = reply
	if done == nil {
		done = make(chan *Call, 10) // buffered.
	} else {
		// If caller passes done != nil, it must arrange that
		// done has enough buffer for the number of simultaneous
		// RPCs that will be using that channel. If the channel
		// is totally unbuffered, it's best not to run at all.
		if cap(done) == 0 {
			log.Panic("rpc: done channel is unbuffered")
		}
	}
	call.Done = done
	invoker.send(call)
	return call
}

// Call invokes the named function, waits for it to complete, and returns its error status.
func (invoker *invoker) Call(serviceMethod string, args interface{}, reply interface{}) *common.RPCError {
	call := <-invoker.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}

// Close calls the underlying codec's Close method. If the connection is already
// shutting down, RPCErrShutdown is returned.
func (invoker *invoker) Close() error {
	invoker.mutex.Lock()
	if invoker.closing {
		invoker.mutex.Unlock()
		return errors.New(common.RPCErrShutdown.Error)
	}
	invoker.closing = true
	invoker.mutex.Unlock()
	return invoker.codec.Close()
}

func (invoker *invoker) send(call *Call) {
	invoker.reqMutex.Lock()
	defer invoker.reqMutex.Unlock()

	// Register this call.
	invoker.mutex.Lock()
	if invoker.shutdown || invoker.closing {
		call.Error = common.RPCErrShutdown
		invoker.mutex.Unlock()
		call.done()
		return
	}
	seq := invoker.seq
	invoker.seq++
	invoker.pending[seq] = call
	invoker.mutex.Unlock()

	// Encode and send the request.
	invoker.request.Seq = seq
	invoker.request.ServiceMethod = call.ServiceMethod
	rpcErr := invoker.codec.WriteRequest(&invoker.request, call.Args)
	if rpcErr != nil {
		invoker.mutex.Lock()
		call = invoker.pending[seq]
		delete(invoker.pending, seq)
		invoker.mutex.Unlock()
		if call != nil {
			call.Error = rpcErr
			call.done()
		}
	}
}

func (invoker *invoker) input() {
	var (
		rpcErr   *common.RPCError
		response rpc.Response
	)
	for rpcErr == nil {
		response = rpc.Response{}
		rpcErr = invoker.codec.ReadResponseHeader(&response)
		if rpcErr != nil {
			break
		}
		seq := response.Seq
		invoker.mutex.Lock()
		call := invoker.pending[seq]
		delete(invoker.pending, seq)
		invoker.mutex.Unlock()

		switch {
		case call == nil:
			// We've got no pending call. That usually means that
			// WriteRequest partially failed, and call was already
			// removed; response is a server telling us about an
			// error reading request body. We should still attempt
			// to read error body, but there's no one to give it to.
			rpcErr = invoker.codec.ReadResponseBody(nil)

		case response.Error != "":
			// We've got an error response. Give this to the request;
			// any subsequent requests will get the ReadResponseBody
			// error if there is one.
			rpcErr = parseResponseError(response.Error)
			call.Error = rpcErr
			rpcErr = invoker.codec.ReadResponseBody(nil)
			call.done()

		default:
			rpcErr = invoker.codec.ReadResponseBody(call.Reply)
			if rpcErr != nil {
				call.Error = rpcErr
			}
			call.done()
		}
	}
	// Terminate pending calls.
	invoker.reqMutex.Lock()
	invoker.mutex.Lock()
	invoker.shutdown = true
	closing := invoker.closing
	if rpcErr != nil && rpcErr.Error == io.EOF.Error() {
		if closing {
			rpcErr = common.RPCErrShutdown
		} else {
			rpcErr.Error = io.ErrUnexpectedEOF.Error()
		}
	} else if !closing {
		log.Debug("rpc: invoker protocol error: " + rpcErr.Error)
	}
	for _, call := range invoker.pending {
		call.Error = rpcErr
		call.done()
	}
	invoker.mutex.Unlock()
	invoker.reqMutex.Unlock()
}

func (call *Call) done() {
	select {
	case call.Done <- call:
		// ok
	default:
		// We don't want to block here. It is the caller's responsibility to make
		// sure the channel has enough buffer space. See comment in Go().
		log.Debug("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

func parseResponseError(errMsg string) *common.RPCError {
	i, _ := strconv.Atoi(errMsg[:1])
	return &common.RPCError{
		Type:  common.ErrorType(i),
		Error: errMsg[1:],
	}
}
