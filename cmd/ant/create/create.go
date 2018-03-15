package create

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/cmd/ant/create/def"
	"github.com/henrylee2cn/ant/cmd/ant/create/test"
	"github.com/henrylee2cn/ant/cmd/ant/create/tpl"
	"github.com/henrylee2cn/ant/cmd/ant/info"
	"github.com/henrylee2cn/goutil"
)

// CreateProject creates a project.
func CreateProject(scriptFile string) {
	noScriptFile := len(scriptFile) == 0
	if !noScriptFile {
		var err error
		scriptFile, err = filepath.Abs(scriptFile)
		if err != nil {
			ant.Fatalf("[ant] Invalid script file path")
		}
	}

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err := os.Chdir(info.AbsPath())
	if err != nil {
		ant.Fatalf("[ant] Jump working directory failed: %v", err)
	}

	if noScriptFile {
		def.Create()
		format()
		return
	}

	// creats base files
	if !goutil.FileExists("main.go") {
		tpl.Create()
	}

	var b []byte
	if noScriptFile {
		b = test.MustAsset("test.ant")
	} else {
		b, err = ioutil.ReadFile(scriptFile)
		if err != nil {
			ant.Fatalf("[ant] Write project files failed: %v", err)
		}
	}

	{
		r := bytes.NewReader(b)
		lexer := Lexer{}
		lexer.init(r)

		parser := Parser{}
		parser.init(&lexer)
		parser.parse()

		codeGen := CodeGen{}
		codeGen.init(&parser)
		codeGen.genForGolang()
	}

	// write script file
	f, err := os.OpenFile(info.ProjName()+".ant", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		ant.Fatalf("[ant] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(b)

	format()
}

// format the codes
func format() {
	cmd := exec.Command("gofmt", "-w", "./")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}
