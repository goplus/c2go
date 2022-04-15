package cl

import (
	"bytes"
	"go/token"
	"go/types"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/gox"
	"github.com/qiniu/x/ctype"

	ctypes "github.com/goplus/c2go/clang/types"
)

// -----------------------------------------------------------------------------

type funcCtx struct {
	labels map[string]*gox.Label
	base   int
}

func newFuncCtx() *funcCtx {
	return &funcCtx{
		labels: make(map[string]*gox.Label),
	}
}

func (p *funcCtx) newLabel(cb *gox.CodeBuilder) *gox.Label {
	p.base++
	name := "_cgol_" + strconv.Itoa(p.base)
	return cb.NewLabel(token.NoPos, name)
}

func (p *funcCtx) label(cb *gox.CodeBuilder) *gox.Label {
	l := p.newLabel(cb)
	cb.Label(l)
	return l
}

// -----------------------------------------------------------------------------

type flowCtx interface { // switch, for
	Parent() flowCtx
	EndLabel(ctx *blockCtx) *gox.Label
	ContinueLabel(ctx *blockCtx) *gox.Label
}

type baseFlowCtx struct {
	parent flowCtx
}

func (p *baseFlowCtx) Parent() flowCtx {
	return p.parent
}

func (p *baseFlowCtx) EndLabel(ctx *blockCtx) *gox.Label {
	return nil
}

func (p *baseFlowCtx) ContinueLabel(ctx *blockCtx) *gox.Label {
	return nil
}

// -----------------------------------------------------------------------------

type switchCtx struct {
	parent flowCtx
	done   *gox.Label
	next   *gox.Label
	defau  *gox.Label
	vdefs  *gox.VarDefs
	scope  *types.Scope
	tag    types.Object
	notmat types.Object // notMatched
}

func (p *switchCtx) Parent() flowCtx {
	return p.parent
}

func (p *switchCtx) EndLabel(ctx *blockCtx) *gox.Label {
	done := p.done
	if done == nil {
		done = ctx.curfn.newLabel(ctx.cb)
		p.done = done
	}
	return done
}

func (p *switchCtx) ContinueLabel(ctx *blockCtx) *gox.Label {
	return p.parent.ContinueLabel(ctx)
}

func (p *switchCtx) nextCaseLabel(ctx *blockCtx) *gox.Label {
	l := ctx.curfn.newLabel(ctx.cb)
	p.next = l
	return l
}

func (p *switchCtx) labelDefault(ctx *blockCtx) {
	p.defau = ctx.curfn.label(ctx.cb)
}

// -----------------------------------------------------------------------------

type loopCtx struct {
	parent flowCtx
	done   *gox.Label
	start  *gox.Label
}

func (p *loopCtx) Parent() flowCtx {
	return p.parent
}

func (p *loopCtx) EndLabel(ctx *blockCtx) *gox.Label {
	done := p.done
	if done == nil {
		done = ctx.curfn.newLabel(ctx.cb)
		p.done = done
	}
	return done
}

func (p *loopCtx) ContinueLabel(ctx *blockCtx) *gox.Label {
	return p.start
}

func (p *loopCtx) labelStart(ctx *blockCtx) {
	p.start = ctx.curfn.label(ctx.cb)
}

// -----------------------------------------------------------------------------

type blockCtx struct {
	pkg      *gox.Package
	cb       *gox.CodeBuilder
	fset     *token.FileSet
	tyValist types.Type
	unnameds map[ast.ID]*types.Named
	typdecls map[string]*gox.TypeDecl
	srcfile  string
	cursrc   []byte
	curfn    *funcCtx
	curflow  flowCtx
	asuBase  int // anonymous struct/union
}

func (p *blockCtx) getSwitchCtx() *switchCtx {
	for f := p.curflow; f != nil; f = f.Parent() {
		if sw, ok := f.(*switchCtx); ok {
			return sw
		}
	}
	return nil
}

func (p *blockCtx) getVarDefs(scope *types.Scope) (vdefs *gox.VarDefs, inSwitch bool) {
	if sw := p.getSwitchCtx(); sw != nil && sw.scope == scope {
		return sw.vdefs, true
	}
	return p.pkg.NewVarDefs(scope), false
}

func (p *blockCtx) enterSwitch() *switchCtx {
	scope := p.cb.Scope()
	vdefs := p.pkg.NewVarDefs(scope)
	f := &switchCtx{parent: p.curflow, vdefs: vdefs, scope: scope}
	p.curflow = f
	return f
}

