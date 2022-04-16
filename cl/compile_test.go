package cl

import (
	"go/types"
	"log"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/parser"
	"github.com/goplus/c2go/clang/preprocessor"
	"github.com/goplus/gox"
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

func findNode(root *ast.Node, kind ast.Kind, name string) *ast.Node {
	if root.Kind == kind && root.Name == name {
		return root
	}
	for i, n := 0, len(root.Inner); i < n; i++ {
		if ret := findNode(root.Inner[i], kind, name); ret != nil {
			return ret
		}
	}
	return nil
}

func check(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

// -----------------------------------------------------------------------------

type testEnv struct {
	doc *ast.Node
	pkg *types.Package
	ctx *blockCtx
}

func newTestEnv(code string) *testEnv {
	doc, src := parse(code)
	p := gox.NewPackage("", "main", nil)
	ctx := &blockCtx{
		pkg: p, cb: p.CB(), fset: p.Fset, src: src,
		unnameds: make(map[ast.ID]*types.Named),
		typdecls: make(map[string]*gox.TypeDecl),
	}
	ctx.initCTypes()
	return &testEnv{doc: doc, pkg: p.Types, ctx: ctx}
}

// -----------------------------------------------------------------------------
