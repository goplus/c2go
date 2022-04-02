package cl

import (
	"go/token"
	"go/types"
	"strconv"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/gox"

	ctypes "github.com/goplus/c2go/clang/types"
)

// -----------------------------------------------------------------------------

type blockCtx struct {
	pkg      *gox.Package
	cb       *gox.CodeBuilder
	fset     *token.FileSet
	tyValist types.Type
	unnameds map[ast.ID]*ast.Node
	asuBase  int
}

func (p *blockCtx) Pkg() *types.Package {
	return p.pkg.Types
}

func (p *blockCtx) LookupType(typ string) (t types.Type, err error) {
	_, o := p.cb.Scope().LookupParent(typ, token.NoPos)
	if o != nil {
		return o.Type(), nil
	}
	return nil, ctypes.ErrNotFound
}

func (p *blockCtx) sizeof(typ types.Type) int {
	return int(p.pkg.Sizeof(typ))
}

func (p *blockCtx) getAsuName(v *ast.Node) string {
	if name := v.Name; name != "" {
		return name
	}
	p.asuBase++
	return "Gopasu_" + strconv.Itoa(p.asuBase)
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
