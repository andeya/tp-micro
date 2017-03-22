package gencodec

import (
	"io"
	"time"
	"unsafe"
)

var (
	_ = unsafe.Sizeof(0)
	_ = io.ReadFull
	_ = time.Now()
)

type RequestHeader struct {
	ServiceMethod string
	Seq           uint64
}

func (d *RequestHeader) Size() (s uint64) {

	{
		l := uint64(len(d.ServiceMethod))

		{

			t := l
			for t >= 0x80 {
				t <<= 7
				s++
			}
			s++

		}
		s += l
	}
	s += 8
	return
}
func (d *RequestHeader) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{
		l := uint64(len(d.ServiceMethod))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+0] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+0] = byte(t)
			i++

		}
		copy(buf[i+0:], d.ServiceMethod)
		i += l
	}
	{

		buf[i+0+0] = byte(d.Seq >> 0)

		buf[i+1+0] = byte(d.Seq >> 8)

		buf[i+2+0] = byte(d.Seq >> 16)

		buf[i+3+0] = byte(d.Seq >> 24)

		buf[i+4+0] = byte(d.Seq >> 32)

		buf[i+5+0] = byte(d.Seq >> 40)

		buf[i+6+0] = byte(d.Seq >> 48)

		buf[i+7+0] = byte(d.Seq >> 56)

	}
	return buf[:i+8], nil
}

func (d *RequestHeader) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+0] & 0x7F)
			for buf[i+0]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+0]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.ServiceMethod = string(buf[i+0 : i+0+l])
		i += l
	}
	{

		d.Seq = 0 | (uint64(buf[i+0+0]) << 0) | (uint64(buf[i+1+0]) << 8) | (uint64(buf[i+2+0]) << 16) | (uint64(buf[i+3+0]) << 24) | (uint64(buf[i+4+0]) << 32) | (uint64(buf[i+5+0]) << 40) | (uint64(buf[i+6+0]) << 48) | (uint64(buf[i+7+0]) << 56)

	}
	return i + 8, nil
}

type ResponseHeader struct {
	ServiceMethod string
	Seq           uint64
	Error         string
}

func (d *ResponseHeader) Size() (s uint64) {

	{
		l := uint64(len(d.ServiceMethod))

		{

			t := l
			for t >= 0x80 {
				t <<= 7
				s++
			}
			s++

		}
		s += l
	}
	{
		l := uint64(len(d.Error))

		{

			t := l
			for t >= 0x80 {
				t <<= 7
				s++
			}
			s++

		}
		s += l
	}
	s += 8
	return
}
func (d *ResponseHeader) Marshal(buf []byte) ([]byte, error) {
	size := d.Size()
	{
		if uint64(cap(buf)) >= size {
			buf = buf[:size]
		} else {
			buf = make([]byte, size)
		}
	}
	i := uint64(0)

	{
		l := uint64(len(d.ServiceMethod))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+0] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+0] = byte(t)
			i++

		}
		copy(buf[i+0:], d.ServiceMethod)
		i += l
	}
	{

		buf[i+0+0] = byte(d.Seq >> 0)

		buf[i+1+0] = byte(d.Seq >> 8)

		buf[i+2+0] = byte(d.Seq >> 16)

		buf[i+3+0] = byte(d.Seq >> 24)

		buf[i+4+0] = byte(d.Seq >> 32)

		buf[i+5+0] = byte(d.Seq >> 40)

		buf[i+6+0] = byte(d.Seq >> 48)

		buf[i+7+0] = byte(d.Seq >> 56)

	}
	{
		l := uint64(len(d.Error))

		{

			t := uint64(l)

			for t >= 0x80 {
				buf[i+8] = byte(t) | 0x80
				t >>= 7
				i++
			}
			buf[i+8] = byte(t)
			i++

		}
		copy(buf[i+8:], d.Error)
		i += l
	}
	return buf[:i+8], nil
}

func (d *ResponseHeader) Unmarshal(buf []byte) (uint64, error) {
	i := uint64(0)

	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+0] & 0x7F)
			for buf[i+0]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+0]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.ServiceMethod = string(buf[i+0 : i+0+l])
		i += l
	}
	{

		d.Seq = 0 | (uint64(buf[i+0+0]) << 0) | (uint64(buf[i+1+0]) << 8) | (uint64(buf[i+2+0]) << 16) | (uint64(buf[i+3+0]) << 24) | (uint64(buf[i+4+0]) << 32) | (uint64(buf[i+5+0]) << 40) | (uint64(buf[i+6+0]) << 48) | (uint64(buf[i+7+0]) << 56)

	}
	{
		l := uint64(0)

		{

			bs := uint8(7)
			t := uint64(buf[i+8] & 0x7F)
			for buf[i+8]&0x80 == 0x80 {
				i++
				t |= uint64(buf[i+8]&0x7F) << bs
				bs += 7
			}
			i++

			l = t

		}
		d.Error = string(buf[i+8 : i+8+l])
		i += l
	}
	return i + 8, nil
}
