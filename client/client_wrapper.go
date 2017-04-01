package client

import (
	"net/rpc"
	"time"

	"github.com/henrylee2cn/rpc2/common"
)

type clientCodecWrapper struct {
	pluginContainer IClientPluginContainer
	codecConn       ClientCodecConn
	timeout         time.Duration
	readTimeout     time.Duration
	writeTimeout    time.Duration
}

func (w *clientCodecWrapper) WriteRequest(r *rpc.Request, body interface{}) *common.RPCError {
	if w.timeout > 0 {
		w.codecConn.SetDeadline(time.Now().Add(w.timeout))
	}
	if w.writeTimeout > 0 {
		w.codecConn.SetWriteDeadline(time.Now().Add(w.writeTimeout))
	}

	//pre
	err := w.pluginContainer.doPreWriteRequest(r, body)
	if err != nil {
		return &common.RPCError{
			Type:  common.ErrorTypeClientPreWriteRequest,
			Error: err.Error(),
		}
	}

	err = w.codecConn.WriteRequest(r, body)
	if err != nil {
		return &common.RPCError{
			Type:  common.ErrorTypeClientWriteRequest,
			Error: err.Error(),
		}
	}

	//post
	err = w.pluginContainer.doPostWriteRequest(r, body)
	if err != nil {
		return &common.RPCError{
			Type:  common.ErrorTypeClientPostWriteRequest,
			Error: err.Error(),
		}
	}
	return nil
}

func (w *clientCodecWrapper) ReadResponseHeader(r *rpc.Response) *common.RPCError {
	if w.timeout > 0 {
		w.codecConn.SetDeadline(time.Now().Add(w.timeout))
	}
	if w.readTimeout > 0 {
		w.codecConn.SetReadDeadline(time.Now().Add(w.readTimeout))
	}

	//pre
	err := w.pluginContainer.doPreReadResponseHeader(r)
	if err != nil {
		return &common.RPCError{
			Type:  common.ErrorTypeClientPreReadResponseHeader,
			Error: err.Error(),
		}
	}

	err = w.codecConn.ReadResponseHeader(r)
	if err != nil {
		return &common.RPCError{
			Type:  common.ErrorTypeClientReadResponseHeader,
			Error: err.Error(),
		}
	}

	//post
	err = w.pluginContainer.doPostReadResponseHeader(r)
	if err != nil {
		return &common.RPCError{
			Type:  common.ErrorTypeClientPostReadResponseHeader,
			Error: err.Error(),
		}
	}
	return nil
}

func (w *clientCodecWrapper) ReadResponseBody(body interface{}) *common.RPCError {
	//pre
	err := w.pluginContainer.doPreReadResponseBody(body)
	if err != nil {
		return &common.RPCError{
			Type:  common.ErrorTypeClientPreReadResponseBody,
			Error: err.Error(),
		}
	}

	err = w.codecConn.ReadResponseBody(body)
	if err != nil {
		return &common.RPCError{
			Type:  common.ErrorTypeClientReadResponseBody,
			Error: err.Error(),
		}
	}

	//post
	err = w.pluginContainer.doPostReadResponseBody(body)
	if err != nil {
		return &common.RPCError{
			Type:  common.ErrorTypeClientPostReadResponseBody,
			Error: err.Error(),
		}
	}
	return nil
}

func (w *clientCodecWrapper) Close() error {
	return w.codecConn.Close()
}
