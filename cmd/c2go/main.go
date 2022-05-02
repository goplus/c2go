package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/goplus/c2go/cl"
	"github.com/goplus/c2go/internal/c2go"
	"github.com/goplus/gox"
)

var (
	verbose    = flag.Bool("v", false, "print verbose information")
	failfast   = flag.Bool("ff", false, "fail fast (stop if an error is encountered)")
	preprocess = flag.Bool("pp", false, "force to run preprocessor")
	gendeps    = flag.Bool("gendeps", false, "generate dependencies automatically")
	test       = flag.Bool("test", false, "run test")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: c2go [-test -ff -gendeps -v] [pkgname] source.c\n")
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
	var flags int
	switch flag.NArg() {
	case 1:
		pkgname, infile, flags = "main", flag.Arg(0), c2go.FlagRunApp
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
	if *test {
		flags |= c2go.FlagRunTest
	}
	if *failfast {
		flags |= c2go.FlagFailFast
	}
	if *gendeps {
		flags |= c2go.FlagDepsAutoGen
	}
	if *preprocess {
		flags |= c2go.FlagForcePreprocess
	}
	c2go.Run(pkgname, infile, flags)
}
