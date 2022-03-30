package cl

import (
	"log"

	"github.com/goplus/c2go/clang/ast"
)

// -----------------------------------------------------------------------------

func compileStmt(ctx *blockCtx, stmt *ast.Node) {
	switch stmt.Kind {
	case ast.IfStmt:
		compileIfStmt(ctx, stmt)
	case ast.ReturnStmt:
		compileReturnStmt(ctx, stmt)
	case ast.DeclStmt:
		compileDeclStmt(ctx, stmt)
	default:
		compileExprEx(ctx, stmt, "compileStmt: unknown kind =", false)
		ctx.cb.EndStmt()
	}
}

func compileDeclStmt(ctx *blockCtx, stmt *ast.Node) {
	for _, item := range stmt.Inner {
		switch item.Kind {
		case ast.VarDecl:
			compileVar(ctx, item, false)
		default:
			log.Fatalln("compileDeclStmt: unknown kind =", item.Kind)
		}
	}
}

func compileIfStmt(ctx *blockCtx, stmt *ast.Node) {
	cb := ctx.cb
	cb.If()
	compileExpr(ctx, stmt.Inner[0])
	cb.Then()
	compileStmt(ctx, stmt.Inner[1])
	if stmt.HasElse {
		cb.Else()
		compileStmt(ctx, stmt.Inner[2])
	}
	cb.End()
}

func compileReturnStmt(ctx *blockCtx, stmt *ast.Node) {
	n := len(stmt.Inner)
	if n > 0 {
		n = 1
		compileExpr(ctx, stmt.Inner[0])
	}
	ctx.cb.Return(n, goNode(stmt))
}

func compileCompoundStmt(ctx *blockCtx, stmts *ast.Node) {
	for _, stmt := range stmts.Inner {
		compileStmt(ctx, stmt)
	}
}

// -----------------------------------------------------------------------------
