package cl

import (
	"go/token"
	"go/types"
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
	case ast.GotoStmt:
		compileGotoStmt(ctx, stmt)
	case ast.LabelStmt:
		compileLabelStmt(ctx, stmt)
	case ast.NullStmt:
	default:
		compileExprEx(ctx, stmt, "compileStmt: unknown kind =", flagIgnoreResult)
		ctx.cb.EndStmt()
	}
}

func compileLabelStmt(ctx *blockCtx, stmt *ast.Node) {
	l := ctx.getLabel(goNodePos(stmt), stmt.Name)
	ctx.cb.Label(l)
	compileStmt(ctx, stmt.Inner[0])
}

func compileGotoStmt(ctx *blockCtx, stmt *ast.Node) {
	label := ctx.labelOfGoto(stmt)
	l := ctx.getLabel(goNodePos(stmt), label)
	ctx.cb.Goto(l)
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
	bodyStmts := switchStmt.Inner[1].Inner
	hasCase := false
	for i, n := 0, len(bodyStmts); i < n; i++ {
		stmt := bodyStmts[i]
	retry:
		var idx int
		switch stmt.Kind {
		case ast.CaseStmt:
			idx = 1
		case ast.DefaultStmt:
		default:
			compileStmt(ctx, stmt)
			continue
		}
		if hasCase {
			cb.End()
			hasCase = false
		}
		if idx != 0 {
			compileExpr(ctx, stmt.Inner[0])
		}
		cb.Case(idx)
		switch caseBody := stmt.Inner[idx]; caseBody.Kind {
		case ast.CaseStmt, ast.DefaultStmt:
			cb.Fallthrough().End()
			stmt = caseBody
			goto retry
		default:
			compileStmt(ctx, caseBody)
			hasCase = true
		}
	}
	if hasCase {
		cb.End() // Case
	}
	cb.End() // switch
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
		cb := ctx.cb
		typeCast(cb, getRetType(cb), cb.Get(-1))
	}
	ctx.cb.Return(n, goNode(stmt))
}

func getRetType(cb *gox.CodeBuilder) types.Type {
	return cb.Func().Type().(*types.Signature).Results().At(0).Type()
}

func compileCompoundStmt(ctx *blockCtx, stmts *ast.Node) {
	for _, stmt := range stmts.Inner {
		compileStmt(ctx, stmt)
	}
}

// -----------------------------------------------------------------------------
