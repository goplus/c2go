package cl

import (
	"bytes"
	"go/format"
	"go/token"
	"go/types"
	"log"
	"os"
	"strconv"
	"sync/atomic"
	"testing"

	goast "go/ast"

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
	SetDebug(DbgFlagAll)
	preprocessor.SetDebug(preprocessor.DbgFlagAll)

	home, err := os.UserHomeDir()
	check(err)

	tmpDir = home + "/.c2go/tmp/"
	err = os.MkdirAll(tmpDir, 0755)
	check(err)
}

func parse(code string, json *[]byte) (doc *ast.Node, src []byte) {
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

	doc, _, err = parser.ParseFileEx(outfile, 0, &parser.Config{Json: json})
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
	doc  *ast.Node
	pkg  *gox.Package
	ctx  *blockCtx
	json []byte
}

func newTestEnv(code string) *testEnv {
	var json []byte
	doc, src := parse(code, &json)
	p := gox.NewPackage("", "main", nil)
	ctx := &blockCtx{
		pkg: p, cb: p.CB(), fset: p.Fset, src: src,
		unnameds: make(map[ast.ID]unnamedType),
	}
	ctx.typdecls = make(map[string]*gox.TypeDecl)
	ctx.initCTypes()
	return &testEnv{doc: doc, pkg: p, ctx: ctx, json: json}
}

// -----------------------------------------------------------------------------

func findFunc(file *goast.File, name string) *goast.FuncDecl {
	for _, decl := range file.Decls {
		switch v := decl.(type) {
		case *goast.FuncDecl:
			if v.Name.Name == name {
				return v
			}
		}
	}
	return nil
}

func testFunc(t *testing.T, name string, code string, outFunc string) Package {
	return testWith(t, name, "test", code, outFunc, nil)
}

func testFuncEx(t *testing.T, name string, code string, outFunc string, initConf func(*Config)) Package {
	return testWith(t, name, "test", code, outFunc, initConf)
}

func testWith(t *testing.T, name string, fn string, code string, outFunc string, initConf func(*Config)) (pkgOut Package) {
	t.Run(name, func(t *testing.T) {
		var json []byte
		doc, src := parse(code, &json)
		conf := &Config{Src: src, NeedPkgInfo: true}
		if initConf != nil {
			initConf(conf)
		}
		pkg, err := NewPackage("", "main", doc, conf)
		check(err)
		file := gox.ASTFile(pkg.Package)
		ret := goast.Node(file)
		if fn != "" {
			ret = findFunc(file, fn)
		}
		w := bytes.NewBuffer(nil)
		err = format.Node(w, pkg.Fset, ret)
		check(err)
		if out := w.String(); out != outFunc {
			t.Fatalf(
				"==> Result:\n%s\n==> Expected:\n%s\n==> AST:\n%s\n",
				out, outFunc, string(json))
		}
		pkgOut = pkg
	})
	return
}

func testPanic(t *testing.T, panicMsg string, doPanic func()) {
	t.Run(panicMsg, func(t *testing.T) {
		if panicMsg != "" {
			defer func() {
				if e := recover(); e == nil {
					t.Fatal("testPanic: no error?")
				} else if msg := e.(string); msg != panicMsg {
					t.Fatalf("\nResult:\n%s\nExpected Panic:\n%s\n", msg, panicMsg)
				}
			}()
		}
		doPanic()
	})
}

// -----------------------------------------------------------------------------

func TestGetVStruct(t *testing.T) {
	defer func() {
		if e := recover(); e != nil {
			if emsg, ok := e.(string); ok {
				if emsg == "foo not found\n" {
					return
				}
			}
			t.Fatal("TestGetVStruct failed:", e)
		}
		t.Fatal("TestGetVStruct: no error?")
	}()
	var ctx blockCtx
	typ := types.NewNamed(types.NewTypeName(0, nil, "foo", nil), nil, nil)
	ctx.getVStruct(typ)
}

