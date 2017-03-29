package appoint_codec

import (
	"github.com/henrylee2cn/rpc2/client"
	"github.com/henrylee2cn/rpc2/codec/jsonrpc"
	"github.com/henrylee2cn/rpc2/codec/protobuf"
	"github.com/henrylee2cn/rpc2/plugin"
	"github.com/henrylee2cn/rpc2/server"
)

type Codec byte

const (
	CODEC_JSON Codec = iota + 1
	CODEC_PROTOBUF
)

// AppointCodecPlugin appoints data exchange format.
type AppointCodecPlugin struct {
	clientCodecType         []byte
	clientCodecFunc         client.ClientCodecFunc
	jsonServerCodecFunc     server.ServerCodecFunc
	protobufServerCodecFunc server.ServerCodecFunc
}

func NewClientAppointCodecPlugin(codec Codec) *AppointCodecPlugin {
	var codecFunc client.ClientCodecFunc
	switch codec {
	case CODEC_JSON:
		codecFunc = jsonrpc.NewJSONRPCClientCodec
	case CODEC_PROTOBUF:
		codecFunc = protobuf.NewProtobufClientCodec
	}
	return &AppointCodecPlugin{
		clientCodecType: []byte{byte(codec)},
		clientCodecFunc: codecFunc,
	}
}

func NewServerAppointCodecPlugin() *AppointCodecPlugin {
	return &AppointCodecPlugin{
		jsonServerCodecFunc:     jsonrpc.NewJSONRPCServerCodec,
		protobufServerCodecFunc: protobuf.NewProtobufServerCodec,
	}
}

var _ plugin.IPlugin = new(AppointCodecPlugin)

// Name returns plugin name.
func (appointCodec *AppointCodecPlugin) Name() string {
	return "AppointCodecPlugin"
}

var _ client.IPostConnectedPlugin = new(AppointCodecPlugin)

func (appointCodec *AppointCodecPlugin) PostConnected(codecConn client.ClientCodecConn) error {
	codecConn.SetClientCodec(appointCodec.clientCodecFunc)
	_, err := codecConn.Write(appointCodec.clientCodecType)
	return err
}

var _ server.IPostConnAcceptPlugin = new(AppointCodecPlugin)

func (appointCodec *AppointCodecPlugin) PostConnAccept(codecConn server.ServerCodecConn) error {
	var b = make([]byte, 1)
	_, err := codecConn.Read(b)
	if err != nil {
		return err
	}
	switch Codec(b[0]) {
	case CODEC_JSON:
		codecConn.SetServerCodec(appointCodec.jsonServerCodecFunc)
	case CODEC_PROTOBUF:
		codecConn.SetServerCodec(appointCodec.protobufServerCodecFunc)
	}
	return nil
}
