package create

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/henrylee2cn/ant"
	"github.com/henrylee2cn/ant/cmd/ant/info"
)

var (
	projName, projPath []byte
	projNameTpl        = []byte("{{PROJ_NAME}}")
	projPathTpl        = []byte("{{PROJ_PATH}}")
)

// CreateProject creates a project.
func CreateProject() {
	os.MkdirAll(info.AbsPath(), os.FileMode(0755))
	err := os.Chdir(info.AbsPath())
	if err != nil {
		ant.Fatalf("[ant] Jump working directory failed: %v", err)
	}
	{
		projName = []byte(info.ProjName())
		projPath = []byte(info.ProjPath())
	}
	err = restoreAssets("./", "")
	if err != nil {
		ant.Fatalf("[ant] Write project files failed: %v", err)
	}
}

// restoreAssets restores an asset under the given directory recursively
func restoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return restoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = restoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

// restoreAsset restores an asset under the given directory
func restoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	data = bytes.Replace(data, projNameTpl, projName, -1)
	data = bytes.Replace(data, projPathTpl, projPath, -1)
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}
