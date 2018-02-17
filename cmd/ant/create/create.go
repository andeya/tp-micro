package create

import (
	"bytes"
	"io"
	"os"
	"os/exec"

	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/cmd/ant/create/def"
	"github.com/henrylee2cn/ant/cmd/ant/create/test"
	"github.com/henrylee2cn/ant/cmd/ant/create/tpl"
	"github.com/henrylee2cn/ant/cmd/ant/info"
)

// CreateProject creates a project.
func CreateProject(scriptFile string) {
	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err := os.Chdir(info.AbsPath())
	if err != nil {
		ant.Fatalf("[ant] Jump working directory failed: %v", err)
	}

	if len(scriptFile) == 0 {
		def.Create()
		format()
		return
	}

	// creats base files
	tpl.Create()

	var r io.Reader
	if len(scriptFile) == 0 {
		b := test.MustAsset("test.ant")
		r = bytes.NewReader(b)
	} else {
		file, err := os.Open(scriptFile)
		if err != nil {
			ant.Fatalf("[ant] Write project files failed: %v", err)
		}
		defer file.Close()
		r = file
	}

	{
		lexer := Lexer{}
		lexer.init(r)

		parser := Parser{}
		parser.init(&lexer)
		parser.parse()

		codeGen := CodeGen{}
		codeGen.init(&parser)
		codeGen.genForGolang()
	}

	format()
}

// format the codes
func format() {
	cmd := exec.Command("gofmt", "-w", "./")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}
