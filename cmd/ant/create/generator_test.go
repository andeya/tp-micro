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
}
