package cl

import (
	"log"
	"time"

	"github.com/goplus/c2go/clang/ast"
)

// -----------------------------------------------------------------------------

type none = struct{}

type blockMarkCtx struct {
	parent    *blockMarkCtx
	owner     *blockMarkCtx
	ownerStmt *ast.Node
}

func markComplicatedByDepth(ctx *markCtx, a, b *blockMarkCtx) {
	aDepth, bDepth := a.depth(), b.depth()
retry:
	if aDepth < bDepth {
		ctx.markComplicated(b.ownerStmt)
		b = b.owner
		bDepth = b.depth()
		goto retry
	} else if bDepth < aDepth {
		ctx.markComplicated(a.ownerStmt)
		a = a.owner
		aDepth = a.depth()
		goto retry
	} else if a != b {
		ctx.markComplicated(a.ownerStmt)
		ctx.markComplicated(b.ownerStmt)
		a, b = a.owner, b.owner
		aDepth, bDepth = a.depth(), b.depth()
		goto retry
	}
}

func (at *blockMarkCtx) markComplicated(ctx *markCtx, ref *blockMarkCtx) {
	if at.isComplicated(ref) {
		ctx.markComplicated(at.ownerStmt)
		markComplicatedByDepth(ctx, at, ref)
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
	complicat bool
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
			p.markComplicated(stmt)
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
		p.markComplicated(switchStmt)
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
			l.at.markComplicated(p, ref)
		}
	}
}

func (p *markCtx) markComplicated(stmt *ast.Node) {
	stmt.Complicated = true
	p.complicat = true
}

func (p *blockCtx) markComplicated(name string, body *ast.Node) bool {
	if debugMarkComplicated {
		start := time.Now()
		defer func() {
			log.Printf("==> Marked %s: %v\n", name, time.Since(start))
		}()
	}
	labels := make(map[string]*labelCtx)
	marker := &markCtx{labels: labels}
	marker.mark(p, body)
	marker.markEnd()
	return marker.complicat
}

// -----------------------------------------------------------------------------
