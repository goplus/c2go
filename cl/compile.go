package cl

import (
	"go/token"
	"go/types"
	"log"
	"syscall"

	goast "go/ast"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/types/parser"
	"github.com/goplus/gox"

	ctypes "github.com/goplus/c2go/clang/types"
)

const (
	DbgFlagCompileDecl = 1 << iota
	DbgFlagMarkComplicated
	DbgFlagAll = DbgFlagCompileDecl | DbgFlagMarkComplicated
)

var (
	debugCompileDecl     bool
	debugMarkComplicated bool
)

func SetDebug(flags int) {
	debugCompileDecl = (flags & DbgFlagCompileDecl) != 0
	debugMarkComplicated = (flags & DbgFlagMarkComplicated) != 0
}

type node struct {
	pos token.Pos
	end token.Pos
	ctx *blockCtx
}

func (p *node) Pos() token.Pos {
	return p.pos
}

func (p *node) End() token.Pos {
	return p.end
}

func (ctx *blockCtx) goNode(v *ast.Node) goast.Node {
	if rg := v.Range; rg != nil && ctx.file != nil {
		base := ctx.file.Base()
		pos := token.Pos(int(rg.Begin.Offset) + base)
		end := token.Pos(int(rg.End.Offset) + rg.End.TokLen + base)
		return &node{pos: pos, end: end, ctx: ctx}
	}
	return nil
}

func (ctx *blockCtx) goNodePos(v *ast.Node) token.Pos {
	if rg := v.Range; rg != nil && ctx.file != nil {
		base := ctx.file.Base()
		return token.Pos(int(rg.Begin.Offset) + base)
	}
	return token.NoPos
}

// -----------------------------------------------------------------------------

type nodeInterp struct {
	fset *token.FileSet
}

func (p *nodeInterp) Position(start token.Pos) token.Position {
	return p.fset.Position(start)
}

func (p *nodeInterp) Caller(v goast.Node) string {
	return "the function call"
}

func (p *nodeInterp) LoadExpr(v goast.Node) (code string, pos token.Position) {
	src := v.(*node)
	ctx := src.ctx
	start := src.Pos()
	n := int(src.End() - start)
	pos = ctx.file.Position(start)
	code = string(ctx.src[pos.Offset : pos.Offset+n])
	return
}

// -----------------------------------------------------------------------------

type PkgInfo struct {
	typdecls map[string]*gox.TypeDecl
	extfns   map[string]none // external functions which are used
}

type Package struct {
	*gox.Package
	pi *PkgInfo
}

// IsValid returns is this package instance valid or not.
func (p Package) IsValid() bool {
	return p.Package != nil
}

type Reused struct {
	pkg    Package
	exists map[string]none
	base   int
}

// Pkg returns the shared package instance.
func (p *Reused) Pkg() Package {
	return p.pkg
}

type Config struct {
	// Fset provides source position information for syntax trees and types.
	// If Fset is nil, Load will use a new fileset, but preserve Fset's value.
	Fset *token.FileSet

	// An Importer resolves import paths to Packages.
	Importer types.Importer

	// SrcFile specifies a *.i (not *.c) source file path.
	SrcFile string

	// Src specifies source code of SrcFile. Will read from SrcFile if nil.
	Src []byte

	// Public specifies all public C names and their corresponding Go names.
	Public map[string]string

	// Reused specifies to reuse the Package instance between processing multiple C source files.
	*Reused

	// NeedPkgInfo allows to check dependencies and write them to c2go_autogen.go file.
	NeedPkgInfo bool
}

const (
	headerGoFile = "c2go_header.i.go"
)

func NewPackage(pkgPath, pkgName string, file *ast.Node, conf *Config) (pkg Package, err error) {
	if reused := conf.Reused; reused != nil && reused.pkg.Package != nil {
		pkg = reused.pkg
	} else {
		interp := &nodeInterp{}
		confGox := &gox.Config{
			Fset:            conf.Fset,
			Importer:        conf.Importer,
			LoadNamed:       nil,
			HandleErr:       nil,
			NewBuiltin:      nil,
			NodeInterpreter: interp,
			CanImplicitCast: implicitCast,
			DefaultGoFile:   headerGoFile,
		}
		pkg.Package = gox.NewPackage(pkgPath, pkgName, confGox)
		interp.fset = pkg.Fset
	}
	pkg.SetVarRedeclarable(true)
	pkg.pi, err = loadFile(pkg.Package, conf, file)
	return
}

