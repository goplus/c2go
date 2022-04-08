package cl

import (
	"bytes"
	"go/token"
	"go/types"
	"log"
	"os"
	"strconv"

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

type source struct {
	data []byte
}

type blockCtx struct {
	pkg      *gox.Package
	cb       *gox.CodeBuilder
	fset     *token.FileSet
	tyValist types.Type
	unnameds map[ast.ID]*types.Named
	uncompls map[string]*gox.TypeDecl
	files    map[string]source
	curfile  string
	curfn    *funcCtx
	asuBase  int // anonymous struct/union
}

func (p *blockCtx) getSource(file string) source {
	if v, ok := p.files[file]; ok {
		return v
	}
	b, err := os.ReadFile(file)
	if err != nil {
		log.Panicln("getSource:", err)
	}
	v := source{data: b}
	p.files[file] = v
	return v
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
	src := p.getSource(p.curfile)
	off := v.Range.Begin.Offset
	n := int64(v.Range.Begin.TokLen)
	s := string(src.data[off : off+n])
	if s != "goto" {
		log.Panicln("gotoLabel:", s)
	}
	label := ident(src.data[off+n:], "label not found")
	return label
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

func (p *blockCtx) getAsuName(v *ast.Node, ns string) (string, bool) {
	if name := v.Name; name != "" {
		if v.CompleteDefinition && ns != "" {
			name = ns + "_" + name // TODO: use sth to replace _
		}
		return name, false
	}
	p.asuBase++
	return "_cgoa_" + strconv.Itoa(p.asuBase), true
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
	decl_builtin(pkg)
}

func initValist(scope *types.Scope, pkg *types.Package) types.Type {
	valist := types.NewTypeName(token.NoPos, pkg, "__va_list_tag", nil)
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
