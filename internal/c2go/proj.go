package c2go

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goplus/c2go/cl"
	"github.com/goplus/c2go/clang/parser"
	"github.com/goplus/c2go/clang/preprocessor"
	"github.com/goplus/gox"

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

	if conf.Target.Name == "" {
		conf.Target.Name = "main"
	}
	if len(conf.Source.Dir) == 0 {
		conf.Source.Dir = []string{"."}
	}
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
		fname := fi.Name()
		if fi.IsDir() {
			pkgDir := filepath.Join(dir, fname)
			execProjDir(pkgDir, conf, flags)
			continue
		}
		if strings.HasSuffix(fi.Name(), ".c") {
			pkgFile := filepath.Join(dir, fname)
			targetFile := filepath.Join(conf.Target.Dir, fname+".i.go")
			execProjFile(pkgFile, targetFile, conf, flags)
		}
	}
}

func execProjFile(infile, gofile string, conf *c2goConf, flags int) {
	fmt.Printf("==> Compiling %s ...\n", infile)

	outfile := infile + ".i"
	if !isFile(outfile) {
		err := preprocessor.Do(infile, outfile, &preprocessor.Config{
			IncludeDirs: conf.Include,
		})
		check(err)
	}

	doc, _, err := parser.ParseFile(outfile, 0)
	check(err)

	pkg, err := cl.NewPackage("", conf.Target.Name, doc, &cl.Config{SrcFile: outfile})
	check(err)

	err = gox.WriteFile(gofile, pkg.Package, false)
	check(err)
}
