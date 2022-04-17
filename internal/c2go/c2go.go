package c2go

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/goplus/c2go/cl"
	"github.com/goplus/c2go/clang/parser"
	"github.com/goplus/c2go/clang/preprocessor"
	"github.com/goplus/gox"
)

const (
	FlagRunApp = 1 << iota
	FlagRunTest
	FlagFailFast
	flagChdir
)

func isDir(name string) bool {
	if fi, err := os.Lstat(name); err == nil {
		return fi.IsDir()
	}
	return false
}

func Run(pkgname, infile string, flags int) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = e.(error)
		}
	}()
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
			execDirRecursively(infile, flags)
		} else if isDir(infile) {
			switch n := execDir(pkgname, infile, flags); n {
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

func execDirRecursively(dir string, flags int) {
	if strings.HasPrefix(dir, "_") {
		return
	}
	fis, err := os.ReadDir(dir)
	check(err)
	var cfiles int
	for _, fi := range fis {
		if fi.IsDir() {
			pkgDir := path.Join(dir, fi.Name())
			execDirRecursively(pkgDir, flags)
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
		execDir("main", dir, flags)
	}
}

func execDir(pkgname string, dir string, flags int) int {
	defer func() {
		if e := recover(); e != nil && (flags&FlagFailFast) != 0 {
			panic(e)
		}
	}()

	cwd := chdir(dir)
	defer os.Chdir(cwd)

	var infile, outfile string
	files, err := filepath.Glob("*.c")
	check(err)
	switch n := len(files); n {
	case 1:
		infile = files[0]
		outfile = infile + ".i"
		err := preprocessor.Do(infile, outfile, nil)
		check(err)
		execFile(pkgname, outfile, flags)
		fallthrough
	default:
		return n
	}
}

func execFile(pkgname string, outfile string, flags int) {
	doc, _, err := parser.ParseFile(outfile, 0)
	check(err)

	pkg, err := cl.NewPackage("", pkgname, doc, &cl.Config{SrcFile: outfile})
	check(err)

	gofile := outfile + ".go"
	err = gox.WriteFile(gofile, pkg, false)
	check(err)

	if (flags & flagChdir) != 0 {
		dir, _ := filepath.Split(gofile)
		cwd := chdir(dir)
		defer os.Chdir(cwd)
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

func runTest(dir string) {
	var goOut, goErr bytes.Buffer
	var cOut, cErr bytes.Buffer
	dontRunTest := runGoApp(dir, &goOut, &goErr, true)
	if dontRunTest {
		return
	}
	runCApp(dir, &cOut, &cErr)
	checkEqual("output", goOut.Bytes(), cOut.Bytes())
	checkEqual("stderr", goErr.Bytes(), cErr.Bytes())
}

func runGoApp(dir string, stdout, stderr io.Writer, doRunTest bool) (dontRunTest bool) {
	files, err := filepath.Glob("*.go")
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
	check(cmd.Run())
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

	cmd2 := exec.Command("./a.out")
	cmd.Dir = dir
	cmd2.Stdout = stdout
	cmd2.Stderr = stderr
	check(cmd2.Run())

	os.Remove("./a.out")
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

func fatalf(format string, args ...interface{}) {
	fatal(fmt.Errorf(format, args...))
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	panic(err)
}
