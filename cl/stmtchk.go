package cl

import (
	"log"

	"github.com/goplus/c2go/clang/ast"
)

// -----------------------------------------------------------------------------

type none = struct{}

type blockMarkCtx struct {
	parent    *blockMarkCtx
	owner     *blockMarkCtx
	ownerStmt *ast.Node
}

func markComplicatedByDepth(a, b *blockMarkCtx) {
	aDepth, bDepth := a.depth(), b.depth()
retry:
	if aDepth < bDepth {
		markComplicated(b.ownerStmt)
		b = b.owner
		bDepth = b.depth()
		goto retry
	} else if bDepth < aDepth {
		markComplicated(a.ownerStmt)
		a = a.owner
		aDepth = a.depth()
		goto retry
	} else if a != b {
		markComplicated(a.ownerStmt)
		markComplicated(b.ownerStmt)
		a, b = a.owner, b.owner
		aDepth, bDepth = a.depth(), b.depth()
		goto retry
	}
}

func (at *blockMarkCtx) markComplicated(ref *blockMarkCtx) {
	if at.isComplicated(ref) {
		markComplicated(at.ownerStmt)
		markComplicatedByDepth(at, ref)
	}
}

func (at *blockMarkCtx) depth() (n int) {
	for at != nil {
		n++
		at = at.parent
	}
	return
}

func (at *blockMarkCtx) isComplicated(ref *blockMarkCtx) bool {
	if at == nil {
		return false
	}
	if at.ownerStmt.Complicated {
		return true
	}
	for ref != nil {
		if at == ref {
			return false
		}
		ref = ref.parent
	}
	return true
}

type labelCtx struct {
	at   *blockMarkCtx          // block that defines this label
	refs map[*blockMarkCtx]none // blocks that refer this label
}

func (p *labelCtx) defineLabel(name string, at *blockMarkCtx) {
	if p.at != nil {
		log.Panicln("defineLabel: label exists -", name)
	}
	p.at = at
}

func (p *labelCtx) useLabel(at *blockMarkCtx) {
	if p.at == at {
		return
	}
	if p.refs == nil {
		p.refs = make(map[*blockMarkCtx]none)
	}
	p.refs[at] = none{}
}

type ownerBlockMarkCtx struct {
	self      *blockMarkCtx
	owner     *blockMarkCtx
	ownerStmt *ast.Node
}

type markCtx struct {
	current   *blockMarkCtx
	owner     *blockMarkCtx
	ownerStmt *ast.Node
	labels    map[string]*labelCtx
}

func (p *markCtx) reqLabel(name string) *labelCtx {
	l, ok := p.labels[name]
	if !ok {
		l = &labelCtx{}
		p.labels[name] = l
	}
	return l
}

func (p *markCtx) enter() *blockMarkCtx {
	self := &blockMarkCtx{parent: p.current, owner: p.owner, ownerStmt: p.ownerStmt}
	p.current = self
	return self
}

func (p *markCtx) leave(self *blockMarkCtx) {
	p.current = self.parent
}

func (p *markCtx) enterOwner(stmt *ast.Node) (ret ownerBlockMarkCtx) {
	ret.self = p.enter()
	ret.owner, ret.ownerStmt = p.owner, p.ownerStmt
	p.owner, p.ownerStmt = ret.self, stmt
	return
}

func (p *markCtx) leaveOwner(ret ownerBlockMarkCtx) {
	p.leave(ret.self)
	p.owner, p.ownerStmt = ret.owner, ret.ownerStmt
}

func (p *markCtx) markSub(ctx *blockCtx, stmt *ast.Node) {
	self := p.enter()
	defer p.leave(self)
	p.markBody(ctx, stmt)
}

func (p *markCtx) markBody(ctx *blockCtx, stmt *ast.Node) {
	switch stmt.Kind {
	case ast.CompoundStmt:
		for _, item := range stmt.Inner {
			p.mark(ctx, item)
		}
		return
	}
	p.mark(ctx, stmt)
}

func (p *markCtx) mark(ctx *blockCtx, stmt *ast.Node) {
	switch stmt.Kind {
	case ast.IfStmt:
		ret := p.enterOwner(stmt)
		defer p.leaveOwner(ret)
		p.markSub(ctx, stmt.Inner[1])
		if stmt.HasElse {
			p.markSub(ctx, stmt.Inner[2])
		}
	case ast.SwitchStmt:
		p.markSwitch(ctx, stmt)
	case ast.LabelStmt:
		name := stmt.Name
		p.reqLabel(name).defineLabel(name, p.current)
	case ast.GotoStmt:
		name := ctx.labelOfGoto(stmt)
		p.reqLabel(name).useLabel(p.current)
	case ast.CompoundStmt:
		for _, item := range stmt.Inner {
			p.mark(ctx, item)
		}
	case ast.CaseStmt, ast.DefaultStmt:
		p.markSwitchComplicated()
	}
}

func (p *markCtx) markSwitchComplicated() {
	owner, stmt := p.owner, p.ownerStmt
	for owner != nil {
		if stmt.Kind == ast.SwitchStmt {
			markComplicated(stmt)
			return
		}
		owner, stmt = owner.owner, owner.ownerStmt
	}
}

func (p *markCtx) markSwitch(ctx *blockCtx, switchStmt *ast.Node) {
	ret := p.enterOwner(switchStmt)
	defer p.leaveOwner(ret)

	body := switchStmt.Inner[1]
	if firstStmtNotCase(body) {
		markComplicated(switchStmt)
	}
	var bodyStmts = body.Inner
	var caseCtx *blockMarkCtx
	for i, n := 0, len(bodyStmts); i < n; i++ {
		stmt := bodyStmts[i]
	retry:
		var idx int
		switch stmt.Kind {
		case ast.CaseStmt:
			idx = 1
		case ast.DefaultStmt:
		default:
			p.mark(ctx, stmt)
			continue
		}
		if caseCtx != nil {
			p.leave(caseCtx)
			caseCtx = nil
		}
		caseCtx = p.enter()
		switch caseBody := stmt.Inner[idx]; caseBody.Kind {
		case ast.CaseStmt, ast.DefaultStmt:
			stmt = caseBody
			goto retry
		default:
			p.mark(ctx, caseBody)
		}
	}
	return
}

func (p *markCtx) markEnd() {
	for _, l := range p.labels {
		for ref := range l.refs {
			l.at.markComplicated(ref)
		}
	}
}

func markComplicated(stmt *ast.Node) {
	stmt.Complicated = true
}

func (p *blockCtx) markComplicated(body *ast.Node) {
	labels := make(map[string]*labelCtx)
	marker := &markCtx{labels: labels}
	marker.mark(p, body)
	marker.markEnd()
}

func isSimpleSwitch(ctx *blockCtx, switchStmt *ast.Node) bool {
	ctx.markComplicated(switchStmt)
	return !switchStmt.Complicated
}

// -----------------------------------------------------------------------------
/*
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
*/
// -----------------------------------------------------------------------------
