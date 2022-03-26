package cl

import (
	"log"

	"github.com/goplus/c2go/clang/ast"
)

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
