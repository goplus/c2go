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
	Dirs  []string `json:"dirs"`
	Files []string `json:"files"`
}

type c2goConf struct {
	Target  c2goTarget `json:"target"`
	Source  c2goSource `json:"source"`
	Include []string   `json:"include"`
	Define  []string   `json:"define"`
	Flags   []string   `json:"flags"`

	public    map[string]string `json:"-"`
	cl.Reused `json:"-"`
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func loadPubFile(pubfile string) map[string]string {
	b, err := os.ReadFile(pubfile)
	check(err)

	text := string(b)
	lines := strings.Split(text, "\n")
	ret := make(map[string]string, len(lines))
	for i, line := range lines {
		flds := strings.Fields(line)
		goName := ""
		switch len(flds) {
		case 1:
		case 2:
			goName = flds[1]
		case 0:
			continue
		default:
			fatalf("line %d: too many fields - %s\n", i+1, line)
		}
		ret[flds[0]] = goName
	}
	return ret
}

func execProj(projfile string, flags int) {
	b, err := os.ReadFile(projfile)
	check(err)

	var conf c2goConf
	err = json.Unmarshal(b, &conf)
	check(err)

	if conf.Target.Name == "" {
		conf.Target.Name = "main"
	}
	if len(conf.Source.Dirs) == 0 && len(conf.Source.Files) == 0 {
		conf.Source.Dirs = []string{"."}
	}

	base, _ := filepath.Split(projfile)

	pubfile := base + "c2go.pub"
	conf.public = loadPubFile(pubfile)

	for _, dir := range conf.Source.Dirs {
		execProjDir(resolvePath(base, dir), &conf, flags)
	}
	for _, file := range conf.Source.Files {
		execProjFile(resolvePath(base, file), &conf, flags)
	}

	if pkg := conf.Reused.Pkg(); pkg != nil {
		pkg.ForEachFile(func(fname string, file *gox.File) {
			dir := resolvePath(base, conf.Target.Dir)
			err = gox.WriteFile(filepath.Join(dir, fname), pkg, fname)
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
	if (flags&FlagForcePreprocess) != 0 || !isFile(outfile) {
		err := preprocessor.Do(infile, outfile, &preprocessor.Config{
			IncludeDirs: conf.Include,
			Defines:     conf.Define,
			Flags:       conf.Flags,
		})
		check(err)
	}

	doc, _, err := parser.ParseFile(outfile, 0)
	check(err)

	_, err = cl.NewPackage("", conf.Target.Name, doc, &cl.Config{
		SrcFile: outfile,
		Public:  conf.public,
		Reused:  &conf.Reused,
	})
	check(err)
}

func resolvePath(base string, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, path)
}
