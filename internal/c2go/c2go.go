package c2go

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
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

func Run(pkgname, infile string, flags int) {
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
			err := execDirRecursively(infile, flags)
			check(err)
		} else if isDir(infile) {
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

func execDirRecursively(dir string, flags int) (last error) {
	if strings.HasPrefix(dir, "_") {
		return
	}
	fis, last := os.ReadDir(dir)
	check(last)
	var cfiles int
	for _, fi := range fis {
		if fi.IsDir() {
			pkgDir := path.Join(dir, fi.Name())
			if e := execDirRecursively(pkgDir, flags); e != nil {
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

	for i, n := 0, len(files); i < n; i++ {
		fname := filepath.Base(files[i])
		if pos := strings.LastIndex(fname, "_"); pos >= 0 {
			switch os := fname[pos+1 : len(fname)-3]; os {
			case "darwin", "linux", "windows":
				if os != runtime.GOOS { // skip
					n--
					files[i], files[n] = files[n], files[i]
					files = files[:n]
					i--
				}
			}
		}
	}

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
	clangOut = "./a.out"
)

func init() {
	if runtime.GOOS == "windows" {
		clangOut = "./a.exe"
	}
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
