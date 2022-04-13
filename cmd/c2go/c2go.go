package main

import (
	"bytes"
	"flag"
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

var (
	verbose = flag.Bool("v", false, "print verbose information")
	test    = flag.Bool("test", false, "run test")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: c2go [-test -v] [pkgname] source.c\n")
	flag.PrintDefaults()
}

func isDir(name string) bool {
	if fi, err := os.Lstat(name); err == nil {
		return fi.IsDir()
	}
	return false
}

func main() {
	flag.Parse()
	var pkgname, infile string
	var run bool
	switch flag.NArg() {
	case 1:
		pkgname, infile, run = "main", flag.Arg(0), true
	case 2:
		pkgname, infile = flag.Arg(0), flag.Arg(1)
	default:
		usage()
		return
	}

	if *verbose {
		cl.SetDebug(cl.DbgFlagAll)
		gox.SetDebug(gox.DbgFlagInstruction) // | gox.DbgFlagMatch)
	}

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
			execDirRecursively(infile, run, *test)
		} else if isDir(infile) {
			switch n := execDir(pkgname, infile, run, *test); n {
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
	execFile(pkgname, outfile, run, *test)
}

func execDirRecursively(dir string, doRunApp, doRunTest bool) {
	if strings.HasPrefix(dir, "_") {
		return
	}
	fis, err := os.ReadDir(dir)
	check(err)
	var cfiles int
	for _, fi := range fis {
		if fi.IsDir() {
			pkgDir := path.Join(dir, fi.Name())
			execDirRecursively(pkgDir, doRunApp, doRunTest)
			continue
		}
		if strings.HasSuffix(fi.Name(), ".c") {
			cfiles++
		}
	}
	if cfiles == 1 {
		if doRunTest {
			fmt.Printf("Testing %s ...\n", dir)
		}
		execDir("main", dir, doRunApp, doRunTest)
	}
}

func execDir(pkgname string, dir string, doRunApp, doRunTest bool) int {
	cwd, err := os.Getwd()
	check(err)

	err = os.Chdir(dir)
	check(err)

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
		execFile(pkgname, outfile, doRunApp, doRunTest)
		fallthrough
	default:
		return n
	}
}

func execFile(pkgname string, outfile string, doRunApp, doRunTest bool) {
	doc, _, err := parser.ParseFile(outfile, 0)
	check(err)

	pkg, err := cl.NewPackage("", pkgname, doc, nil)
	check(err)

	gofile := outfile + ".go"
	err = gox.WriteFile(gofile, pkg, false)
	check(err)

	if doRunTest {
		runTest("")
	} else if doRunApp {
		runGoApp("", os.Stdout, os.Stderr, false)
	}
}

func checkEqual(prompt string, a, expected []byte) {
	if bytes.Equal(a, expected) {
		return
	}

	fmt.Fprintln(os.Stderr, "==> Result of", prompt)
	os.Stderr.Write(a)

	fmt.Fprintln(os.Stderr, "\n==> Expected", prompt)
	os.Stderr.Write(expected)
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
				dontRunTest = true
				return
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

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
