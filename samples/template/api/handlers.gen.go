package api

import (
	tp "github.com/henrylee2cn/teleport"
	"github.com/henrylee2cn/tp-micro/samples/template/logic"
	"github.com/henrylee2cn/tp-micro/samples/template/types"
)

// Math mathematical calculation controller
type Math struct {
	tp.PullCtx
}

// Divide division
func (m *Math) Divide(args *types.MathDivideArgs) (*types.MathDivideReply, *tp.Rerror) {
	return logic.MathDivide(args)
}
