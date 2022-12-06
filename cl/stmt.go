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

func compileSub(ctx *blockCtx, stmt *ast.Node) {
	switch stmt.Kind {
	case ast.CompoundStmt:
		for _, item := range stmt.Inner {
			compileStmt(ctx, item)
		}
		return
	}
	compileStmt(ctx, stmt)
}

func compileDoStmt(ctx *blockCtx, stmt *ast.Node) {
	if stmt.Complicated {
		compileComplicatedDoStmt(ctx, stmt)
		return
	}

	flow := ctx.enterFlow(flowKindLoop)
	defer ctx.leave(flow)

	cb := ctx.cb.For().None().Then()
	{
		compileSub(ctx, stmt.Inner[0])
		cb.If()
		compileExpr(ctx, stmt.Inner[1])
		castToBoolExpr(cb)
		cb.UnaryOp(token.NOT).Then().
			Break(nil).
			End().
			End()
	}
}

func compileComplicatedDoStmt(ctx *blockCtx, stmt *ast.Node) {
	loop := ctx.enterLoop()
	defer ctx.leave(loop)

	loop.labelStart(ctx)
	compileSub(ctx, stmt.Inner[0])

	cb := ctx.cb.If()
	compileExpr(ctx, stmt.Inner[1])
	castToBoolExpr(cb)
	cb.Then().Goto(loop.start).End()

	if loop.done != nil {
		cb.Label(loop.done)
	}
}

func checkNeedReturn(ctx *blockCtx, body *ast.Node) {
	n := len(body.Inner)
	if n > 0 {
		last := body.Inner[n-1]
		if last.Kind == ast.ReturnStmt {
			return
		}
	}
	cb := ctx.cb
	if ret, ok := getRetTypeEx(cb); ok {
		cb.ZeroLit(ret).Return(1)
	}
}

// -----------------------------------------------------------------------------

func compileWhileStmt(ctx *blockCtx, stmt *ast.Node) {
	if stmt.Complicated {
		compileComplicatedWhileStmt(ctx, stmt)
		return
	}

	flow := ctx.enterFlow(flowKindLoop)
	defer ctx.leave(flow)

	cb := ctx.cb.For()
	compileExpr(ctx, stmt.Inner[0])
	castToBoolExpr(cb)
	cb.Then()
	compileSub(ctx, stmt.Inner[1])
	cb.End()
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

	compileSub(ctx, stmt.Inner[1])
	cb.Goto(loop.start).Label(done)
}

// -----------------------------------------------------------------------------

func compileComplicatedIfStmt(ctx *blockCtx, stmt *ast.Node) {
	flow := ctx.enterIf()
	defer ctx.leave(flow)

	var done = flow.EndLabel(ctx)
	var label *gox.Label
	if stmt.HasElse {
		label = flow.elseLabel(ctx)
	} else {
		label = done
	}

	cb := ctx.cb.If()
	compileExpr(ctx, stmt.Inner[0])
	castToBoolExpr(cb)
	cb.UnaryOp(token.NOT).Then().Goto(label).End()
	compileSub(ctx, stmt.Inner[1])

	if stmt.HasElse {
		cb.Goto(done).Label(label)
		compileSub(ctx, stmt.Inner[2])
	}
	cb.Label(done)
}

func compileIfStmt(ctx *blockCtx, stmt *ast.Node) {
	if stmt.Complicated {
		compileComplicatedIfStmt(ctx, stmt)
		return
	}

	flow := ctx.enterFlow(flowKindIf)
	defer ctx.leave(flow)

	cb := ctx.cb.If()
	compileExpr(ctx, stmt.Inner[0])
	castToBoolExpr(cb)
	cb.Then()
	compileSub(ctx, stmt.Inner[1])
	if stmt.HasElse {
		cb.Else()
		compileSub(ctx, stmt.Inner[2])
	}
	cb.End()
}

// -----------------------------------------------------------------------------

func compileCompoundStmt(ctx *blockCtx, cStmt *ast.Node) {
	cb := ctx.cb
	if cStmt.Complicated {
		cb.VBlock()
	} else {
		cb.Block()
	}
	for _, stmt := range cStmt.Inner {
		compileStmt(ctx, stmt)
	}
	cb.End()
}

// -----------------------------------------------------------------------------

