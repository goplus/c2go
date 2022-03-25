package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/goplus/c2go/clang/parser"
)

var (
	dump = flag.Bool("dump", false, "dump AST")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: clangast [-dump] source.i\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
		return
	}
	var file = flag.Arg(0)
	var out []byte
	var err error
	if *dump {
		out, err = parser.DumpAST(file)
	} else if doc, e := parser.ParseFile(file, 0); e == nil {
		out, _ = json.MarshalIndent(doc, "", "  ")
	} else {
		err = e
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	os.Stdout.Write(out)
}
