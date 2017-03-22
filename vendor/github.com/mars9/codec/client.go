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

type clientCodec struct {
	mu  sync.Mutex // exclusive writer lock
	req wirepb.RequestHeader
	enc *Encoder
	w   *bufio.Writer

	resp wirepb.ResponseHeader
	dec  *Decoder
	c    io.Closer
}

// NewClientCodec returns a new rpc.Client.
//
// A ClientCodec implements writing of RPC requests and reading of RPC
// responses for the client side of an RPC session. The client calls
// WriteRequest to write a request to the connection and calls
// ReadResponseHeader and ReadResponseBody in pairs to read responses. The
// client calls Close when finished with the connection. ReadResponseBody
// may be called with a nil argument to force the body of the response to
// be read and then discarded.
func NewClientCodec(rwc io.ReadWriteCloser) rpc.ClientCodec {
	w := bufio.NewWriterSize(rwc, defaultBufferSize)
	r := bufio.NewReaderSize(rwc, defaultBufferSize)
	return &clientCodec{
		enc: NewEncoder(w),
		w:   w,
		dec: NewDecoder(r),
		c:   rwc,
	}
}

func (c *clientCodec) WriteRequest(req *rpc.Request, body interface{}) error {
	c.mu.Lock()
	c.req.Method = req.ServiceMethod
	c.req.Seq = req.Seq

	err := encode(c.enc, &c.req)
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

func (c *clientCodec) ReadResponseHeader(resp *rpc.Response) error {
	c.resp.Reset()
	if err := c.dec.Decode(&c.resp); err != nil {
		return err
	}

	resp.ServiceMethod = c.resp.Method
	resp.Seq = c.resp.Seq
	resp.Error = c.resp.Error
	return nil
}

func (c *clientCodec) ReadResponseBody(body interface{}) (err error) {
	if pb, ok := body.(proto.Message); ok {
		return c.dec.Decode(pb)
	}
	return fmt.Errorf("%T does not implement proto.Message", body)
}

func encode(enc *Encoder, m interface{}) (err error) {
	if pb, ok := m.(proto.Message); ok {
		return enc.Encode(pb)
	}
	return fmt.Errorf("%T does not implement proto.Message", m)
}

func (c *clientCodec) Close() error { return c.c.Close() }
