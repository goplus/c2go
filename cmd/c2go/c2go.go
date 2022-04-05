package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/goplus/c2go/cl"
	"github.com/goplus/c2go/clang/parser"
	"github.com/goplus/c2go/clang/preprocessor"
	"github.com/goplus/gox"
	"github.com/goplus/gox/packages"
)

var (
	verbose = flag.Bool("v", false, "print verbose information")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: c2go [-v] [pkgname] source.c\n")
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
		gox.SetDebug(gox.DbgFlagInstruction)
	}

	outfile := infile
	switch filepath.Ext(infile) {
	case ".i":
	case ".c":
		outfile = infile + ".i"
		err := preprocessor.Do(infile, outfile, nil)
		check(err)
	default:
		if isDir(infile) {
			files, err := filepath.Glob("*.c")
			check(err)
			switch len(files) {
			case 1:
				infile = files[0]
				outfile = infile + ".i"
				err := preprocessor.Do(infile, outfile, nil)
				check(err)
			case 0:
				fatalf("no *.c files in this directory.\n")
			default:
				fatalf("multiple .c files found (currently only support one .c file).\n")
			}
		} else {
			fatalf("%s is not a .c file.\n", infile)
		}
	}

	doc, _, err := parser.ParseFile(outfile, 0)
	check(err)

	imp, _, _ := packages.NewImporter(nil, "fmt", "strings", "strconv")
	confCl := &cl.Config{Importer: imp}
	pkg, err := cl.NewPackage("", pkgname, doc, confCl)
	check(err)

	gofile := outfile + ".go"
	err = gox.WriteFile(gofile, pkg, false)
	check(err)

	if run {
		files, err := filepath.Glob("*.go")
		check(err)

		cmd := exec.Command("go", append([]string{"run"}, files...)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		check(cmd.Run())
	}
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