func TestFuncAndDecl(t *testing.T) {
	testFunc(t, "testKeyword", `
void test(int var) {
}
`, `func test(var_ int32) {
}`)
}

func TestParseEnv(t *testing.T) {
	tyI128 := types.Typ[types.Complex128]
	tyU128 := types.Typ[types.Complex64]
	ctx := &blockCtx{tyI128: tyI128, tyU128: tyU128}
	if ctx.Uint128() != tyU128 {
		t.Fatal("ctx.Uint128")
	}
	if ctx.Int128() != tyI128 {
		t.Fatal("ctx.Int128")
	}
}

// -----------------------------------------------------------------------------

func TestNodeInterp(t *testing.T) {
	fset := token.NewFileSet()
	ctx := &blockCtx{
		fset: fset,
		src: []byte(`
void test(int var) {
}
`)}
	ctx.file = fset.AddFile(ctx.srcfile, -1, 1<<30)
	interp := &nodeInterp{fset: fset}
	base := token.Pos(ctx.file.Base())
	v := &node{ctx: ctx, pos: base, end: base}
	src, pos := interp.LoadExpr(v)
	if src != "" || pos.String() != "1:1" {
		t.Fatal("interp.LoadExpr:", src, pos)
	}
	interp.Caller(v)
	if ret := interp.Position(v.Pos()); ret != pos {
		t.Fatal("interp.Position:", ret, "expected:", pos)
	}
}

func TestComplicatedFor(t *testing.T) {
	testFunc(t, "testFor1", `
void printf(int);
void test() {
    int i = 10;
    if (i < 20) {
       goto resume;
    }
    for (i=5;i>0;i--) {
resume:
        printf(i);
    };
}`, `func test() {
	var
	var i int32 = int32(10)
	if !(i < int32(20)) {
		goto _cgol_1
	}
	goto resume
_cgol_1:
	i = int32(5)
_cgol_2:
	if !(i > int32(0)) {
		goto _cgol_3
	}
resume:
	printf(i)
	i--
	goto _cgol_2
_cgol_3:
}`)
	testFunc(t, "testFor2", `
void printf(int);
void test() {
    int i = 10;
    if (i < 20) {
       goto resume;
    }
    for (;;) {
       i--;
       if (i<0) {
          break;
       }
resume:
       printf(i);
    };
}`, `func test() {
	var
	var i int32 = int32(10)
	if !(i < int32(20)) {
		goto _cgol_1
	}
	goto resume
_cgol_1:
	;
_cgol_2:
	i--
	if i < int32(0) {
		goto _cgol_3
	}
resume:
	printf(i)
	goto _cgol_2
_cgol_3:
}`)
}

func TestSimpleForStmt(t *testing.T) {
	testFunc(t, "testSimpleFor1", `
void printf(int,int);
void test() {
    int j = 0;
    for (int i = 0; i < 5; ++i, j += 7)
    {
        printf(i,j);
    }
}
`, `func test() {
	var j int32 = int32(0)
	for i := int32(int32(0)); i < int32(5); func() int32 {
		i++
		return func() (_cgo_ret int32) {
			_cgo_addr := &j
			*_cgo_addr += int32(7)
			return *_cgo_addr
		}()
	}() {
		printf(i, j)
	}
}`)
	testFunc(t, "testSimpleFor2", `
void printf(int,int);
void test() {
    for (int i = 0, j = 0; i < 5; ++i, j += 7)
    {
        printf(i,j);
    }
}
`, `func test() {
	{
		var i int32 = int32(0)
		var j int32 = int32(0)
		for ; i < int32(5); func() int32 {
			i++
			return func() (_cgo_ret int32) {
				_cgo_addr := &j
				*_cgo_addr += int32(7)
				return *_cgo_addr
			}()
		}() {
			printf(i, j)
		}
	}
}`)
}

// -----------------------------------------------------------------------------
