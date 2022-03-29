package main

import (
	"fmt"
	"os"

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
	outfile := infile + ".i"
	if err := preprocessor.Do(infile, outfile, nil); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
