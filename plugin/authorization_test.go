package rpc2

import (
	"testing"
)

func TestClientAuthorization(t *testing.T) {
	var auth = NewClientAuthorization(
		"adflj记录；；接啊地方&ljf东方巨龙== 啊两地分居",
		"ljqr3456l&&asdlj就啦看电视感觉%=",
	)

	t.Logf("String(): %v\n", auth)
}
