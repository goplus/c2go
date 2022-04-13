package cl

import (
	"bytes"
	"go/token"
	"go/types"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/gox"
	"github.com/qiniu/x/ctype"

	ctypes "github.com/goplus/c2go/clang/types"
)

// -----------------------------------------------------------------------------

type funcCtx struct {
	labels map[string]*gox.Label
}

func newFuncCtx() *funcCtx {
	return &funcCtx{
		labels: make(map[string]*gox.Label),
	}
}

// -----------------------------------------------------------------------------

type blockCtx struct {
	pkg      *gox.Package
	cb       *gox.CodeBuilder
	fset     *token.FileSet
	tyValist types.Type
	unnameds map[ast.ID]*types.Named
	typdecls map[string]*gox.TypeDecl
	srcfile  string
	cursrc   []byte
	curfn    *funcCtx
	asuBase  int // anonymous struct/union
}

func (p *blockCtx) getSource() []byte {
	if v := p.cursrc; v != nil {
		return v
	}
	b, err := os.ReadFile(p.srcfile)
	if err != nil {
		log.Panicln("getSource:", err)
	}
	p.cursrc = b
	return b
}

func (p *blockCtx) getLabel(pos token.Pos, name string) *gox.Label {
	if fn := p.curfn; fn != nil {
		l, ok := fn.labels[name]
		if !ok {
			l = p.cb.NewLabel(pos, name)
			fn.labels[name] = l
		}
		return l
	}
	log.Panicln("can't use label out of func")
	return nil
}

func (p *blockCtx) labelOfGoto(v *ast.Node) string {
	src := p.getSource()
	off := v.Range.Begin.Offset
	n := int64(v.Range.Begin.TokLen)
	op := string(src[off : off+n])
	if op != "goto" {
		log.Panicln("gotoOp:", op)
	}
	label := ident(src[off+n:], "label not found")
	return label
}

func (p *blockCtx) typeOfSizeof(v *ast.Node) string {
	src := p.getSource()
	off := v.Range.Begin.Offset
	n := int64(v.Range.Begin.TokLen)
	op := string(src[off : off+n])
	if op != "sizeof" {
		log.Panicln("sizeofOp:", op)
	}
	typ := string(src[off+n : v.Range.End.Offset])
	return strings.TrimPrefix(strings.TrimLeft(typ, " \t\r\n"), "(")
}

func (p *blockCtx) getInstr(v *ast.Node) string {
	src := p.getSource()
	off := v.Range.Begin.Offset
	n := int64(v.Range.Begin.TokLen)
	return string(src[off : off+n])
}

func ident(b []byte, msg string) string {
	b = bytes.TrimLeft(b, " \t\r\n")
	idx := bytes.IndexFunc(b, func(r rune) bool {
		return !ctype.Is(ctype.CSYMBOL_NEXT_CHAR, r)
	})
	if idx <= 0 {
		log.Panicln(msg)
	}
	return string(b[:idx])
}

func (p *blockCtx) sizeof(typ types.Type) int {
	return int(p.pkg.Sizeof(typ))
}

const (
	suNormal = iota
	suAnonymous
	suNested
)

func (p *blockCtx) getSuName(v *ast.Node, ns, tag string) (string, int) {
	if name := v.Name; name != "" {
		if v.CompleteDefinition && ns != "" {
			return ns + "_" + name, suNested // TODO: use sth to replace _
		}
		return ctypes.MangledName(tag, name), suNormal
	}
	p.asuBase++
	return "_cgoa_" + strconv.Itoa(p.asuBase), suAnonymous
}

func (p *blockCtx) initCTypes() {
	pkg := p.pkg.Types
	scope := pkg.Scope()
	p.tyValist = initValist(scope, pkg)
	aliasType(scope, pkg, "char", types.Typ[types.Int8])
	aliasType(scope, pkg, "void", ctypes.Void)
	aliasType(scope, pkg, "float", types.Typ[types.Float32])
	aliasType(scope, pkg, "double", types.Typ[types.Float64])
	aliasType(scope, pkg, "__int128", ctypes.Int128)
	decl_builtin(p)
}

func initValist(scope *types.Scope, pkg *types.Package) types.Type {
	valist := types.NewTypeName(token.NoPos, pkg, ctypes.MangledName("struct", "__va_list_tag"), nil)
	t := types.NewNamed(valist, types.Typ[types.Int8], nil)
	scope.Insert(valist)
	tyValist := types.NewPointer(t)
	aliasType(scope, pkg, "__builtin_va_list", tyValist)
	return tyValist
}

func aliasType(scope *types.Scope, pkg *types.Package, name string, typ types.Type) {
	o := types.NewTypeName(token.NoPos, pkg, name, typ)
	scope.Insert(o)
}

// -----------------------------------------------------------------------------
