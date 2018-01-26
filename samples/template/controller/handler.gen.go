package controller

import (
	"github.com/henrylee2cn/ant/samples/template/model"
	"github.com/henrylee2cn/ant/samples/template/types"
	tp "github.com/henrylee2cn/teleport"
)

// Math mathematical calculation controller
type Math struct {
	tp.PullCtx
}

// Divide division
func (m *Math) Divide(args *types.MathDivideArgs) (*types.MathDivideReply, *tp.Rerror) {
	return model.MathDivide(args)
}
