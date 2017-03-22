package codec

import (
	"bufio"
	"fmt"
	"io"
	"net/rpc"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/mars9/codec/wirepb"
)

const defaultBufferSize = 4 * 1024

type serverCodec struct {
	mu   sync.Mutex // exclusive writer lock
	resp wirepb.ResponseHeader
	enc  *Encoder
	w    *bufio.Writer

	req wirepb.RequestHeader
	dec *Decoder
	c   io.Closer
}

// NewServerCodec returns a new rpc.ServerCodec.
//
// A ServerCodec implements reading of RPC requests and writing of RPC
// responses for the server side of an RPC session. The server calls
// ReadRequestHeader and ReadRequestBody in pairs to read requests from the
// connection, and it calls WriteResponse to write a response back. The
// server calls Close when finished with the connection. ReadRequestBody
// may be called with a nil argument to force the body of the request to be
// read and discarded.
func NewServerCodec(rwc io.ReadWriteCloser) rpc.ServerCodec {
	w := bufio.NewWriterSize(rwc, defaultBufferSize)
	r := bufio.NewReaderSize(rwc, defaultBufferSize)
	return &serverCodec{
		enc: NewEncoder(w),
		w:   w,
		dec: NewDecoder(r),
		c:   rwc,
	}
}

func (c *serverCodec) WriteResponse(resp *rpc.Response, body interface{}) error {
	c.mu.Lock()
	c.resp.Method = resp.ServiceMethod
	c.resp.Seq = resp.Seq
	c.resp.Error = resp.Error

	err := encode(c.enc, &c.resp)
	if err != nil {
		c.mu.Unlock()
		return err
	}
	if err = encode(c.enc, body); err != nil {
		c.mu.Unlock()
		return err
	}
	err = c.w.Flush()
	c.mu.Unlock()
	return err
}

func (c *serverCodec) ReadRequestHeader(req *rpc.Request) error {
	c.req.Reset()
	if err := c.dec.Decode(&c.req); err != nil {
		return err
	}

	req.ServiceMethod = c.req.Method
	req.Seq = c.req.Seq
	return nil
}

func (c *serverCodec) ReadRequestBody(body interface{}) error {
	if pb, ok := body.(proto.Message); ok {
		return c.dec.Decode(pb)
	}
	return fmt.Errorf("%T does not implement proto.Message", body)
}

func (c *serverCodec) Close() error { return c.c.Close() }
