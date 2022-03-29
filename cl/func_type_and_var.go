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
	if debugCompileDecl {
		log.Println("typedef", decl.Name, "-", decl.Type.QualType, decl.Loc.PresumedLine)
	}
}

func compileStructOrUnion(ctx *blockCtx, decl *ast.Node) {
	if debugCompileDecl {
		log.Println(decl.TagUsed, decl.Name, "-", decl.Loc.PresumedLine)
	}
}

func compileVar(ctx *blockCtx, decl *ast.Node) {
	if debugCompileDecl {
		log.Println("var", decl.Name, "-", decl.Loc.PresumedLine)
	}
}

// -----------------------------------------------------------------------------
