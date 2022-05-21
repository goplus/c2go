package c2go

import (
	"flag"
	"fmt"
	"os"

	"github.com/goplus/c2go"
	"github.com/goplus/c2go/cl"
	"github.com/goplus/c2go/clang/preprocessor"
	"github.com/goplus/gox"
)

func isDir(name string) bool {
	if fi, err := os.Lstat(name); err == nil {
		return fi.IsDir()
	}
	return false
}

func Main(flag *flag.FlagSet, args []string) {
	var (
		verbose    = flag.Bool("v", false, "print verbose information")
		failfast   = flag.Bool("ff", false, "fail fast (stop if an error is encountered)")
		preprocess = flag.Bool("pp", false, "force to run preprocessor")
		gendeps    = flag.Bool("gendeps", false, "generate dependencies automatically")
		json       = flag.Bool("json", false, "dump C AST to a file in json format")
		test       = flag.Bool("test", false, "run test")
		testmain   = flag.Bool("testmain", false, "generate TestMain as entry instead of main (only for cmd/test_xxx)")
		sel        = flag.String("sel", "", "select a file (only available in project mode)")
	)
	flag.Parse(args)
	var pkgname, infile string
	var flags int
	switch flag.NArg() {
	case 1:
		pkgname, infile, flags = "main", flag.Arg(0), c2go.FlagRunApp
	case 2:
		pkgname, infile = flag.Arg(0), flag.Arg(1)
	default:
		fmt.Fprintf(os.Stderr, "Usage: c2go [-test -ff -gendeps -v] [pkgname] source.c\n")
		flag.PrintDefaults()
		return
	}

	if *verbose {
		cl.SetDebug(cl.DbgFlagAll)
		preprocessor.SetDebug(preprocessor.DbgFlagAll)
		gox.SetDebug(gox.DbgFlagInstruction) // | gox.DbgFlagMatch)
	}
	if *test {
		flags |= c2go.FlagRunTest
	}
	if *testmain {
		flags |= c2go.FlagTestMain
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
	if *json {
		flags |= c2go.FlagDumpJson
	}
	var conf *c2go.Config
	if *sel != "" {
		conf = &c2go.Config{Select: *sel}
	}
	c2go.Run(pkgname, infile, flags, conf)
}
