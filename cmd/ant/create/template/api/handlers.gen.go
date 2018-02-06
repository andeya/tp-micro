package api

import (
	"{{PROJ_PATH}}/logic"
	"{{PROJ_PATH}}/types"
	tp "github.com/henrylee2cn/teleport"
)

// Math mathematical calculation controller
type Math struct {
	tp.PullCtx
}

// Divide division
func (m *Math) Divide(args *types.MathDivideArgs) (*types.MathDivideReply, *tp.Rerror) {
	return logic.MathDivide(args)
}