func compileComplicatedForStmt(ctx *blockCtx, stmt *ast.Node) {
	loop := ctx.enterLoop()
	defer ctx.leave(loop)

	compileInitStmt(ctx, stmt.Inner[0])
	if stmt := stmt.Inner[1]; stmt.Kind != "" {
		log.Panicln("compileForStmt: unexpected -", stmt.Kind)
	}

	loop.labelStart(ctx)
	done := loop.EndLabel(ctx)

	cb := ctx.cb
	if cond := stmt.Inner[2]; cond.Kind != "" {
		cb = cb.If()
		compileExpr(ctx, stmt.Inner[2])
		castToBoolExpr(cb)
		cb.UnaryOp(token.NOT).Then().Goto(done).End()
	}
	compileSub(ctx, stmt.Inner[4])
	if postStmt := stmt.Inner[3]; postStmt.Kind != "" {
		compileStmt(ctx, postStmt)
	}
	cb.Goto(loop.start).Label(done)
}

func compileForStmt(ctx *blockCtx, stmt *ast.Node) {
	if stmt.Complicated {
		compileComplicatedForStmt(ctx, stmt)
		return
	}

	flow := ctx.enterFlow(flowKindLoop)
	defer ctx.leave(flow)

	cb := ctx.cb
	if hasMultiInitDeclStmt(ctx, stmt.Inner[0]) {
		cb = ctx.cb.Block()
		defer cb.End()
		compileInitStmt(ctx, stmt.Inner[0])
		cb = cb.For()
	} else {
		cb = cb.For()
		compileInitStmt(ctx, stmt.Inner[0])
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
	compileSub(ctx, stmt.Inner[4])
	if postStmt := stmt.Inner[3]; postStmt.Kind != "" {
		cb.Post()
		compileStmt(ctx, postStmt)
	}
	cb.End()
}

func compileInitStmt(ctx *blockCtx, initStmt *ast.Node) {
	switch initStmt.Kind {
	case "":
	case ast.DeclStmt:
		if inner := initStmt.Inner; len(inner) == 1 {
			if stmt := inner[0]; stmt.Kind == ast.VarDecl {
				compileVarDef(ctx, stmt)
				return
			}
		}
		fallthrough
	default:
		compileStmt(ctx, initStmt)
	}
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
	const (
		tagNamePrefix        = "_tag"
		notMatchedNamePrefix = "_nm"
	)

	sw := ctx.enterSwitch()
	defer ctx.leave(sw)

	compileExpr(ctx, switchStmt.Inner[0])
	tag := ctx.cb.InternalStack().Pop()

	scope := ctx.cb.Scope()
	ctx.newVar(scope, token.NoPos, tag.Type, tagNamePrefix)
	ctx.newVar(scope, token.NoPos, types.Typ[types.Bool], notMatchedNamePrefix)

	sw.tag = gox.Lookup(scope, tagNamePrefix)
	sw.notmat = gox.Lookup(scope, notMatchedNamePrefix)

	cb := ctx.cb.VarRef(sw.tag).VarRef(sw.notmat).Val(tag).Val(true).Assign(2)

	body := switchStmt.Inner[1]
	if firstStmtNotCase(body) {
		l := sw.nextCaseLabel(ctx)
		cb.Goto(l)
	}
	compileSub(ctx, body)
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
	flow := ctx.enterFlow(flowKindSwitch)
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

func compileLabelStmt(ctx *blockCtx, stmt *ast.Node) {
	l := ctx.getLabel(ctx.goNodePos(stmt), stmt.Name)
	ctx.cb.Label(l)
	compileStmt(ctx, stmt.Inner[0])
}

func compileGotoStmt(ctx *blockCtx, stmt *ast.Node) {
	label := ctx.labelOfGoto(stmt)
	l := ctx.getLabel(ctx.goNodePos(stmt), label)
	ctx.cb.Goto(l)
}

func compileReturnStmt(ctx *blockCtx, stmt *ast.Node) {
	n := len(stmt.Inner)
	if n > 0 {
		n = 1
		compileExpr(ctx, stmt.Inner[0])
		cb := ctx.cb
		typeCast(ctx, getRetType(cb), cb.Get(-1))
	}
	ctx.cb.Return(n, ctx.goNode(stmt))
}

func getRetType(cb *gox.CodeBuilder) types.Type {
	return cb.Func().Type().(*types.Signature).Results().At(0).Type()
}

func getRetTypeEx(cb *gox.CodeBuilder) (ret types.Type, ok bool) {
	results := cb.Func().Type().(*types.Signature).Results()
	if results.Len() == 1 {
		return results.At(0).Type(), true
	}
	return
}

func hasMultiInitDeclStmt(ctx *blockCtx, initStmt *ast.Node) bool {
	if initStmt.Kind == ast.DeclStmt && len(initStmt.Inner) > 1 {
		return true
	}
	return false
}

// -----------------------------------------------------------------------------
