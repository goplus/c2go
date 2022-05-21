package main

import (
	"flag"
	"fmt"
	"os"

	c2go "github.com/goplus/c2go/cmd/c2go/impl"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: "+c2go.ShortUsage)
		flag.PrintDefaults()
	}
	c2go.Main(flag.CommandLine, os.Args[1:])
}
