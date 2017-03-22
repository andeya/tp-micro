package codec

import (
	"bytes"
	"net/rpc"
	"reflect"
	"testing"

	"github.com/mars9/codec/internal"
)

type buffer struct {
	*bytes.Buffer
}

func (c buffer) Close() error { return nil }

func TestClientCodecBasic(t *testing.T) {
	t.Parallel()

	buf := buffer{Buffer: &bytes.Buffer{}}
	cc := NewClientCodec(buf)
	sc := NewServerCodec(buf)

	req := rpc.Request{
		ServiceMethod: "test.service.method",
		Seq:           1<<64 - 1,
	}
	resp := rpc.Request{}
	body := internal.Struct{}

	if err := cc.WriteRequest(&req, &testMessage); err != nil {
		t.Fatalf("write client request: %v", err)
	}

	if err := sc.ReadRequestHeader(&resp); err != nil {
		t.Fatalf("read client request header: %v", err)
	}
	if err := sc.ReadRequestBody(&body); err != nil {
		t.Fatalf("read client request body: %v", err)
	}

	if !reflect.DeepEqual(req, resp) {
		t.Fatalf("encode/decode requeset header: expected %#v, got %#v", req, resp)
	}
	if !reflect.DeepEqual(testMessage, body) {
		t.Fatalf("encode/decode request body: expected %#v, got %#v", testMessage, body)
	}
}

func TestServerCodecBasic(t *testing.T) {
	t.Parallel()

	buf := buffer{Buffer: &bytes.Buffer{}}
	cc := NewClientCodec(buf)
	sc := NewServerCodec(buf)

	req := rpc.Response{
		ServiceMethod: "test.service.method",
		Seq:           1<<64 - 1,
		Error:         "test error message",
	}
	resp := rpc.Response{}
	body := internal.Struct{}

	if err := sc.WriteResponse(&req, &testMessage); err != nil {
		t.Fatalf("write server response: %v", err)
	}

	if err := cc.ReadResponseHeader(&resp); err != nil {
		t.Fatalf("read server response header: %v", err)
	}
	if err := cc.ReadResponseBody(&body); err != nil {
		t.Fatalf("read server request body: %v", err)
	}

	if !reflect.DeepEqual(req, resp) {
		t.Fatalf("encode/decode response header: expected %#v, got %#v", req, resp)
	}
	if !reflect.DeepEqual(testMessage, body) {
		t.Fatalf("encode/decode response body: expected %#v, got %#v", testMessage, body)
	}
}
