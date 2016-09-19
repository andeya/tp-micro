package rpc2

import (
	"testing"
)

func TestString(t *testing.T) {
	var auth = &AuthorizationPlugin{
		Token: "adflj记录；；接啊地方&ljf东方巨龙== 啊两地分居",
		Tag:   "ljqr3456l&&asdlj就啦看电视感觉%=",
	}
	t.Logf("%#v\n", auth.String())
}
