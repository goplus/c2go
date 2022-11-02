package cl

import (
	"bytes"
	"fmt"
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

const (
	space = " \t\r\n"
)

// -----------------------------------------------------------------------------

type funcCtx struct {
	labels  map[string]*gox.Label
	vdefs   *gox.VarDefs
	basel   int
	basev   int
	orgName string
}

func newFuncCtx(pkg *gox.Package, complicated bool, orgName string) *funcCtx {
	ctx := &funcCtx{
		labels:  make(map[string]*gox.Label),
		orgName: orgName,
	}
	if complicated {
		ctx.vdefs = pkg.NewVarDefs(pkg.CB().Scope())
	}
	return ctx
}

func (p *funcCtx) newLabel(cb *gox.CodeBuilder) *gox.Label {
	p.basel++
	name := "_cgol_" + strconv.Itoa(p.basel)
	return cb.NewLabel(token.NoPos, name)
}

func (p *funcCtx) label(cb *gox.CodeBuilder) *gox.Label {
	l := p.newLabel(cb)
	cb.Label(l)
	return l
}

func (p *funcCtx) newAutoVar(pos token.Pos, typ types.Type, name string) (*gox.VarDecl, types.Object) {
	p.basev++
	realName := name + "_cgo" + strconv.Itoa(p.basev)
	ret := p.vdefs.New(pos, typ, realName)
	return ret, ret.Ref(realName)
}

// -----------------------------------------------------------------------------

type flowCtx interface { // switch, for
	Parent() flowCtx
	EndLabel(ctx *blockCtx) *gox.Label
	ContinueLabel(ctx *blockCtx) *gox.Label
}

// -----------------------------------------------------------------------------

const (
	flowKindIf = 1 << iota
	flowKindSwitch
	flowKindLoop
)

type baseFlowCtx struct {
	parent flowCtx
	kind   int // flowKindIf|Switch|Loop
}

func (p *baseFlowCtx) Parent() flowCtx {
	return p.parent
}

func (p *baseFlowCtx) EndLabel(ctx *blockCtx) *gox.Label {
	if (p.kind & (flowKindLoop | flowKindSwitch)) != 0 {
		return nil
	}
	return p.parent.EndLabel(ctx)
}

func (p *baseFlowCtx) ContinueLabel(ctx *blockCtx) *gox.Label {
	if (p.kind & flowKindLoop) != 0 {
		return nil
	}
	return p.parent.ContinueLabel(ctx)
}

// -----------------------------------------------------------------------------

type endLabelCtx struct {
	done *gox.Label
}

func (p *endLabelCtx) EndLabel(ctx *blockCtx) *gox.Label {
	done := p.done
	if done == nil {
		done = ctx.curfn.newLabel(ctx.cb)
		p.done = done
	}
	return done
}

// -----------------------------------------------------------------------------

type switchCtx struct {
	endLabelCtx
	parent flowCtx
	next   *gox.Label
	defau  *gox.Label
	tag    types.Object
	notmat types.Object // notMatched
}

func (p *switchCtx) Parent() flowCtx {
	return p.parent
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

type ifCtx struct {
	endLabelCtx
	parent flowCtx
}

func (p *ifCtx) Parent() flowCtx {
	return p.parent
}

func (p *ifCtx) ContinueLabel(ctx *blockCtx) *gox.Label {
	return p.parent.ContinueLabel(ctx)
}

func (p *ifCtx) elseLabel(ctx *blockCtx) *gox.Label {
	return ctx.curfn.newLabel(ctx.cb)
}

// -----------------------------------------------------------------------------

type loopCtx struct {
	endLabelCtx
	parent flowCtx
	start  *gox.Label
}

func (p *loopCtx) Parent() flowCtx {
	return p.parent
}

func (p *loopCtx) ContinueLabel(ctx *blockCtx) *gox.Label {
	return p.start
}

func (p *loopCtx) labelStart(ctx *blockCtx) {
	p.start = ctx.curfn.label(ctx.cb)
}

// -----------------------------------------------------------------------------

type delfunc = []string

type unnamedType struct {
	typ *types.Named
	del delfunc
}

type blockCtx struct {
	pkg      *gox.Package
	cb       *gox.CodeBuilder
	fset     *token.FileSet
	tyI128   types.Type
	tyU128   types.Type
	unnameds map[ast.ID]unnamedType
	srcenums map[ast.ID]string
	gblvars  map[string]*gox.VarDefs
	public   map[string]string
	ignored  []string
	srcdir   string
	srcfile  string
	src      []byte
	file     *token.File
	curfn    *funcCtx
	curflow  flowCtx
	bfm      BFMode
	multiFileCtl
	testMain bool
}

func (p *blockCtx) Pkg() *types.Package {
	return p.pkg.Types
}

func (p *blockCtx) Int128() types.Type {
	return p.tyI128
}

func (p *blockCtx) Uint128() types.Type {
	return p.tyU128
}

func (p *blockCtx) deleteUnnamed(id ast.ID) {
	if u, ok := p.unnameds[id]; ok {
		for _, delName := range u.del {
			p.deleteByName(delName)
		}
		p.deleteByName(u.typ.Obj().Name())
	}
}

func (p *blockCtx) deleteByName(name string) {
	if t, decled := p.typdecls[name]; decled {
		t.Delete()
	}
}

func (p *blockCtx) addExternFunc(name string) {
	for _, ign := range p.ignored {
		if ign == name {
			return
		}
	}
	p.extfns[name] = none{}
}

func (p *blockCtx) lookupParent(name string) types.Object {
	_, o := gox.LookupParent(p.cb.Scope(), name, token.NoPos)
	return o
}

func (p *blockCtx) newVar(scope *types.Scope, pos token.Pos, typ types.Type, name string) (ret *gox.VarDecl, inVBlock bool) {
	cb, pkg := p.cb, p.pkg
	if inVBlock = cb.InVBlock(); inVBlock {
		var obj types.Object
		ret, obj = p.curfn.newAutoVar(pos, typ, name)
		if scope.Insert(gox.NewSubst(pos, pkg.Types, name, obj)) != nil {
			log.Panicf("newVar: variable %v exists already\n", name)
		}
	} else {
		inGlobal := scope == pkg.Types.Scope()
		if inGlobal {
			if defs, ok := p.gblvars[name]; ok {
				defs.Delete(name)
				delete(p.gblvars, name)
			}
		}
		defs := pkg.NewVarDefs(scope)
		ret = defs.New(pos, typ, name)
		if inGlobal {
			p.gblvars[name] = defs
		}
	}
	return
}

func (p *blockCtx) getSwitchCtx() *switchCtx {
	for f := p.curflow; f != nil; f = f.Parent() {
		if sw, ok := f.(*switchCtx); ok {
			return sw
		}
	}
	return nil
}

func (p *blockCtx) enterIf() *ifCtx {
	f := &ifCtx{parent: p.curflow}
	p.cb.VBlock()
	p.curflow = f
	return f
}

func (p *blockCtx) enterSwitch() *switchCtx {
	f := &switchCtx{parent: p.curflow}
	p.cb.VBlock()
	p.curflow = f
	return f
}

func (p *blockCtx) enterLoop() *loopCtx {
	f := &loopCtx{parent: p.curflow}
	p.cb.VBlock()
	p.curflow = f
	return f
}

func (p *blockCtx) enterFlow(kind int) *baseFlowCtx {
	f := &baseFlowCtx{parent: p.curflow, kind: kind}
	p.curflow = f
	return f
}

func (p *blockCtx) leave(cur flowCtx) {
	if _, simple := cur.(*baseFlowCtx); !simple {
		p.cb.End()
	}
	p.curflow = cur.Parent()
}

func (p *blockCtx) initFile() {
	src := p.initSource()
	p.file = p.fset.AddFile(p.srcfile, -1, len(src))
	p.file.SetLinesForContent(src)
}

func (p *blockCtx) initSource() []byte {
	if p.src != nil {
		return p.src
	}
	b, err := os.ReadFile(p.srcfile)
	if err != nil {
		log.Panicln("initSource:", err)
	}
	p.src = b
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
	src := p.src
	off := v.Range.Begin.Offset
	n := int64(v.Range.Begin.TokLen)
	op := string(src[off : off+n])
	if op != "goto" {
		log.Panicln("gotoOp:", op)
	}
	label := ident(src[off+n:], "label not found")
	return label
}

func (p *blockCtx) paramsOfOfsetof(v *ast.Node) (string, string) {
	src := p.src
	off := v.Range.Begin.Offset
	n := int64(v.Range.Begin.TokLen)
	op := string(src[off : off+n])
	if op != "__builtin_offsetof" {
		log.Panicln("unknown offsetofOp:", op)
	}
	params := strings.SplitN(paramsOf(src[off+n:v.Range.End.Offset]), ",", 2)
	return params[0], strings.Trim(params[1], space)
}

func paramsOf(v []byte) string {
	return strings.TrimPrefix(strings.TrimLeft(string(v), space), "(")
}

func (p *blockCtx) paramOfSizeof(v *ast.Node) string {
	src := p.src
	off := v.Range.Begin.Offset
	n := int64(v.Range.Begin.TokLen)
	op := string(src[off : off+n])
	if op != "sizeof" {
		log.Panicln("unknown sizeofOp:", op)
	}
	return paramsOf(src[off+n : v.Range.End.Offset])
}

func (p *blockCtx) getInstr(v *ast.Node) string {
	src := p.src
	off := v.Range.Begin.Offset
	n := int64(v.Range.Begin.TokLen)
	return string(src[off : off+n])
}

func ident(b []byte, msg string) string {
	b = bytes.TrimLeft(b, space)
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

func (p *blockCtx) offsetof(typ types.Type, name string) int {
retry:
	switch t := typ.(type) {
	case *types.Struct:
		if flds, idx := getFld(t, name, 0); idx >= 0 {
			return int(p.pkg.Offsetsof(flds)[idx])
		}
	case *types.Named:
		typ = t.Underlying()
		goto retry
	}
	log.Panicf("offsetof(%v, %v): field not found", typ, name)
	return -1
}

func getFld(t *types.Struct, name string, from int) (flds []*types.Var, i int) {
	var n int
	for i, n = from, t.NumFields(); i < n; i++ {
		f := t.Field(i)
		flds = append(flds, f)
		if f.Name() == name {
			return
		}
	}
	return nil, -1
}

func (p *blockCtx) buildVStruct(struc *types.Struct, vfs gox.VFields) *types.Struct {
	var pkg = p.pkg.Types
	var vFlds []*types.Var
	switch v := vfs.(type) {
	case *gox.BitFields:
		from, n := 0, v.Len()
		for i := 0; i < n; i++ {
			f := v.At(i)
			name := f.FldName
			flds, idx := getFld(struc, name, from)
			if idx < 0 {
				log.Panicln("buildVStruct: field not found -", name)
			}
			realf := flds[idx-from]
			vft := &bfType{Type: realf.Type(), BitField: f, first: true}
			vFlds = append(vFlds, flds[:idx-from]...)
			vFlds = append(vFlds, types.NewField(token.NoPos, pkg, f.Name, vft, false))
			for i+1 < n {
				nextf := v.At(i + 1)
				if nextf.FldName != name {
					break
				}
				vft = &bfType{Type: realf.Type(), BitField: nextf}
				vFlds = append(vFlds, types.NewField(token.NoPos, pkg, nextf.Name, vft, false))
				i++
			}
			from = idx + 1
		}
		for n = struc.NumFields(); from < n; from++ {
			vFlds = append(vFlds, struc.Field(from))
		}
		return types.NewStruct(vFlds, nil)
	}
	return struc
}

func (p *blockCtx) getVStruct(typ *types.Named) *types.Struct {
	t, ok := typ.Underlying().(*types.Struct)
	if !ok {
		log.Panicln(typ.Obj().Name(), "not found")
	}
	if vfs, ok := p.pkg.VFields(typ); ok {
		t = p.buildVStruct(t, vfs)
	}
	return t
}

type bfType struct {
	types.Type
	*gox.BitField
	first bool
}

func (p *bfType) String() string {
	return fmt.Sprintf("bfType{t: %v, bf: %v, first: %v}", p.Type, p.BitField, p.first)
}

/*
func (p *blockCtx) initCTypes() {
	pkg := p.pkg.Types
	scope := pkg.Scope()
	p.tyI128 = ctypes.NotImpl
	p.tyU128 = ctypes.NotImpl
	c := p.pkg.Import("github.com/goplus/c2go/clang")

	aliasType(scope, pkg, "__int128", p.tyI128)
	aliasType(scope, pkg, "void", ctypes.Void)

	aliasCType(scope, pkg, "char", c, "Char")
	aliasCType(scope, pkg, "float", c, "Float")
	aliasCType(scope, pkg, "double", c, "Double")
	aliasCType(scope, pkg, "_Bool", c, "Bool")

	decl_builtin(p)
}

func aliasCType(scope *types.Scope, pkg *types.Package, name string, c *gox.PkgRef, cname string) {
	aliasType(scope, pkg, name, c.Ref(cname).Type())
}
*/
func (p *blockCtx) initCTypes() {
	pkg := p.pkg.Types
	scope := pkg.Scope()
	p.tyI128 = ctypes.NotImpl
	p.tyU128 = ctypes.NotImpl

	aliasType(scope, pkg, "__int128", p.tyI128)
	aliasType(scope, pkg, "void", ctypes.Void)

	aliasType(scope, pkg, "char", types.Typ[types.Int8])
	aliasType(scope, pkg, "float", types.Typ[types.Float32])
	aliasType(scope, pkg, "double", types.Typ[types.Float64])
	aliasType(scope, pkg, "_Bool", types.Typ[types.Bool])

	aliasType(scope, pkg, "__builtin_va_list", ctypes.Valist)

	decl_builtin(p)
}

func aliasType(scope *types.Scope, pkg *types.Package, name string, typ types.Type) {
	o := types.NewTypeName(token.NoPos, pkg, name, typ)
	scope.Insert(o)
}

// -----------------------------------------------------------------------------
