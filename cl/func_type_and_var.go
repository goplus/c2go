package cl

import (
	"go/types"
	"log"

	"github.com/goplus/c2go/clang/ast"
)

// -----------------------------------------------------------------------------

func toType(ctx *blockCtx, typ *ast.Type) types.Type {
	log.Fatalln("toType:", typ.QualType)
	return nil
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
