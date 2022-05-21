package main

import (
	"flag"
	"os"

	c2go "github.com/goplus/c2go/cmd/c2go/impl"
)

func main() {
	c2go.Main(flag.CommandLine, os.Args[1:])
}
