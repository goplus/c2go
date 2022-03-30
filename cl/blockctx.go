package cl

import (
	"go/token"
	"go/types"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/gox"

	ctypes "github.com/goplus/c2go/clang/types"
)

// -----------------------------------------------------------------------------

type blockCtx struct {
	pkg      *gox.Package
	cb       *gox.CodeBuilder
	fset     *token.FileSet
	unnameds map[ast.ID]*ast.Node
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

func initValist(scope *types.Scope, pkg *types.Package) {
	valist := types.NewTypeName(token.NoPos, pkg, "__va_list_tag", nil)
	t := types.NewNamed(valist, types.Typ[types.Int8], nil)
	scope.Insert(valist)
	aliasType(scope, pkg, "__builtin_va_list", types.NewPointer(t))
}

func aliasType(scope *types.Scope, pkg *types.Package, name string, typ types.Type) {
	o := types.NewTypeName(token.NoPos, pkg, name, typ)
	scope.Insert(o)
}

func initCTypes(pkg *types.Package) {
	scope := types.Universe
	initValist(scope, pkg)
	aliasType(scope, pkg, "char", types.Typ[types.Int8])
	aliasType(scope, pkg, "void", ctypes.Void)
	aliasType(scope, pkg, "__int128", ctypes.Int128)
}

// -----------------------------------------------------------------------------
