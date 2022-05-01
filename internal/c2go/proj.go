package c2go

import (
	"fmt"
	"os"
	"path"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

type c2goTarget struct {
	Name string `json:"name"`
	Dir  string `json:"dir"`
}

type c2goSource struct {
	Dir []string `json:"dir"`
}

type c2goConf struct {
	Target  c2goTarget `json:"target"`
	Source  c2goSource `json:"source"`
	Include []string   `json:"include"`
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func execProj(projfile string, flags int) {
	b, err := os.ReadFile(projfile)
	check(err)

	var conf c2goConf
	err = json.Unmarshal(b, &conf)
	check(err)

	for _, dir := range conf.Source.Dir {
		execProjDir(dir, &conf, flags)
	}
}

func execProjDir(dir string, conf *c2goConf, flags int) {
	if strings.HasPrefix(dir, "_") {
		return
	}
	fis, err := os.ReadDir(dir)
	check(err)
	for _, fi := range fis {
		if fi.IsDir() {
			pkgDir := path.Join(dir, fi.Name())
			execProjDir(pkgDir, conf, flags)
			continue
		}
		if strings.HasSuffix(fi.Name(), ".c") {
			pkgFile := path.Join(dir, fi.Name())
			execProjFile(pkgFile, conf, flags)
		}
	}
}

func execProjFile(file string, conf *c2goConf, flags int) {
	fmt.Printf("==> Compiling %s ...\n", file)
}
