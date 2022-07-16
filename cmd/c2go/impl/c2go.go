package c2go

import (
	"flag"
	"os"

	"github.com/goplus/c2go"
	"github.com/goplus/c2go/cl"
	"github.com/goplus/c2go/clang/preprocessor"
	"github.com/goplus/gox"
)

const (
	ShortUsage = "c2go [-test -testmain -ff -pp -json -sel selectfile -gendeps -v] [pkgname] source\n"
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
		runcmd     = flag.String("run", "", "select a command to run (only available in project mode)")
		selfile    = flag.String("sel", "", "select a file (only available in project mode)")
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
		flag.Usage()
		return
	}

	if *verbose {
		cl.SetDebug(cl.DbgFlagAll)
		preprocessor.SetDebug(preprocessor.DbgFlagAll)
		gox.SetDebug(gox.DbgFlagInstruction) // | gox.DbgFlagMatch)
	}
	if *test || *runcmd != "" {
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
	conf := &c2go.Config{
		SelectFile: *selfile,
		SelectCmd:  *runcmd,
	}
	c2go.Run(pkgname, infile, flags, conf)
}