func (p *blockCtx) enterLoop() *loopCtx {
	f := &loopCtx{parent: p.curflow}
	p.curflow = f
	return f
}

func (p *blockCtx) enterFlow() *baseFlowCtx {
	f := &baseFlowCtx{parent: p.curflow}
	p.curflow = f
	return f
}

func (p *blockCtx) leave(cur flowCtx) {
	p.curflow = cur.Parent()
}

func (p *blockCtx) getSource() []byte {
	if v := p.cursrc; v != nil {
		return v
	}
	b, err := os.ReadFile(p.srcfile)
	if err != nil {
		log.Panicln("getSource:", err)
	}
	p.cursrc = b
	return b
}

func (p *blockCtx) getLabel(pos token.Pos, name string) *gox.Label {
	if fn := p.curfn; fn != nil {
		l, ok := fn.labels[name]
		if !ok {
			l = p.cb.NewLabel(pos, name)
			fn.labels[name] = l
		}
		return l
	}
	log.Panicln("can't use label out of func")
	return nil
}

func (p *blockCtx) labelOfGoto(v *ast.Node) string {
	src := p.getSource()
	off := v.Range.Begin.Offset
	n := int64(v.Range.Begin.TokLen)
	op := string(src[off : off+n])
	if op != "goto" {
		log.Panicln("gotoOp:", op)
	}
	label := ident(src[off+n:], "label not found")
	return label
}

func (p *blockCtx) typeOfSizeof(v *ast.Node) string {
	src := p.getSource()
	off := v.Range.Begin.Offset
	n := int64(v.Range.Begin.TokLen)
	op := string(src[off : off+n])
	if op != "sizeof" {
		log.Panicln("sizeofOp:", op)
	}
	typ := string(src[off+n : v.Range.End.Offset])
	return strings.TrimPrefix(strings.TrimLeft(typ, " \t\r\n"), "(")
}

func (p *blockCtx) getInstr(v *ast.Node) string {
	src := p.getSource()
	off := v.Range.Begin.Offset
	n := int64(v.Range.Begin.TokLen)
	return string(src[off : off+n])
}

func ident(b []byte, msg string) string {
	b = bytes.TrimLeft(b, " \t\r\n")
	idx := bytes.IndexFunc(b, func(r rune) bool {
		return !ctype.Is(ctype.CSYMBOL_NEXT_CHAR, r)
	})
	if idx <= 0 {
		log.Panicln(msg)
	}
	return string(b[:idx])
}

func (p *blockCtx) sizeof(typ types.Type) int {
	return int(p.pkg.Sizeof(typ))
}

const (
	suNormal = iota
	suAnonymous
)

func (p *blockCtx) getSuName(v *ast.Node, tag string) (string, int) {
	if name := v.Name; name != "" {
		return ctypes.MangledName(tag, name), suNormal
	}
	p.asuBase++
	return "_cgoa_" + strconv.Itoa(p.asuBase), suAnonymous
}

func (p *blockCtx) initCTypes() {
	pkg := p.pkg.Types
	scope := pkg.Scope()
	p.tyValist = initValist(scope, pkg)
	aliasType(scope, pkg, "char", types.Typ[types.Int8])
	aliasType(scope, pkg, "void", ctypes.Void)
	aliasType(scope, pkg, "float", types.Typ[types.Float32])
	aliasType(scope, pkg, "double", types.Typ[types.Float64])
	aliasType(scope, pkg, "__int128", ctypes.Int128)
	aliasType(scope, pkg, "_Bool", types.Typ[types.Bool])
	decl_builtin(p)
}

func initValist(scope *types.Scope, pkg *types.Package) types.Type {
	valist := types.NewTypeName(token.NoPos, pkg, ctypes.MangledName("struct", "__va_list_tag"), nil)
	t := types.NewNamed(valist, types.Typ[types.Int8], nil)
	scope.Insert(valist)
	tyValist := types.NewPointer(t)
	aliasType(scope, pkg, "__builtin_va_list", tyValist)
	return tyValist
}

func aliasType(scope *types.Scope, pkg *types.Package, name string, typ types.Type) {
	o := types.NewTypeName(token.NoPos, pkg, name, typ)
	scope.Insert(o)
}

// -----------------------------------------------------------------------------
