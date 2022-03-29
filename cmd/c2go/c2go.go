package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/goplus/c2go/cl"
	"github.com/goplus/c2go/clang/parser"
	"github.com/goplus/c2go/clang/preprocessor"
	"github.com/goplus/gox"
	"github.com/goplus/gox/packages"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: c2go pkgname source.c\n")
}

func main() {
	if len(os.Args) < 3 {
		usage()
		return
	}
	cl.SetDebug(cl.DbgFlagAll)
	gox.SetDebug(gox.DbgFlagInstruction)

	pkgname, infile := os.Args[1], os.Args[2]
	outfile := infile
	if filepath.Ext(infile) != ".i" {
		outfile = infile + ".i"
		err := preprocessor.Do(infile, outfile, nil)
		checkerr(err)
	}

	doc, _, err := parser.ParseFile(outfile, 0)
	checkerr(err)

	imp, _, _ := packages.NewImporter(nil, "fmt", "strings", "strconv")
	confCl := &cl.Config{Importer: imp}
	pkg, err := cl.NewPackage("", pkgname, doc, confCl)
	checkerr(err)

	gofile := infile + ".go"
	err = gox.WriteFile(gofile, pkg, false)
	checkerr(err)
}

func checkerr(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
