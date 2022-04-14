package cl

import (
	"log"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/parser"
	"github.com/goplus/c2go/clang/preprocessor"
)

// -----------------------------------------------------------------------------

var (
	tmpDir     string
	tmpFileIdx int64
)

func init() {
	home, err := os.UserHomeDir()
	check(err)

	tmpDir = home + "/.c2go/tmp/"
	err = os.MkdirAll(tmpDir, 0755)
	check(err)
}

func parse(code string) (doc *ast.Node, src []byte) {
	idx := atomic.AddInt64(&tmpFileIdx, 1)
	infile := tmpDir + strconv.FormatInt(idx, 10) + ".c"
	err := os.WriteFile(infile, []byte(code), 0666)
	check(err)

	outfile := infile + ".i"
	err = preprocessor.Do(infile, outfile, nil)
	check(err)
	os.Remove(infile)

	src, err = os.ReadFile(outfile)
	check(err)

	doc, _, err = parser.ParseFile(outfile, 0)
	check(err)
	os.Remove(outfile)
	return
}

func check(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

// -----------------------------------------------------------------------------
