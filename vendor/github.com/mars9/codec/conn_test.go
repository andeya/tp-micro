package codec

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"io"
	"reflect"
	"testing"

	"github.com/mars9/codec/internal"
)

var (
	testMessage internal.Struct
	testData    [defaultBufferSize + 8]byte
)

func init() {
	if _, err := io.ReadFull(rand.Reader, testData[:]); err != nil {
		panic("not enough entropy")
	}
	testMessage.Method = "test.service.method"
	testMessage.Seq = 1<<64 - 1
	testMessage.Bucket = make([][]byte, 16)
	for i := 0; i < 16; i++ {
		testMessage.Bucket[i] = testData[:]
	}
	testMessage.Data = testData[:]
}

func TestWriteReadFrame(t *testing.T) {
	t.Parallel()

	buf := &bytes.Buffer{}
	enc := NewEncoder(buf)

	if err := enc.writeFrame(testData[:]); err != nil {
		t.Fatalf("write frame: %v", err)
	}

	data, err := readFull(bufio.NewReader(buf), nil)
	if err != nil {
		t.Fatalf("read frame: %v", err)
	}

	if bytes.Compare(testData[:], data) != 0 {
		t.Fatalf("expected frame %q, got %q", testData[:], data)
	}
}

func TestEncodeDecode(t *testing.T) {
	t.Parallel()

	req := testMessage
	resp := internal.Struct{}

	buf := &bytes.Buffer{}
	enc := NewEncoder(buf)
	dec := NewDecoder(buf)

	if err := enc.Encode(&req); err != nil {
		t.Fatalf("encode request: %v", err)
	}
	if err := dec.Decode(&resp); err != nil {
		t.Fatalf("decode request: %v", err)
	}

	if !reflect.DeepEqual(req, resp) {
		t.Fatalf("encode/decode: expected %#v, got %#v", req, resp)
	}
}
