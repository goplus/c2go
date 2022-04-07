package cl

import (
	"go/token"
	"log"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

func compileStmt(ctx *blockCtx, stmt *ast.Node) {
	switch stmt.Kind {
	case ast.IfStmt:
		compileIfStmt(ctx, stmt)
	case ast.ForStmt:
		compileForStmt(ctx, stmt)
	case ast.SwitchStmt:
		compileSwitchStmt(ctx, stmt)
	case ast.WhileStmt:
		compileWhileStmt(ctx, stmt)
	case ast.DoStmt:
		compileDoStmt(ctx, stmt)
	case ast.ReturnStmt:
		compileReturnStmt(ctx, stmt)
	case ast.ContinueStmt:
		ctx.cb.Continue(nil)
	case ast.BreakStmt:
		ctx.cb.Break(nil)
	case ast.DeclStmt:
		compileDeclStmt(ctx, stmt, false)
	case ast.CompoundStmt:
		compileCompoundStmt(ctx, stmt)
	case ast.NullStmt:
	default:
		compileExprEx(ctx, stmt, "compileStmt: unknown kind =", flagIgnoreResult)
		ctx.cb.EndStmt()
	}
}

func compileDoStmt(ctx *blockCtx, stmt *ast.Node) {
	cb := ctx.cb.For().None().Then()
	{
		compileStmt(ctx, stmt.Inner[0])
		cb.If()
		compileExpr(ctx, stmt.Inner[1])
		castToBoolExpr(cb)
		cb.UnaryOp(token.NOT).Then().
			Break(nil).
			End().
			End()
	}
}

func compileWhileStmt(ctx *blockCtx, stmt *ast.Node) {
	cb := ctx.cb.For()
	compileExpr(ctx, stmt.Inner[0])
	castToBoolExpr(cb)
	cb.Then()
	compileStmt(ctx, stmt.Inner[1])
	cb.End()
}

func compileForStmt(ctx *blockCtx, stmt *ast.Node) {
	cb := ctx.cb.For()
	if initStmt := stmt.Inner[0]; initStmt.Kind != "" {
		compileStmt(ctx, initStmt)
	}
	if stmt := stmt.Inner[1]; stmt.Kind != "" {
		log.Panicln("compileForStmt: unexpected -", stmt.Kind)
	}
	if cond := stmt.Inner[2]; cond.Kind != "" {
		compileExpr(ctx, cond)
		castToBoolExpr(cb)
	} else {
		cb.None()
	}
	cb.Then()
	compileStmt(ctx, stmt.Inner[4])
	if postStmt := stmt.Inner[3]; postStmt.Kind != "" {
		cb.Post()
		compileStmt(ctx, postStmt)
	}
	cb.End()
}

func compileSwitchStmt(ctx *blockCtx, switchStmt *ast.Node) {
	cb := ctx.cb.Switch()
	compileExpr(ctx, switchStmt.Inner[0])
	cb.Then()
	switchBody := switchStmt.Inner[1]
	for _, caseStmt := range switchBody.Inner {
		switch caseStmt.Kind {
		case ast.CaseStmt, ast.DefaultStmt:
			caseBody := compileCaseCond(ctx, cb, caseStmt)
			for _, stmt := range caseBody {
				compileStmt(ctx, stmt)
			}
			cb.End()
		case ast.BreakStmt:
		default:
			log.Panicln("compileSwitchStmt: unknown case kind =", caseStmt.Kind)
		}
	}
	cb.End() // switch
}

func compileCaseCond(ctx *blockCtx, cb *gox.CodeBuilder, caseStmt *ast.Node) (body []*ast.Node) {
	var idx int
	if caseStmt.Kind == ast.CaseStmt {
		idx = 1
		compileExpr(ctx, caseStmt.Inner[0])
		cb.Case(1)
	} else {
		cb.Case(0)
	}
	if len(caseStmt.Inner) > 1 {
		switch v := caseStmt.Inner[idx]; v.Kind {
		case ast.CaseStmt, ast.DefaultStmt:
			cb.Fallthrough().End()
			return compileCaseCond(ctx, cb, v)
		}
	}
	return caseStmt.Inner[idx:]
}

func compileIfStmt(ctx *blockCtx, stmt *ast.Node) {
	cb := ctx.cb
	cb.If()
	compileExpr(ctx, stmt.Inner[0])
	castToBoolExpr(cb)
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
