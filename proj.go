/*
 * Copyright (c) 2022 The GoPlus Authors (goplus.org). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

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
	"github.com/goplus/gox/cpackages"

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
	For    []string   `json:"for"`
}

type c2goTarget struct {
	Name string    `json:"name"`
	Dir  string    `json:"dir"`
	Cmds []c2goCmd `json:"cmds"`
}

type c2goPublic struct {
	From   []string `json:"from"`
	Ignore []string `json:"ignore"`
}

type c2goConf struct {
	Public   c2goPublic `json:"public"`
	Target   c2goTarget `json:"target"`
	Source   c2goSource `json:"source"`
	Include  []string   `json:"include"`
	Deps     []string   `json:"deps"`
	Define   []string   `json:"define"`
	Flags    []string   `json:"flags"`
	PPFlag   string     `json:"pp"` // default: -E
	Compiler string     `json:"cc"`

	cl.Reused `json:"-"`

	dir         string            `json:"-"`
	public      map[string]string `json:"-"`
	needPkgInfo bool              `json:"-"`

	InLibC bool `json:"libc"` // bfm = BFM_InLibC

	SimpleProj bool `json:"simpleProj"` // bfm = BFM_Default
}

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func execProj(projfile string, flags int, in *Config) {
	b, err := os.ReadFile(projfile)
	check(err)

	var conf c2goConf
	err = json.Unmarshal(b, &conf)
	check(err)

	base, _ := filepath.Split(projfile)
	conf.needPkgInfo = (flags & FlagDepsAutoGen) != 0
	conf.dir = base
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

		appFlags := flags &^ (FlagTestMain | FlagRunTest)
		pubfile := base + "c2go.pub"
		conf.public, err = cpackages.ReadPubFile(pubfile)
		check(err)

		if in != nil && in.SelectFile != "" {
			execProjFile(canonical(base, in.SelectFile), &conf, appFlags)
			return
		}
		execProjSource(base, appFlags, &conf)

		err = cpackages.WritePubFile(base+"c2go.a.pub", conf.public)
		check(err)
	}
	if cmds := conf.Target.Cmds; len(cmds) != 0 {
		conf.Target.Cmds = nil
		conf.public = nil
		cmdSel := ""
		if in != nil {
			cmdSel = in.SelectCmd
		}
		doCmd := func(cmd c2goCmd) {
			if cmdSel != "" && cmd.Dir != cmdSel {
				return
			}
			conf.Target.Name = "main"
			appFlags := flags
			if (flags & FlagTestMain) != 0 {
				appFlags &= ^FlagRunTest // not allow both FlagTestMain and FlagRunTest
				dir, fname := filepath.Split(cmd.Dir)
				if strings.HasPrefix(fname, "test_") {
					conf.Target.Name = fname[5:]
					cmd.Dir = dir + "test/" + conf.Target.Name
				} else {
					appFlags &= ^FlagTestMain
				}
			} else if (appFlags & FlagRunTest) != 0 {
				fname := filepath.Base(cmd.Dir)
				if !strings.HasPrefix(fname, "test_") { // only test cmd/test_xxx
					appFlags &= ^FlagRunTest
				}
			}
			fmt.Printf("==> Building %s ...\n", cmd.Dir)
			conf.Source = cmd.Source
			conf.Deps = cmd.Deps
			conf.Target.Dir = cmd.Dir
			execProjSource(base, appFlags, &conf)
			if (appFlags & FlagRunTest) != 0 {
				cmd2 := exec.Command(clangOut)
				cmd2.Dir = canonical(base, cmd.Dir)
				cmd2.Stdout = os.Stdout
				cmd2.Stderr = os.Stderr
				fmt.Printf("==> Running %s ...\n", cmd2.Dir)
				check(cmd2.Run())
				os.Remove(filepath.Join(cmd2.Dir, clangOut))
			}
		}
		doCmdTempl := func(cmd c2goCmd, it string) {
			cmd.Dir = substText(cmd.Dir, it)
			cmd.Deps = substTexts(cmd.Deps, it)
			cmd.Source.Files = substTexts(cmd.Source.Files, it)
			cmd.Source.Dirs = substTexts(cmd.Source.Dirs, it)
			doCmd(cmd)
		}
		for _, cmd := range cmds {
			if len(cmd.For) > 0 {
				for _, it := range cmd.For {
					doCmdTempl(cmd, it)
				}
			} else {
				doCmd(cmd)
			}
		}
	}
}

func substText(tpl string, it string) string {
	return strings.ReplaceAll(tpl, "$(it)", it)
}

func substTexts(tpl []string, it string) []string {
	for n, v := range tpl {
		nv := substText(v, it)
		if v != nv {
			ret := make([]string, len(tpl))
			for i := 0; i < n; i++ {
				ret[i] = tpl[i]
			}
			ret[n] = nv
			for i, n := n+1, len(tpl); i < n; i++ {
				ret[i] = substText(tpl[i], it)
			}
			return ret
		}
	}
	return tpl
}

func execProjSource(base string, flags int, conf *c2goConf) {
	conf.Reused = cl.Reused{}
	for _, dir := range conf.Source.Dirs {
		recursively := strings.HasSuffix(dir, "/...")
		if recursively {
			dir = dir[:len(dir)-4]
		}
		execProjDir(canonical(base, dir), conf, flags, recursively)
	}
	for _, file := range conf.Source.Files {
		execProjFile(canonical(base, file), conf, flags)
	}
	execProjDone(base, flags, conf)
}

func execProjDone(base string, flags int, conf *c2goConf) {
	if pkg := conf.Reused.Pkg(); pkg.IsValid() {
		dir := canonical(base, conf.Target.Dir)
		os.MkdirAll(dir, 0777)
		pkg.ForEachFile(func(fname string, file *gox.File) {
			gofile := fname
			if strings.HasPrefix(fname, "_") {
				gofile = "x2g" + fname
			}
			err := pkg.WriteFile(filepath.Join(dir, gofile), fname)
			check(err)
		})
		if conf.needPkgInfo {
			err := pkg.WriteDepFile(filepath.Join(dir, "c2go_autogen.go"))
			check(err)
		}
		var cmd *exec.Cmd
		if (flags&FlagRunTest) != 0 && conf.Target.Name == "main" {
			cmd = exec.Command("go", "build", "-o", clangOut, ".")
		} else {
			cmd = exec.Command("go", "install", ".")
		}
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		check(cmd.Run())
	} else {
		fatalf("empty project: no *.c files in this directory.\n")
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
			BaseDir:     conf.dir,
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

	procDepPkg := func(pkgDir string) {
		headerFile := pkgDir + "/c2go_header.i.go"
		if !isFile(headerFile) {
			Run("", pkgDir, flags, nil)
		}
	}
	var bfm cl.BFMode
	if conf.InLibC {
		bfm = cl.BFM_InLibC
	} else if !conf.SimpleProj {
		bfm = cl.BFM_FromLibC
	}
	_, err = cl.NewPackage("", conf.Target.Name, doc, &cl.Config{
		SrcFile:      outfile,
		ProcDepPkg:   procDepPkg,
		Public:       conf.public,
		PublicFrom:   conf.Public.From,
		PublicIgnore: conf.Public.Ignore,
		NeedPkgInfo:  conf.needPkgInfo,
		Dir:          conf.dir,
		Deps:         conf.Deps,
		Include:      conf.Include,
		Ignored:      conf.Source.Ignore.Names,
		Reused:       &conf.Reused,
		TestMain:     (flags & FlagTestMain) != 0,
		// BuiltinFuncMode: compiling mode of builtin functions
		BuiltinFuncMode: bfm,
	})
	check(err)
}

func canonical(baseDir string, uri string) string {
	if filepath.IsAbs(uri) {
		return uri
	}
	return filepath.Join(baseDir, uri)
}
