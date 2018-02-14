
package api

import (
    "logic"
    "types"
    tp "github.com/henrylee2cn/teleport"
)

type ApiHandlers struct {
    tp.PullCtx
}

func (handlers *ApiHandlers) getUserList(args *RpcArgs) (*RpcReply, *tp.Rerror) {
    return logic.getUserList(args)
}

func (handlers *ApiHandlers) getUserList2(a int, b string, args *RpcArgs, args2 *RpcArgs) (*RpcReply, *tp.Rerror) {
    return logic.getUserList2(a, b, args, args2)
}



