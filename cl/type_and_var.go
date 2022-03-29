package cl

import (
	"go/types"
	"log"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/types/parser"
)

// -----------------------------------------------------------------------------

func toType(ctx *blockCtx, typ *ast.Type, flags int) types.Type {
	t, err := parser.ParseType(ctx, ctx.fset, typ.QualType, flags)
	if err != nil {
		log.Fatalln("toType:", err)
	}
	return t
}

func toStructType(ctx *blockCtx, decl *ast.Node) *types.Struct {
	fields := make([]*types.Var, 0, len(decl.Inner))
	for _, item := range decl.Inner {
		switch item.Kind {
		case ast.FieldDecl:
			if debugCompileDecl {
				log.Println("  => field", item.Name, "-", item.Type.QualType)
			}
			fld := newField(ctx, item)
			fields = append(fields, fld)
		default:
			log.Fatalln("toStructType: unknown field kind =", item.Kind)
		}
	}
	return types.NewStruct(fields, nil)
}

func toUnionType(ctx *blockCtx, decl *ast.Node) types.Type {
	// TODO: union
	return parser.TyNotImpl
}

func newField(ctx *blockCtx, decl *ast.Node) *types.Var {
	typ := toType(ctx, decl.Type, 0)
	return types.NewField(goNodePos(decl), ctx.pkg.Types, decl.Name, typ, false)
}

// -----------------------------------------------------------------------------

func compileTypedef(ctx *blockCtx, decl *ast.Node) {
	name, qualType := decl.Name, decl.Type.QualType
	if debugCompileDecl {
		log.Println("typedef", name, "-", qualType, decl.Loc.PresumedLine)
	}
	if len(decl.Inner) > 0 {
		item := decl.Inner[0]
		if item.Kind == "ElaboratedType" {
			if owned := item.OwnedTagDecl; owned != nil && owned.Name == "" {
				id := owned.ID
				if detail, ok := ctx.unnameds[id]; ok {
					compileStructOrUnion(ctx, name, detail)
					return
				}
				log.Fatalln("compileTypedef: unknown id =", id)
			}
		}
	}
	typ := toType(ctx, decl.Type, 0)
	ctx.pkg.AliasType(name, typ, goNodePos(decl))
}

func compileStructOrUnion(ctx *blockCtx, name string, decl *ast.Node) {
	if debugCompileDecl {
		log.Println(decl.TagUsed, name, "-", decl.Loc.PresumedLine)
	}
	if name == "" {
		ctx.unnameds[decl.ID] = decl
		return
	}
	var inner types.Type
	var pkg = ctx.pkg
	var t = pkg.NewType(name, goNodePos(decl))
	switch decl.TagUsed {
	case "struct":
		inner = toStructType(ctx, decl)
	default:
		inner = toUnionType(ctx, decl)
	}
	t.InitType(pkg, inner)
}

func compileVar(ctx *blockCtx, decl *ast.Node) {
	if debugCompileDecl {
		log.Println("var", decl.Name, "-", decl.Loc.PresumedLine)
	}
}

// -----------------------------------------------------------------------------
