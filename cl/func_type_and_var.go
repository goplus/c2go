package cl

import (
	"go/types"
	"log"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/types/parser"
)

// -----------------------------------------------------------------------------

func toType(ctx *blockCtx, typ *ast.Type, isParam bool) types.Type {
	t, err := parser.ParseType(ctx, ctx.fset, typ.QualType, isParam)
	if err != nil {
		log.Fatalln("toType:", err)
	}
	return t
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
			id := item.OwnedTagDecl.ID
			if detail, ok := ctx.unnameds[id]; ok {
				compileStructOrUnion(ctx, name, detail)
				return
			}
			log.Fatalln("compileTypedef: unknown id =", id)
		}
	}
	typ := toType(ctx, decl.Type, false)
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
	// TODO:
	ctx.pkg.NewType(name, goNodePos(decl)).InitType(ctx.pkg, types.Typ[types.Int])
}

func compileVar(ctx *blockCtx, decl *ast.Node) {
	if debugCompileDecl {
		log.Println("var", decl.Name, "-", decl.Loc.PresumedLine)
	}
}

// -----------------------------------------------------------------------------
