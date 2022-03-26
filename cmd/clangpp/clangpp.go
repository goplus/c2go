package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/goplus/c2go/clang/preprocessor"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: clangpp source.c\n")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	infile := os.Args[1]
	dir, fname := filepath.Split(infile)
	if pos := strings.LastIndex(fname, "."); pos >= 0 {
		fname = fname[:pos]
	}
	outfile := dir + fname + ".i"
	if err := preprocessor.Do(infile, outfile, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
