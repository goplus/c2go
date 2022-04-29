package cl

import (
	"log"
	"time"

	"github.com/goplus/c2go/clang/ast"
)

// -----------------------------------------------------------------------------

type none = struct{}

type blockMarkCtx struct {
	parent *blockMarkCtx
	owner  *ownerStmtCtx
	name   string
}

func markComplicatedByDepth(ctx *markCtx, a, b *blockMarkCtx) {
	aDepth, bDepth := a.depth(), b.depth()
retry:
	if aDepth < bDepth {
		b = b.owner.markComplicated(ctx)
		bDepth = b.depth()
		goto retry
	} else if bDepth < aDepth {
		a = a.owner.markComplicated(ctx)
		aDepth = a.depth()
		goto retry
	} else if a != b {
		a, b = a.owner.markComplicated(ctx), b.owner.markComplicated(ctx)
		aDepth, bDepth = a.depth(), b.depth()
		goto retry
	}
}

func (at *blockMarkCtx) markComplicated(ctx *markCtx, ref *blockMarkCtx) {
	if at.isComplicated(ref) {
		at.owner.markComplicated(ctx)
		markComplicatedByDepth(ctx, at, ref)
	}
}

func (at *blockMarkCtx) getName() string {
	if at != nil {
		return at.name
	}
	return "funcBody"
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
	if at.owner.isComplicated() {
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
	at      *blockMarkCtx          // block that defines this label
	refs    map[*blockMarkCtx]none // blocks that refer this label
	defined bool
}

func (p *labelCtx) defineLabel(name string, at *blockMarkCtx) {
	if p.defined {
		log.Panicln("defineLabel: label exists -", name)
	}
	p.at, p.defined = at, true
	if debugMarkComplicated {
		log.Println("--> label", name, "depth:", at.depth())
	}
}

func (p *labelCtx) useLabel(name string, at *blockMarkCtx) {
	if p.refs == nil {
		p.refs = make(map[*blockMarkCtx]none)
	}
	p.refs[at] = none{}
	if debugMarkComplicated {
		log.Println("--> goto", name, "from depth:", at.depth())
	}
}

type ownerStmtCtx struct {
	parent *blockMarkCtx
	stmt   *ast.Node
}

func (p *ownerStmtCtx) isComplicated() bool {
	if p != nil {
		return p.stmt.Complicated
	}
	return false
}

func (p *ownerStmtCtx) markComplicated(ctx *markCtx) *blockMarkCtx {
	if p != nil {
		ctx.markComplicated(p.stmt)
		return p.parent
	}
	return nil
}

type markCtx struct {
	current   *blockMarkCtx
	owner     *ownerStmtCtx
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

func (p *markCtx) enter(name string) *blockMarkCtx {
	self := &blockMarkCtx{parent: p.current, owner: p.owner, name: name}
	p.current = self
	if debugMarkComplicated {
		log.Println("--> enter", name, "depth:", self.depth())
	}
	return self
}

func (p *markCtx) leave(self *blockMarkCtx) {
	p.current = self.parent
}

func (p *markCtx) enterOwner(stmt *ast.Node) (old *ownerStmtCtx) {
	if debugMarkComplicated {
		log.Println("--> stmt", stmt.Kind, "depth:", p.current.depth())
	}
	p.owner, old = &ownerStmtCtx{parent: p.current, stmt: stmt}, p.owner
	return
}

func (p *markCtx) leaveOwner(old *ownerStmtCtx) {
	p.owner = old
}

func (p *markCtx) markSub(ctx *blockCtx, name string, stmt *ast.Node) {
	self := p.enter(name)
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
		p.markSub(ctx, "ifBody", stmt.Inner[1])
		if stmt.HasElse {
			p.markSub(ctx, "elseBody", stmt.Inner[2])
		}
	case ast.SwitchStmt:
		p.markSwitch(ctx, stmt)
	case ast.ForStmt:
		ret := p.enterOwner(stmt)
		defer p.leaveOwner(ret)
		p.markSub(ctx, "forBody", stmt.Inner[4])
	case ast.WhileStmt:
		ret := p.enterOwner(stmt)
		defer p.leaveOwner(ret)
		p.markSub(ctx, "whileBody", stmt.Inner[1])
	case ast.DoStmt:
		ret := p.enterOwner(stmt)
		defer p.leaveOwner(ret)
		p.markSub(ctx, "doBody", stmt.Inner[0])
	case ast.LabelStmt:
		name := stmt.Name
		p.reqLabel(name).defineLabel(name, p.current)
	case ast.GotoStmt:
		name := ctx.labelOfGoto(stmt)
		p.reqLabel(name).useLabel(name, p.current)
	case ast.CompoundStmt:
		ret := p.enterOwner(stmt)
		defer p.leaveOwner(ret)
		p.markSub(ctx, "blockBody", stmt)
	case ast.CaseStmt, ast.DefaultStmt:
		p.markSwitchComplicated()
	}
}

func (p *markCtx) markSwitchComplicated() {
	owner := p.owner
	for owner != nil {
		p.markComplicated(owner.stmt)
		if owner.stmt.Kind == ast.SwitchStmt {
			return
		}
		owner = owner.parent.owner
	}
}

func (p *markCtx) markSwitch(ctx *blockCtx, switchStmt *ast.Node) {
	ret := p.enterOwner(switchStmt)
	defer p.leaveOwner(ret)

	self := p.enter("switchBody")
	defer p.leave(self)

	body := switchStmt.Inner[1]
	if firstStmtNotCase(body) {
		p.markComplicated(switchStmt)
		p.markComplicated(body)
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
		caseCtx = p.enter("caseBody")
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
	if stmt == nil {
		return
	}
	if debugMarkComplicated && !stmt.Complicated {
		log.Println("==> markComplicated", stmt.Kind, *stmt.Range)
	}
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
	marker.markBody(p, body)
	marker.markEnd()
	return marker.complicat
}

// -----------------------------------------------------------------------------