func implicitCast(pkg *gox.Package, V, T types.Type, pv *gox.Element) bool {
	switch t := T.(type) {
	case *types.Basic:
		/* TODO:
		if vt.Kind() == types.UntypedInt {
			if ival, ok := constant.Int64Val(v.CVal); ok && ival == 0 { // nil
				v.Val, v.Type = &ast.Ident{Name: "nil"}, types.Typ[types.UntypedNil]
				break
			}
		}
		*/
		if (t.Info() & types.IsUntyped) != 0 { // untyped
			return false
		}
		if (t.Info() & types.IsInteger) != 0 { // int type
			if e, ok := gox.CastFromBool(pkg.CB(), T, pv); ok {
				pv.Type, pv.Val = T, e.Val
				return true
			}
			if v, ok := V.(*types.Basic); ok && (v.Info()&types.IsInteger) != 0 { // int => int
				e := pkg.CB().Typ(T).Val(pv).Call(1).InternalStack().Pop()
				pv.Type, pv.Val = T, e.Val
				return true
			}
		}
	case *types.Pointer:
		return false
	}
	log.Panicln("==> implicitCast:", V, "to:", T)
	return false
}

// -----------------------------------------------------------------------------

func loadFile(p *gox.Package, conf *Config, file *ast.Node) (pi *PkgInfo, err error) {
	if file.Kind != ast.TranslationUnitDecl {
		return nil, syscall.EINVAL
	}
	ctx := &blockCtx{
		pkg: p, cb: p.CB(), fset: p.Fset,
		unnameds: make(map[ast.ID]*types.Named),
		gblvars:  make(map[string]*gox.VarDefs),
		public:   conf.Public,
		srcfile:  conf.SrcFile,
		src:      conf.Src,
	}
	ctx.initMultiFileCtl(p, conf)
	ctx.initCTypes()
	ctx.initFile()
	compileDeclStmt(ctx, file, true)
	if conf.NeedPkgInfo {
		pi = ctx.genPkgInfo()
	}
	return
}

func compileDeclStmt(ctx *blockCtx, node *ast.Node, global bool) {
	scope := ctx.cb.Scope()
	n := len(node.Inner)
	for i := 0; i < n; i++ {
		decl := node.Inner[i]
		if global {
			ctx.logFile(decl)
			if decl.IsImplicit {
				continue
			}
		}
		switch decl.Kind {
		case ast.VarDecl:
			compileVarDecl(ctx, decl, global)
		case ast.TypedefDecl:
			compileTypedef(ctx, decl, global)
		case ast.RecordDecl:
			name, suKind := ctx.getSuName(decl, decl.TagUsed)
			if global && suKind != suAnonymous && decl.CompleteDefinition && ctx.checkExists(name) {
				continue
			}
			typ := compileStructOrUnion(ctx, name, decl)
			if suKind != suAnonymous {
				break
			} else {
				ctx.unnameds[decl.ID] = typ
			}
			for i+1 < n {
				next := node.Inner[i+1]
				if next.Kind == ast.VarDecl {
					if ret, ok := checkAnonymous(ctx, scope, typ, next); ok {
						compileVarWith(ctx, ret, next)
						i++
						continue
					}
				}
				break
			}
		case ast.EnumDecl:
			compileEnum(ctx, decl)
		case ast.FunctionDecl:
			if global {
				compileFunc(ctx, decl)
				continue
			}
			fallthrough
		default:
			log.Panicln("compileDeclStmt: unknown kind =", decl.Kind)
		}
	}
}

