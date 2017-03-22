package gencodec

import (
	"encoding/binary"
	"fmt"
	"io"
)

const defaultBufferSize = 1024

// DecodeReader manages the receipt of type and data information read from the
// remote side of a connection.
type DecodeReader interface {
	io.ByteReader
	io.Reader
}

type gencodeDecoder struct {
	r   DecodeReader
	buf []byte
}

func newGencodeDecoder(r DecodeReader) *gencodeDecoder {
	return &gencodeDecoder{r: r}
}

func (d *gencodeDecoder) Decode(pv genCodeMessage) (err error) {
	if pv == nil {
		return nil
	}

	buf := make([]byte, defaultBufferSize)
	buf, err = readFull(d.r, buf)

	if err != nil {
		return err
	}
	_, err = pv.Unmarshal(buf)
	return
}

func readFull(r DecodeReader, buf []byte) ([]byte, error) {
	val, err := binary.ReadUvarint(r.(io.ByteReader))
	if err != nil {
		return buf[:0], err
	}
	size := int(val)

	if cap(buf) < size {
		buf = make([]byte, size)
	}
	buf = buf[:size]

	_, err = io.ReadFull(r, buf)
	return buf, err
}

type gencodeEncoder struct {
	size [binary.MaxVarintLen64]byte
	w    io.Writer
}

func newGencodeEncoder(w io.Writer) *gencodeEncoder {
	return &gencodeEncoder{
		w: w,
	}
}

func (e *gencodeEncoder) Encode(body interface{}) (err error) {
	msg, ok := body.(genCodeMessage)
	if !ok {
		return fmt.Errorf("%T does not implement genCodeMessage", body)
	}

	buf := make([]byte, msg.Size())
	buf, err = msg.Marshal(buf)

	e.writeFrame(buf)
	return
}

func (e *gencodeEncoder) writeFrame(data []byte) (err error) {
	n := binary.PutUvarint(e.size[:], uint64(len(data)))
	if _, err = e.w.Write(e.size[:n]); err != nil {
		return err
	}
	_, err = e.w.Write(data)
	return err
}

type genCodeMessage interface {
	Size() uint64
	Marshal(buf []byte) ([]byte, error)
	Unmarshal(buf []byte) (uint64, error)
}
