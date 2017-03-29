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

func (w *clientCodecWrapper) WriteRequest(r *rpc.Request, body interface{}) error {
	if w.timeout > 0 {
		w.codecConn.SetDeadline(time.Now().Add(w.timeout))
	}
	if w.writeTimeout > 0 {
		w.codecConn.SetWriteDeadline(time.Now().Add(w.writeTimeout))
	}

	//pre
	err := w.pluginContainer.doPreWriteRequest(r, body)
	if err != nil {
		return err
	}

	err = w.codecConn.WriteRequest(r, body)
	if err != nil {
		return common.NewRPCError("WriteRequest: " + err.Error())
	}

	//post
	err = w.pluginContainer.doPostWriteRequest(r, body)
	return err
}

func (w *clientCodecWrapper) ReadResponseHeader(r *rpc.Response) error {
	if w.timeout > 0 {
		w.codecConn.SetDeadline(time.Now().Add(w.timeout))
	}
	if w.readTimeout > 0 {
		w.codecConn.SetReadDeadline(time.Now().Add(w.readTimeout))
	}

	//pre
	err := w.pluginContainer.doPreReadResponseHeader(r)
	if err != nil {
		return err
	}

	err = w.codecConn.ReadResponseHeader(r)
	if err != nil {
		return common.NewRPCError("ReadResponseHeader: " + err.Error())
	}

	//post
	err = w.pluginContainer.doPostReadResponseHeader(r)
	return err
}

func (w *clientCodecWrapper) ReadResponseBody(body interface{}) error {
	//pre
	err := w.pluginContainer.doPreReadResponseBody(body)
	if err != nil {
		return err
	}

	err = w.codecConn.ReadResponseBody(body)
	if err != nil {
		return common.NewRPCError("ReadResponseBody: " + err.Error())
	}

	//post
	err = w.pluginContainer.doPostReadResponseBody(body)
	return err
}

func (w *clientCodecWrapper) Close() error {
	return w.codecConn.Close()
}
