package c2go

import (
	"fmt"
	"os"
	"os/exec"
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

	cl.Reused `json:"-"`
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

	base, _ := filepath.Split(projfile)
	for _, dir := range conf.Source.Dir {
		execProjDir(filepath.Join(base, dir), &conf, flags)
	}

	if pkg := conf.Reused.Pkg(); pkg != nil {
		pkg.ForEachFile(func(fname string, file *gox.File) {
			err = gox.WriteFile(filepath.Join(conf.Target.Dir, fname), pkg, fname)
			check(err)
		})
		cmd := exec.Command("go", "build", ".")
		cmd.Dir = base
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		check(cmd.Run())
	} else {
		fatalf("empty project: no *.c files in this directory.\n")
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
			execProjFile(pkgFile, conf, flags)
		}
	}
}

func execProjFile(infile string, conf *c2goConf, flags int) {
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

	_, err = cl.NewPackage("", conf.Target.Name, doc, &cl.Config{
		SrcFile: outfile,
		Reused:  &conf.Reused,
	})
	check(err)
}
