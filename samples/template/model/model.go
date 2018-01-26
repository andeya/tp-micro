package model

import (
	"github.com/henrylee2cn/ant/samples/template/types"
	tp "github.com/henrylee2cn/teleport"
)

// MathDivide division.
func MathDivide(args *types.MathDivideArgs) (*types.MathDivideReply, *tp.Rerror) {
	r := args.A / args.B
	return &r, nil
}
