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
	"bytes"
	"errors"
	"fmt"
	"go/build"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/goplus/c2go/cl"
	"github.com/goplus/c2go/clang/parser"
	"github.com/goplus/c2go/clang/preprocessor"
)

const (
	FlagRunApp = 1 << iota
	FlagRunTest
	FlagFailFast
	FlagDepsAutoGen
	FlagForcePreprocess
	FlagDumpJson
	FlagTestMain

	flagChdir
)

func isDir(name string) bool {
	if fi, err := os.Lstat(name); err == nil {
		return fi.IsDir()
	}
	return false
}

func isFile(name string) bool {
	if fi, err := os.Lstat(name); err == nil {
		return !fi.IsDir()
	}
	return false
}

type Config struct {
	SelectFile string
	SelectCmd  string
}

func Run(pkgname, infile string, flags int, conf *Config) {
	outfile := infile
	switch filepath.Ext(infile) {
	case ".i":
	case ".c":
		outfile = infile + ".i"
		err := preprocessor.Do(infile, outfile, nil)
		check(err)
	default:
		if strings.HasSuffix(infile, "/...") {
			infile = strings.TrimSuffix(infile, "/...")
			err := execDirRecursively(infile, flags, conf)
			check(err)
		} else if isDir(infile) {
			projfile := filepath.Join(infile, "c2go.cfg")
			if isFile(projfile) {
				execProj(projfile, flags, conf)
				return
			}
			n, err := execDir(pkgname, infile, flags)
			check(err)
			switch n {
			case 1:
			case 0:
				fatalf("no *.c files in this directory.\n")
			default:
				fatalf("multiple .c files found (currently only support one .c file).\n")
			}
		} else {
			fatalf("%s is not a .c file.\n", infile)
		}
		return
	}
	execFile(pkgname, outfile, flags|flagChdir)
	return
}

func execDirRecursively(dir string, flags int, conf *Config) (last error) {
	if strings.HasPrefix(dir, "_") {
		return
	}

	projfile := filepath.Join(dir, "c2go.cfg")
	if isFile(projfile) {
		fmt.Printf("==> Compiling %s ...\n", dir)
		execProj(projfile, flags, conf)
		return
	}

	fis, last := os.ReadDir(dir)
	check(last)
	var cfiles int
	for _, fi := range fis {
		if fi.IsDir() {
			pkgDir := filepath.Join(dir, fi.Name())
			if e := execDirRecursively(pkgDir, flags, conf); e != nil {
				last = e
			}
			continue
		}
		if strings.HasSuffix(fi.Name(), ".c") {
			cfiles++
		}
	}
	if cfiles == 1 {
		var action string
		switch {
		case (flags & FlagRunTest) != 0:
			action = "Testing"
		case (flags & FlagRunApp) != 0:
			action = "Running"
		default:
			action = "Compiling"
		}
		fmt.Printf("==> %s %s ...\n", action, dir)
		if _, e := execDir("main", dir, flags); e != nil {
			last = e
		}
	}
	return
}

func execDir(pkgname string, dir string, flags int) (n int, err error) {
	if (flags & FlagFailFast) == 0 {
		defer func() {
			if e := recover(); e != nil {
				err = newError(e)
			}
		}()
	}
	n = -1

	cwd := chdir(dir)
	defer os.Chdir(cwd)

	var infile, outfile string
	files, err := filepath.Glob("*.c")
	check(err)
	switch n = len(files); n {
	case 1:
		infile = files[0]
		outfile = infile + ".i"
		err = preprocessor.Do(infile, outfile, nil)
		check(err)
		execFile(pkgname, outfile, flags)
	}
	return
}

func execFile(pkgname string, outfile string, flags int) {
	var json []byte
	doc, _, err := parser.ParseFileEx(outfile, 0, &parser.Config{
		Json:   &json,
		Stderr: true,
	})
	check(err)

	if (flags & FlagDumpJson) != 0 {
		os.WriteFile(strings.TrimSuffix(outfile, ".i")+".json", json, 0666)
	}

	needPkgInfo := (flags & FlagDepsAutoGen) != 0
	pkg, err := cl.NewPackage("", pkgname, doc, &cl.Config{
		SrcFile: outfile, NeedPkgInfo: needPkgInfo, ClangTarget: clangTarget,
	})
	check(err)

	gofile := outfile + ".go"
	err = pkg.WriteFile(gofile)
	check(err)

	dir, _ := filepath.Split(gofile)

	if needPkgInfo {
		err = pkg.WriteDepFile(filepath.Join(dir, "c2go_autogen.go"))
		check(err)
	}

	if (flags & flagChdir) != 0 {
		if dir != "" {
			cwd := chdir(dir)
			defer os.Chdir(cwd)
		}
	}

	if (flags & FlagRunTest) != 0 {
		runTest("")
	} else if (flags & FlagRunApp) != 0 {
		runGoApp("", os.Stdout, os.Stderr, false)
	}
}

