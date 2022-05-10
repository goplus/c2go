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

type c2goIgnore struct {
	Names []string `json:"names"`
}

type c2goSource struct {
	Dirs   []string   `json:"dirs"`
	Files  []string   `json:"files"`
	Ignore c2goIgnore `json:"ignore"`
}

type c2goCmd struct {
	Dir    string     `json:"dir"`
	Source c2goSource `json:"source"`
	Deps   []string   `json:"deps"`
}

type c2goTarget struct {
	Name string    `json:"name"`
	Dir  string    `json:"dir"`
	Cmds []c2goCmd `json:"cmds"`
}

type c2goConf struct {
	Target   c2goTarget `json:"target"`
	Source   c2goSource `json:"source"`
	Include  []string   `json:"include"`
	Deps     []string   `json:"deps"`
	Define   []string   `json:"define"`
	Flags    []string   `json:"flags"`
	PPFlag   string     `json:"pp"` // default: -E
	Compiler string     `json:"cc"`

	cl.Reused `json:"-"`

	public      map[string]string `json:"-"`
	needPkgInfo bool              `json:"-"`
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func loadPubFile(pubfile string) map[string]string {
	b, err := os.ReadFile(pubfile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		check(err)
	}

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

func execProj(projfile string, flags int, in *Config) {
	b, err := os.ReadFile(projfile)
	check(err)

	var conf c2goConf
	err = json.Unmarshal(b, &conf)
	check(err)

	conf.needPkgInfo = (flags & FlagDepsAutoGen) != 0
	base, _ := filepath.Split(projfile)
	noSource := len(conf.Source.Dirs) == 0 && len(conf.Source.Files) == 0
	if noSource {
		if len(conf.Target.Cmds) == 0 {
			fatalf("empty project: no source files specified in c2go.cfg.\n")
			return
		}
	} else {
		if conf.Target.Name == "" {
			conf.Target.Name = "main"
		}

		pubfile := base + "c2go.pub"
		conf.public = loadPubFile(pubfile)

		if in != nil && in.Select != "" {
			execProjFile(resolvePath(base, in.Select), &conf, flags)
		} else {
			execProjSource(base, flags, &conf)
		}
		execProjDone(base, flags, &conf)
	}
	if cmds := conf.Target.Cmds; len(cmds) != 0 {
		conf.Target.Cmds = nil
		conf.Target.Name = "main"
		for _, cmd := range cmds {
			conf.Source = cmd.Source
			conf.Deps = cmd.Deps
			conf.Target.Dir = cmd.Dir
			execProjSource(base, flags, &conf)
			execProjDone(base, flags, &conf)
		}
	}
}

func execProjDone(base string, flags int, conf *c2goConf) {
	if pkg := conf.Reused.Pkg(); pkg.IsValid() {
		dir := resolvePath(base, conf.Target.Dir)
		os.MkdirAll(dir, 0777)
		pkg.ForEachFile(func(fname string, file *gox.File) {
			gofile := fname
			if strings.HasPrefix(fname, "_") {
				gofile = "c2go" + fname
			}
			err := pkg.WriteFile(filepath.Join(dir, gofile), fname)
			check(err)
		})
		if conf.needPkgInfo {
			err := pkg.WriteDepFile(filepath.Join(dir, "c2go_autogen.go"))
			check(err)
		}
		cmd := exec.Command("go", "install", ".")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		check(cmd.Run())
	} else {
		fatalf("empty project: no *.c files in this directory.\n")
	}
}

func execProjSource(base string, flags int, conf *c2goConf) {
	for _, dir := range conf.Source.Dirs {
		recursively := strings.HasSuffix(dir, "/...")
		if recursively {
			dir = dir[:len(dir)-4]
		}
		execProjDir(resolvePath(base, dir), conf, flags, recursively)
	}
	for _, file := range conf.Source.Files {
		execProjFile(resolvePath(base, file), conf, flags)
	}
}

func execProjDir(dir string, conf *c2goConf, flags int, recursively bool) {
	if strings.HasPrefix(dir, "_") {
		return
	}
	fis, err := os.ReadDir(dir)
	check(err)
	for _, fi := range fis {
		fname := fi.Name()
		if fi.IsDir() {
			if recursively {
				pkgDir := filepath.Join(dir, fname)
				execProjDir(pkgDir, conf, flags, true)
			}
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
			PPFlag:      conf.PPFlag,
			Compiler:    conf.Compiler,
		})
		check(err)
	}

	var json []byte
	doc, _, err := parser.ParseFileEx(outfile, 0, &parser.Config{
		Json:   &json,
		Flags:  conf.Flags,
		Stderr: true,
	})
	check(err)

	if (flags & FlagDumpJson) != 0 {
		os.WriteFile(infile+".json", json, 0666)
	}

	_, err = cl.NewPackage("", conf.Target.Name, doc, &cl.Config{
		SrcFile:     outfile,
		Public:      conf.public,
		NeedPkgInfo: conf.needPkgInfo,
		Ignored:     conf.Source.Ignore.Names,
		Reused:      &conf.Reused,
	})
	check(err)
}

func resolvePath(base string, path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(base, path)
}
