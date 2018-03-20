package create

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/cmd/ant/create/test"
	"github.com/henrylee2cn/ant/cmd/ant/create/tpl"
	"github.com/henrylee2cn/ant/cmd/ant/info"
	"github.com/henrylee2cn/goutil"
)

const (
	defAntTpl = "__ant__tpl__.go"
)

// CreateProject creates a project.
func CreateProject(tplFile string) {
	ant.Infof("Generating project: %s", info.ProjPath())

	noScriptFile := len(tplFile) == 0
	if !noScriptFile {
		var err error
		tplFile, err = filepath.Abs(tplFile)
		if err != nil {
			ant.Fatalf("[ant] Invalid script file path")
		}
	}

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err := os.Chdir(info.AbsPath())
	if err != nil {
		ant.Fatalf("[ant] Jump working directory failed: %v", err)
	}

	// creats base files
	if !goutil.FileExists("main.go") {
		tpl.Create()
	}

	var b []byte
	if noScriptFile {
		b = test.MustAsset(defAntTpl)
	} else {
		b, err = ioutil.ReadFile(tplFile)
		if err != nil {
			ant.Fatalf("[ant] Write project files failed: %v", err)
		}
	}

	{
		proj := NewProject(b)
		proj.Prepare()
		proj.Generator()
	}

	// write script file
	f, err := os.OpenFile(defAntTpl, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		ant.Fatalf("[ant] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	ant.Infof("Completed code generation!")
}
