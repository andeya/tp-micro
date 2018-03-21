package create

import (
	"testing"

	"github.com/henrylee2cn/ant/cmd/ant/create/test"
)

func TestGenerator(t *testing.T) {
	src := test.MustAsset(defAntTpl)
	proj := NewProject(src)
	proj.Prepare()
	proj.genTypesFile()
	proj.genRouterFile()
	proj.genHandlerAndLogicAndSdkFiles()
	t.Logf("types/types.gen.go:\n%s", codeFiles["types/types.gen.go"])
	t.Logf("api/handler.gen.go:\n%s", codeFiles["api/handler.gen.go"])
	t.Logf("api/router.gen.go:\n%s", codeFiles["api/router.gen.go"])
	t.Logf("sdk/rpc.gen.go:\n%s", codeFiles["sdk/rpc.gen.go"])
	t.Logf("sdk/rpc_test.gen.go:\n%s", codeFiles["sdk/rpc_test.gen.go"])
}
