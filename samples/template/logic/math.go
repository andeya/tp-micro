package logic

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/tp-micro/samples/template/types"
)

// MathDivide division.
func MathDivide(args *types.MathDivideArgs) (*types.MathDivideReply, *tp.Rerror) {
	r := args.A / args.B
	return &r, nil
}
