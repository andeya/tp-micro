package logic

import (
	"{{PROJ_PATH}}/types"
	tp "github.com/henrylee2cn/teleport"
)

// MathDivide division.
func MathDivide(args *types.MathDivideArgs) (*types.MathDivideReply, *tp.Rerror) {
	r := args.A / args.B
	return &r, nil
}
