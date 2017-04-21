package protobuf

import (
	"io"
	"net/rpc"

	codec "github.com/henrylee2cn/codec_protobuf"
)

// NewProtobufServerCodec creates a protobuf ServerCodec by https://github.com/henrylee2cn/codec_protobuf
func NewProtobufServerCodec(conn io.ReadWriteCloser) rpc.ServerCodec {
	return codec.NewServerCodec(conn)
}

// NewProtobufClientCodec creates a protobuf ClientCodec by https://github.com/henrylee2cn/codec_protobuf
func NewProtobufClientCodec(conn io.ReadWriteCloser) rpc.ClientCodec {
	return codec.NewClientCodec(conn)
}
