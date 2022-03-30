package cl

import (
	"go/token"
	"go/types"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/types/parser"
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
	switch typ {
	case "int":
		return types.Typ[types.Int], nil
	case "char":
		return types.Typ[types.Int8], nil
	case "void":
		return ctypes.Void, nil
	case "__int128":
		return ctypes.Int128, nil
	case "__builtin_va_list", "__va_list_tag":
		return ctypes.NotImpl, nil
	}
	_, o := p.cb.Scope().LookupParent(typ, token.NoPos)
	if o != nil {
		return o.Type(), nil
	}
	return nil, parser.ErrTypeNotFound
}

// -----------------------------------------------------------------------------
