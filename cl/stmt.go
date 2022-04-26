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
	case ast.BreakStmt:
		compileBreakStmt(ctx, stmt)
	case ast.ContinueStmt:
		compileContinueStmt(ctx, stmt)
	case ast.DeclStmt:
		compileDeclStmt(ctx, stmt, false)
	case ast.CompoundStmt:
		compileCompoundStmt(ctx, stmt)
	case ast.GotoStmt:
		compileGotoStmt(ctx, stmt)
	case ast.LabelStmt:
		compileLabelStmt(ctx, stmt)
	case ast.CaseStmt, ast.DefaultStmt:
		compileCaseStmt(ctx, stmt)
	case ast.NullStmt:
	case ast.GCCAsmStmt:
		// TODO: skip asm
	default:
		compileExprEx(ctx, stmt, "compileStmt: unknown kind =", flagIgnoreResult)
		ctx.cb.EndStmt()
	}
}

// -----------------------------------------------------------------------------

func compileDoStmt(ctx *blockCtx, stmt *ast.Node) {
	flow := ctx.enterFlow()
	defer ctx.leave(flow)

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

// -----------------------------------------------------------------------------

func compileWhileStmt(ctx *blockCtx, stmt *ast.Node) {
	if stmt.Complicated {
		compileComplicatedWhileStmt(ctx, stmt)
		return
	}
	compileSimpleWhileStmt(ctx, stmt)
}

func compileComplicatedWhileStmt(ctx *blockCtx, stmt *ast.Node) {
	loop := ctx.enterLoop()
	defer ctx.leave(loop)

	loop.labelStart(ctx)

	cb := ctx.cb.If()
	compileExpr(ctx, stmt.Inner[0])
	castToBoolExpr(cb)
	done := loop.EndLabel(ctx)
	cb.UnaryOp(token.NOT).Then().Goto(done).End()

	compileStmt(ctx, stmt.Inner[1])
	cb.Goto(loop.start).Label(done)
}

func compileSimpleWhileStmt(ctx *blockCtx, stmt *ast.Node) {
	flow := ctx.enterFlow()
	defer ctx.leave(flow)

	cb := ctx.cb.For()
	compileExpr(ctx, stmt.Inner[0])
	castToBoolExpr(cb)
	cb.Then()
	compileStmt(ctx, stmt.Inner[1])
	cb.End()
}

// -----------------------------------------------------------------------------

func compileForStmt(ctx *blockCtx, stmt *ast.Node) {
	flow := ctx.enterFlow()
	defer ctx.leave(flow)

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

// -----------------------------------------------------------------------------

func compileContinueStmt(ctx *blockCtx, stmt *ast.Node) {
	if l := ctx.curflow.ContinueLabel(ctx); l != nil {
		ctx.cb.Goto(l)
		return
	}
	ctx.cb.Continue(nil)
}

func compileBreakStmt(ctx *blockCtx, stmt *ast.Node) {
	if l := ctx.curflow.EndLabel(ctx); l != nil {
		ctx.cb.Goto(l)
		return
	}
	ctx.cb.Break(nil)
}

// -----------------------------------------------------------------------------

func compileSwitchStmt(ctx *blockCtx, switchStmt *ast.Node) {
	if switchStmt.Complicated {
		compileComplicatedSwitchStmt(ctx, switchStmt)
		return
	}
	compileSimpleSwitchStmt(ctx, switchStmt)
}

func compileComplicatedSwitchStmt(ctx *blockCtx, switchStmt *ast.Node) {
	sw := ctx.enterSwitch()
	defer ctx.leave(sw)

	const (
		tagName        = "_cgo_tag"
		notMatchedName = "_cgo_nm"
	)
	cb := ctx.cb.DefineVarStart(token.NoPos, notMatchedName, tagName).Val(true)
	compileExpr(ctx, switchStmt.Inner[0])
	cb.EndInit(2)

	scope := cb.Scope()
	sw.notmat = scope.Lookup(notMatchedName)
	sw.tag = scope.Lookup(tagName)

	body := switchStmt.Inner[1]
	if firstStmtNotCase(body) {
		l := sw.nextCaseLabel(ctx)
		cb.Goto(l)
	}
	compileStmt(ctx, body)
	done := sw.EndLabel(ctx)
	cb.Goto(done)
	if sw.next != nil {
		cb.Label(sw.next)
	}
	if sw.defau != nil {
		cb.Goto(sw.defau)
	}
	cb.Label(done)
}

func compileCaseStmt(ctx *blockCtx, stmt *ast.Node) {
	isCaseStmt := stmt.Kind == ast.CaseStmt
	cb := ctx.cb
	sw := ctx.getSwitchCtx()
	if sw == nil {
		log.Panicln("compileCaseStmt: case stmt isn't in switch")
	}
	var idx int
	if isCaseStmt {
		if sw.next != nil {
			cb.Label(sw.next)
		}
		cb.If().Val(sw.notmat).Val(sw.tag)
		compileExpr(ctx, stmt.Inner[0])
		cb.BinaryOp(token.NEQ).BinaryOp(token.LAND).Then()
		l := sw.nextCaseLabel(ctx)
		cb.Goto(l).End()
		idx = 1
	} else {
		sw.labelDefault(ctx)
	}
	cb.VarRef(sw.notmat).Val(false).Assign(1)
	compileStmt(ctx, stmt.Inner[idx])
}

func firstStmtNotCase(body *ast.Node) bool {
	if body.Kind != ast.CompoundStmt || len(body.Inner) == 0 {
		return true
	}
	switch body.Inner[0].Kind {
	case ast.CaseStmt, ast.DefaultStmt:
		return false
	}
	return true
}

func compileSimpleSwitchStmt(ctx *blockCtx, switchStmt *ast.Node) {
	flow := ctx.enterFlow()
	defer ctx.leave(flow)

	cb := ctx.cb.Switch()
	compileExpr(ctx, switchStmt.Inner[0])
	cb.Then()
	body := switchStmt.Inner[1]
	if body.Kind != ast.CompoundStmt {
		log.Panicln("compileSimpleSwitchStmt: not a simple switch stmt")
	}
	bodyStmts := body.Inner
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
		cb.End() // case
	}
	cb.End() // switch
}

// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------

func compileCompoundStmt(ctx *blockCtx, stmts *ast.Node) {
	for _, stmt := range stmts.Inner {
		compileStmt(ctx, stmt)
	}
}

// -----------------------------------------------------------------------------