func checkEqual(prompt string, a, expected []byte) {
	if bytes.Equal(a, expected) {
		return
	}

	fmt.Fprintln(os.Stderr, "=> Result of", prompt)
	os.Stderr.Write(a)

	fmt.Fprintln(os.Stderr, "\n=> Expected", prompt)
	os.Stderr.Write(expected)

	fatal(errors.New("checkEqual: unexpected " + prompt))
}

func cleanEndLine(data []byte) []byte {
	return bytes.ReplaceAll(data, []byte{'\r', '\n'}, []byte{'\n'})
}

func runTest(dir string) {
	var goOut, goErr bytes.Buffer
	var cOut, cErr bytes.Buffer
	dontRunTest := runGoApp(dir, &goOut, &goErr, true)
	if dontRunTest {
		return
	}
	runCApp(dir, &cOut, &cErr)
	checkEqual("output", goOut.Bytes(), cleanEndLine(cOut.Bytes()))
	checkEqual("stderr", goErr.Bytes(), cleanEndLine(cErr.Bytes()))
}

func goFiles(dir string) ([]string, error) {
	if dir == "" {
		dir = "."
	}
	ctx := build.Default
	ctx.BuildTags = []string{getBuildTags()}
	bp, err := ctx.ImportDir(dir, 0)
	if err != nil {
		return nil, err
	}
	return append(bp.GoFiles, bp.TestGoFiles...), nil
}

func runGoApp(dir string, stdout, stderr io.Writer, doRunTest bool) (dontRunTest bool) {
	files, err := goFiles(dir)
	check(err)

	if doRunTest {
		for _, file := range files {
			if filepath.Base(file) == "main.go" {
				stdout, stderr = os.Stdout, os.Stderr
				dontRunTest = true
				break
			}
		}
	}
	cmd := exec.Command("go", append([]string{"run"}, files...)...)
	cmd.Dir = dir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	checkWith(cmd.Run(), stdout, stderr)
	return
}

func runCApp(dir string, stdout, stderr io.Writer) {
	files, err := filepath.Glob("*.c")
	check(err)

	cmd := exec.Command("clang", files...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	check(cmd.Run())

	cmd2 := exec.Command(clangOut)
	cmd.Dir = dir
	cmd2.Stdout = stdout
	cmd2.Stderr = stderr
	checkWith(cmd2.Run(), stdout, stderr)

	os.Remove(clangOut)
}

var (
	clangOut    = "./a.out"
	clangTarget string
)

func init() {
	if runtime.GOOS == "windows" {
		clangOut = "./a.exe"
	}
	clangTarget = getClangTarget()
}

func chdir(dir string) string {
	cwd, err := os.Getwd()
	check(err)

	err = os.Chdir(dir)
	check(err)

	return cwd
}

func check(err error) {
	if err != nil {
		fatal(err)
	}
}

func checkWith(err error, stdout, stderr io.Writer) {
	if err != nil {
		fatalWith(err, stdout, stderr)
	}
}

func fatalf(format string, args ...interface{}) {
	fatal(fmt.Errorf(format, args...))
}

func fatal(err error) {
	log.Panicln(err)
}

func fatalWith(err error, stdout, stderr io.Writer) {
	if o, ok := getBytes(stdout, stderr); ok {
		os.Stderr.Write(o.Bytes())
	}
	log.Panicln(err)
}

func newError(v interface{}) error {
	switch e := v.(type) {
	case error:
		return e
	case string:
		return errors.New(e)
	}
	fatalf("newError failed: %v", v)
	return nil
}

type iBytes interface {
	Bytes() []byte
}

func getBytes(stdout, stderr io.Writer) (o iBytes, ok bool) {
	if o, ok = stderr.(iBytes); ok {
		return
	}
	o, ok = stdout.(iBytes)
	return
}

func getClangTarget() string {
	cmd := exec.Command("clang", "--version")
	data, err := cmd.CombinedOutput()
	check(err)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "Target:") {
			return strings.TrimSpace(line[7:])
		}
	}
	return ""
}

func getBuildTags() string {
	if strings.HasSuffix(clangTarget, "-windows-msvc") {
		return "windows_msvc"
	} else if strings.HasSuffix(clangTarget, "-windows-gnu") {
		return "windows_gnu"
	}
	return ""
}
