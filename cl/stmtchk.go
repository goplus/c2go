package cl

import "github.com/goplus/c2go/clang/ast"

// -----------------------------------------------------------------------------

type labelStat struct {
	firstCase int
	multi     bool
	define    bool
}

type nonSimpleSwChecker struct {
	labels  map[string]*labelStat
	idxCase int
}

func (p *nonSimpleSwChecker) checkStmt(ctx *blockCtx, stmt *ast.Node) {
	var name string
	var define bool
	switch stmt.Kind {
	case ast.LabelStmt:
		p.checkStmt(ctx, stmt.Inner[0])
		name, define = stmt.Name, true
	case ast.GotoStmt:
		name = ctx.labelOfGoto(stmt)
	case ast.CompoundStmt:
		for _, item := range stmt.Inner {
			p.checkStmt(ctx, item)
		}
		return
	case ast.IfStmt:
		p.checkStmt(ctx, stmt.Inner[1])
		if stmt.HasElse {
			p.checkStmt(ctx, stmt.Inner[2])
		}
		return
	case ast.CaseStmt, ast.DefaultStmt:
		panic(true)
	default:
		return
	}
	l, ok := p.labels[name]
	if !ok {
		l = &labelStat{firstCase: p.idxCase, define: define}
		p.labels[name] = l
		return
	}
	if l.firstCase != p.idxCase {
		l.multi = true
	}
	if define {
		l.define = true
	}
	if l.multi && l.define {
		panic(true)
	}
}

func isSimpleSwitch(ctx *blockCtx, switchStmt *ast.Node) (simple bool) {
	body := switchStmt.Inner[1]
	if firstStmtNotCase(body) {
		return false
	}
	bodyStmts := body.Inner
	checker := &nonSimpleSwChecker{
		labels: make(map[string]*labelStat), // labelName => idxCase
	}
	defer func() {
		simple = recover() == nil
	}()
	for i, n := 0, len(bodyStmts); i < n; i++ {
		stmt := bodyStmts[i]
	retry:
		var idx int
		switch stmt.Kind {
		case ast.CaseStmt:
			idx = 1
		case ast.DefaultStmt:
		default:
			checker.checkStmt(ctx, stmt)
			continue
		}
		checker.idxCase++
		switch caseBody := stmt.Inner[idx]; caseBody.Kind {
		case ast.CaseStmt, ast.DefaultStmt:
			stmt = caseBody
			checker.idxCase++
			goto retry
		default:
			checker.checkStmt(ctx, caseBody)
		}
	}
	return
}

// -----------------------------------------------------------------------------
