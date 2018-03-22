package sdk

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/teleport/socket"
	micro "github.com/henrylee2cn/tp-micro"
	"github.com/henrylee2cn/tp-micro/samples/template/types"
)

var client *micro.Client

// Init init client with config and linker.
func Init(cliConfig micro.CliConfig, linker micro.Linker) {
	client = micro.NewClient(
		cliConfig,
		linker,
	)
}

// InitWithClient init client with current client.
func InitWithClient(cli *micro.Client) {
	client = cli
}

// PullMathDivide division form mathematical calculation controller.
func PullMathDivide(args *types.MathDivideArgs, setting ...socket.PacketSetting) (*types.MathDivideReply, *tp.Rerror) {
	reply := new(types.MathDivideReply)
	rerr := client.Pull("/root/math/divide", args, reply, setting...).Rerror()
	return reply, rerr
}
