package create

import (
	"testing"

	"github.com/henrylee2cn/ant/cmd/ant/info"
)

func TestGenerator(t *testing.T) {
	info.Init("test")
	proj := NewProject([]byte(src))
	proj.Prepare()
	proj.genTypesFile()
	proj.genRouterFile()
	proj.genHandlerAndLogicAndSdkFiles()
	t.Logf("types/types.gen.go:\n%s", codeFiles["types/types.gen.go"])
	t.Logf("logic/tmp_code.gen.go:\n%s", codeFiles["logic/tmp_code.gen.go"])
	t.Logf("api/handler.gen.go:\n%s", codeFiles["api/handler.gen.go"])
	t.Logf("api/router.gen.go:\n%s", codeFiles["api/router.gen.go"])
	t.Logf("sdk/rpc.gen.go:\n%s", codeFiles["sdk/rpc.gen.go"])
	t.Logf("sdk/rpc_test.gen.go:\n%s", codeFiles["sdk/rpc_test.gen.go"])
}

const src = `// package __ANT__TPL__ is the project template
package __ANT__TPL__

// __API__PULL__ register PULL router:
//  /home
//  /math/divide
type __API__PULL__ interface {
	Home(*struct{}) *HomeReply
	Math
}

// __API__PUSH__ register PUSH router:
//  /stat
type __API__PUSH__ interface {
	Stat(*StatArgs)
}

// Math controller
type Math interface {
	// Divide handler
	Divide(*DivideArgs) *DivideReply
}

// HomeReply home reply
type HomeReply struct {
	Content string // text
}

type (
	// DivideArgs divide api args
	DivideArgs struct {
		// dividend
		A float64
		// divisor
		B float64 ` + "`param:\"<range: 0.01:100000>\"`" + `
	}
	// DivideReply divide api result
	DivideReply struct {
		// quotient
		C float64
	}
)

// StatArgs stat handler args
type StatArgs struct {
	Ts int64 // timestamps
}

// StatArgsCopy StatArgs copy
type StatArgsCopy StatArgs
`