func compileFunc(ctx *blockCtx, fn *ast.Node) {
	fnType := fn.Type
	if debugCompileDecl {
		log.Println("func", fn.Name, "-", fnType.QualType, fn.Loc.PresumedLine)
	}
	var hasName bool
	var params []*types.Var
	var body *ast.Node
	var results *types.Tuple
	for _, item := range fn.Inner {
		switch item.Kind {
		case ast.ParmVarDecl:
			if debugCompileDecl {
				log.Println("  => param", item.Name, "-", item.Type.QualType)
			}
			if item.Name != "" {
				hasName = true
			}
			params = append(params, newParam(ctx, item))
		case ast.CompoundStmt:
			body = item
		case ast.BuiltinAttr, ast.FormatAttr, ast.AsmLabelAttr, ast.AvailabilityAttr, ast.ColdAttr, ast.DeprecatedAttr,
			ast.AlwaysInlineAttr, ast.WarnUnusedResultAttr, ast.NoThrowAttr, ast.NoInlineAttr, ast.AllocSizeAttr,
			ast.NonNullAttr, ast.ConstAttr, ast.PureAttr, ast.GNUInlineAttr, ast.ReturnsTwiceAttr, ast.NoSanitizeAttr,
			ast.RestrictAttr, ast.MSAllocatorAttr, ast.VisibilityAttr, ast.C11NoReturnAttr:
		default:
			log.Panicln("compileFunc: unknown kind =", item.Kind)
		}
	}
	variadic := fn.Variadic
	if variadic {
		params = append(params, newVariadicParam(ctx, hasName))
	}
	pkg := ctx.pkg
	if tyRet := toType(ctx, fnType, parser.FlagGetRetType); ctypes.NotVoid(tyRet) {
		results = types.NewTuple(pkg.NewParam(token.NoPos, "", tyRet))
	}
	sig := gox.NewCSignature(types.NewTuple(params...), results, variadic)
	if body != nil {
		fnName, isMain := fn.Name, false
		if fnName == "main" && (results != nil || params != nil) {
			fnName, isMain = "_cgo_main", true
		} else {
			ctx.getPubName(&fnName)
		}
		f, err := pkg.NewFuncWith(ctx.goNodePos(fn), fnName, sig, nil)
		if err != nil {
			log.Panicln("compileFunc:", err)
		}
		cb := f.BodyStart(pkg)
		ctx.curfn = newFuncCtx(pkg, ctx.markComplicated(fn.Name, body))
		compileSub(ctx, body)
		checkNeedReturn(ctx, body)
		ctx.curfn = nil
		cb.End()
		if isMain {
			pkg.NewFunc(nil, "main", nil, nil, false).BodyStart(pkg)
			if results != nil {
				cb.Val(pkg.Import("os").Ref("Exit")).Typ(types.Typ[types.Int])
			}
			cb.Val(f.Func)
			if params != nil {
				panic("TODO: main func with params")
			}
			cb.Call(len(params))
			if results != nil {
				cb.Call(1).Call(1)
			}
			cb.EndStmt().End()
		} else {
			delete(ctx.extfns, fnName)
		}
	} else if fn.IsUsed {
		f := types.NewFunc(ctx.goNodePos(fn), pkg.Types, fn.Name, sig)
		if pkg.Types.Scope().Insert(f) == nil {
			ctx.extfns[fn.Name] = none{}
		}
	}
}

func (p *blockCtx) getPubName(pfnName *string) (ok bool) {
	goName, ok := p.public[*pfnName]
	if ok {
		if goName != "" {
			*pfnName = goName
		} else {
			*pfnName = title(*pfnName)
		}
	}
	return
}

func title(name string) string {
	if r := name[0]; 'a' <= r && r <= 'z' {
		r -= 'a' - 'A'
		return string(r) + name[1:]
	}
	return name
}

const (
	valistName = "__cgo_args"
)

func compileVAArgExpr(ctx *blockCtx, expr *ast.Node) {
	pkg := ctx.pkg
	ap := expr.Inner[0]
	typ := toType(ctx, expr.Type, 0)
	args := pkg.NewParam(token.NoPos, valistName, ctypes.Valist)
	ret := pkg.NewParam(token.NoPos, "_cgo_ret", typ)
	cb := ctx.cb.NewClosure(types.NewTuple(args), types.NewTuple(ret), false).BodyStart(pkg)
	//
	// func(__cgo_args []any) (_cgo_ret T) {
	//    _cgo_ret = __cgo_args[0].(typ)
	//    ap = __cgo_args[1:]
	//    return
	// }(ap)
	cb.VarRef(ret).Val(args).Val(0).Index(1, false).TypeAssert(typ, false).Assign(1)
	ap = compileValistLHS(ctx, ap)
	cb.Val(args).Val(1).None().Slice(false).Assign(1).Return(0).End()
	compileExpr(ctx, ap)
	cb.Call(1)
}

func newVariadicParam(ctx *blockCtx, hasName bool) *types.Var {
	name := ""
	if hasName {
		name = valistName
	}
	return types.NewParam(token.NoPos, ctx.pkg.Types, name, ctypes.Valist)
}

func newParam(ctx *blockCtx, decl *ast.Node) *types.Var {
	typ := toType(ctx, decl.Type, parser.FlagIsParam)
	avoidKeyword(&decl.Name)
	return types.NewParam(ctx.goNodePos(decl), ctx.pkg.Types, decl.Name, typ)
}

// -----------------------------------------------------------------------------
