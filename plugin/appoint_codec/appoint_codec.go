package appoint_codec

import (
	"github.com/henrylee2cn/rpc2/client"
	"github.com/henrylee2cn/rpc2/codec/bson"
	"github.com/henrylee2cn/rpc2/codec/colfer"
	"github.com/henrylee2cn/rpc2/codec/gencode"
	"github.com/henrylee2cn/rpc2/codec/gob"
	"github.com/henrylee2cn/rpc2/codec/jsonrpc"
	"github.com/henrylee2cn/rpc2/codec/protobuf"
	"github.com/henrylee2cn/rpc2/plugin"
	"github.com/henrylee2cn/rpc2/server"
)

type CodecType byte

const (
	CODEC_TYPE_BSON CodecType = iota + 1
	CODEC_TYPE_COLFER
	CODEC_TYPE_GENCODE
	CODEC_TYPE_GOB
	CODEC_TYPE_JSON
	CODEC_TYPE_PROTOBUF
)

// AppointCodecPlugin appoints data exchange format.
type AppointCodecPlugin struct {
	clientCodecType []byte
	clientCodecFunc client.ClientCodecFunc
}

func NewClientAppointCodecPlugin(codecType CodecType) *AppointCodecPlugin {
	var codecFunc client.ClientCodecFunc
	switch codecType {
	case CODEC_TYPE_BSON:
		codecFunc = bson.NewBsonClientCodec
	case CODEC_TYPE_COLFER:
		codecFunc = colfer.NewClientCodec
	case CODEC_TYPE_GENCODE:
		codecFunc = gencode.NewGencodeClientCodec
	case CODEC_TYPE_GOB:
		codecFunc = gob.NewGobClientCodec
	case CODEC_TYPE_JSON:
		codecFunc = jsonrpc.NewJSONRPCClientCodec
	case CODEC_TYPE_PROTOBUF:
		codecFunc = protobuf.NewProtobufClientCodec
	}
	return &AppointCodecPlugin{
		clientCodecType: []byte{byte(codecType)},
		clientCodecFunc: codecFunc,
	}
}

func NewServerAppointCodecPlugin() *AppointCodecPlugin {
	return &AppointCodecPlugin{}
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
	switch CodecType(b[0]) {
	case CODEC_TYPE_BSON:
		codecConn.SetServerCodec(bson.NewBsonServerCodec)
	case CODEC_TYPE_COLFER:
		codecConn.SetServerCodec(colfer.NewServerCodec)
	case CODEC_TYPE_GENCODE:
		codecConn.SetServerCodec(gencode.NewGencodeServerCodec)
	case CODEC_TYPE_GOB:
		codecConn.SetServerCodec(gob.NewGobServerCodec)
	case CODEC_TYPE_JSON:
		codecConn.SetServerCodec(jsonrpc.NewJSONRPCServerCodec)
	case CODEC_TYPE_PROTOBUF:
		codecConn.SetServerCodec(protobuf.NewProtobufServerCodec)
	}
	return nil
}
