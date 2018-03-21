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

	// read temptale file
	var noTplFile = len(tplFile) == 0
	if noTplFile {
		tplFile = defAntTpl
	}

	absTplFile, err := filepath.Abs(tplFile)
	if err != nil {
		ant.Fatalf("[ant] Invalid template file: %s", tplFile)
	}

	b, err := ioutil.ReadFile(absTplFile)
	if err != nil {
		if !noTplFile {
			ant.Fatalf("[ant] Write project files failed: %v", err)
		} else {
			b = test.MustAsset(defAntTpl)
		}
	}

	// creates project

	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err = os.Chdir(info.AbsPath())
	if err != nil {
		ant.Fatalf("[ant] Jump working directory failed: %v", err)
	}

	// creates base files
	if !goutil.FileExists("main.go") {
		tpl.Create()
	}

	// new project code
	proj := NewProject(b)
	proj.Prepare()
	proj.Generator()

	// write template file
	f, err := os.OpenFile(defAntTpl, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		ant.Fatalf("[ant] Create files error: %v", err)
	}
	defer f.Close()
	f.Write(formatSource(b))

	ant.Infof("Completed code generation!")
}
