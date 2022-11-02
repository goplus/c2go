package cl

import (
	"go/token"
	"go/types"
	"log"
	"path/filepath"
	"syscall"

	goast "go/ast"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/cmod"
	"github.com/goplus/c2go/clang/types/parser"
	"github.com/goplus/gox"

	ctypes "github.com/goplus/c2go/clang/types"
)

const (
	DbgFlagCompileDecl = 1 << iota
	DbgFlagMarkComplicated
	DbgFlagLoadDeps
	DbgFlagAll = DbgFlagCompileDecl | DbgFlagMarkComplicated | DbgFlagLoadDeps
)

var (
	debugCompileDecl     bool
	debugLoadDeps        bool
	debugMarkComplicated bool
)

func SetDebug(flags int) {
	debugCompileDecl = (flags & DbgFlagCompileDecl) != 0
	debugLoadDeps = (flags & DbgFlagLoadDeps) != 0
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

func goPos(v goast.Node) token.Pos {
	if v == nil {
		return token.NoPos
	}
	return v.Pos()
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

type BFMode int8

const (
	BFM_Default  BFMode = iota
	BFM_InLibC          // define builtin functions in libc
	BFM_FromLibC        // import builtin functions from libc
)

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
	pkg     Package
	exists  map[string]none
	autopub map[string]none
	base    int
	deps    depPkgs
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

	// Ignored specifies all ignored symbols (types, functions, etc).
	Ignored []string

	// Reused specifies to reuse the Package instance between processing multiple C source files.
	*Reused

	// Dir specifies root directory of a c2go project (where there is a c2go.cfg file).
	Dir string

	// ProcDepPkg specifies how to process a dependent package.
	// If ProcDepPkg is nil, it means nothing to do.
	// It's for compiling depended pkgs (to gen c2go.a.pub file) if needed.
	ProcDepPkg func(depPkgDir string)

	// Deps specifies all dependent packages.
	Deps []*cmod.Package

	// Include specifies include searching directories.
	Include []string

	// Public specifies all public C names and their corresponding Go names.
	Public map[string]string

	// PublicFrom specifies header files to fetch public symbols.
	PublicFrom []string

	// BuiltinFuncMode sets compiling mode of builtin functions.
	BuiltinFuncMode BFMode

	// SkipLibcHeader specifies to ignore standard library headers.
	SkipLibcHeader bool

	// NeedPkgInfo allows to check dependencies and write them to c2go_autogen.go file.
	NeedPkgInfo bool

	// TestMain specifies to generate TestMain func as entry, not main func.
	TestMain bool
}

const (
	headerGoFile = "c2go_header.i.go"
)

// NewPackage create a Go package from C file AST.
// If conf.Reused isn't nil, it shares the Go package instance in multi C files.
// Otherwise it creates a single Go file in the Go package.
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
	srcFile := conf.SrcFile
	if srcFile != "" {
		srcFile, _ = filepath.Abs(srcFile)
	}
	ctx := &blockCtx{
		pkg: p, cb: p.CB(), fset: p.Fset,
		unnameds: make(map[ast.ID]unnamedType),
		srcenums: make(map[ast.ID]string),
		gblvars:  make(map[string]*gox.VarDefs),
		ignored:  conf.Ignored,
		public:   conf.Public,
		srcdir:   filepath.Dir(srcFile),
		srcfile:  srcFile,
		src:      conf.Src,
		bfm:      conf.BuiltinFuncMode,
		testMain: conf.TestMain,
	}
	baseDir, _ := filepath.Abs(conf.Dir)
	ctx.initMultiFileCtl(p, baseDir, conf)
	ctx.initCTypes()
	ctx.initFile()
	ctx.initPublicFrom(baseDir, conf, file)
	for _, ign := range ctx.ignored {
		if ctx.getPubName(&ign) {
			ctx.ignored = append(ctx.ignored, ign)
		}
	}
	compileDeclStmt(ctx, file, true)
	if conf.NeedPkgInfo {
		pkgInfo := ctx.PkgInfo // make a copy: don't keep a ref to blockCtx
		pi = &pkgInfo
	}
	return
}

// NOTE: call isPubTypedef only in global scope
func isPubTypedef(ctx *blockCtx, decl *ast.Node) bool {
	if decl.Kind == ast.TypedefDecl {
		name := decl.Name
		return ctx.getPubName(&name)
	}
	return false
}

func compileDeclStmt(ctx *blockCtx, node *ast.Node, global bool) {
	scope := ctx.cb.Scope()
	n := len(node.Inner)
	for i := 0; i < n; i++ {
		decl := node.Inner[i]
		if global {
			ctx.logFile(decl)
			if decl.IsImplicit || ctx.inDepPkg {
				continue
			}
		}
		switch decl.Kind {
		case ast.VarDecl:
			compileVarDecl(ctx, decl, global)
		case ast.TypedefDecl:
			origName, pub := decl.Name, false
			if global {
				pub = ctx.getPubName(&decl.Name)
			}
			compileTypedef(ctx, decl, global, pub)
			if pub {
				substObj(ctx.pkg.Types, scope, origName, scope.Lookup(decl.Name))
			}
		case ast.RecordDecl:
			pub := false
			name, suKind := ctx.getSuName(decl, decl.TagUsed)
			origName := name
			if global {
				if suKind == suAnonymous {
					// pub = true if this is a public typedef
					pub = i+1 < n && isPubTypedef(ctx, node.Inner[i+1])
				} else {
					pub = ctx.getPubName(&name)
					if decl.CompleteDefinition && ctx.checkExists(name) {
						continue
					}
				}
			}
			typ, del := compileStructOrUnion(ctx, name, decl, pub)
			if suKind != suAnonymous {
				if pub {
					substObj(ctx.pkg.Types, scope, origName, scope.Lookup(name))
				}
				break
			}
			ctx.unnameds[decl.ID] = unnamedType{typ: typ, del: del}
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
			compileEnum(ctx, decl, global)
		case ast.EmptyDecl:
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
	fnName, fnType := fn.Name, fn.Type
	if debugCompileDecl {
		log.Println("func", fnName, "-", fnType.QualType, fn.Loc.PresumedLine)
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
			ast.RestrictAttr, ast.MSAllocatorAttr, ast.VisibilityAttr, ast.C11NoReturnAttr, ast.StrictFPAttr:
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
	origName, rewritten := fnName, false
	if !ctx.inHeader && fn.StorageClass == ast.Static {
		fnName, rewritten = ctx.autoStaticName(origName), true
	} else {
		rewritten = ctx.getPubName(&fnName)
	}
	if body != nil {
		if ctx.checkExists(fnName) {
			return
		}
		isMain := false
		if fnName == "main" && (results != nil || params != nil) {
			fnName, isMain = "_cgo_main", true
		}
		f, err := pkg.NewFuncWith(ctx.goNodePos(fn), fnName, sig, nil)
		if err != nil {
			log.Panicln("compileFunc:", err)
		}
		if rewritten { // for fnName is a recursive function
			scope := pkg.Types.Scope()
			substObj(pkg.Types, scope, origName, f.Func)
			rewritten = false
		}
		cb := f.BodyStart(pkg)
		ctx.curfn = newFuncCtx(pkg, ctx.markComplicated(fnName, body), origName)
		compileSub(ctx, body)
		checkNeedReturn(ctx, body)
		ctx.curfn = nil
		cb.End()
		if isMain {
			var t *types.Var
			var entryParams *types.Tuple
			var entry = "main"
			var testMain = ctx.testMain
			if testMain {
				entry = "TestMain"
				testing := pkg.Import("testing")
				t = pkg.NewParam(token.NoPos, "t", types.NewPointer(testing.Ref("T").Type()))
				entryParams = types.NewTuple(t)
			}
			pkg.NewFunc(nil, entry, entryParams, nil, false).BodyStart(pkg)
			if results != nil {
				if testMain {
					// if _cgo_ret := _cgo_main(); _cgo_ret != 0 {
					//   t.Fatal("exit status", _cgo_ret)
					// }
					cb.If().DefineVarStart(token.NoPos, retName)
				} else {
					// os.Exit(int(_cgo_main()))
					cb.Val(pkg.Import("os").Ref("Exit")).Typ(types.Typ[types.Int])
				}
			}
			cb.Val(f.Func)
			if params != nil {
				panic("TODO: main func with params")
			}
			cb.Call(len(params))
			if results != nil {
				if testMain {
					cb.EndInit(1)
					ret := cb.Scope().Lookup(retName)
					cb.Val(ret).Val(0).BinaryOp(token.NEQ).Then().
						Val(t).MemberVal("Fatal").Val("exit status").Val(ret).Call(2).EndStmt().
						End()
				} else {
					cb.Call(1).Call(1)
				}
			}
			cb.EndStmt().End()
		} else {
			delete(ctx.extfns, fnName)
		}
	} else if fn.IsUsed {
		f := types.NewFunc(ctx.goNodePos(fn), pkg.Types, fnName, sig)
		if pkg.Types.Scope().Insert(f) == nil {
			ctx.addExternFunc(fnName)
		}
	}
	if rewritten {
		scope := pkg.Types.Scope()
		substObj(pkg.Types, scope, origName, scope.Lookup(fnName))
	}
}

func (p *blockCtx) getPubName(pfnName *string) (ok bool) {
	name := *pfnName
	goName, ok := p.public[name]
	if ok {
		if goName == "" {
			goName = gox.CPubName(name)
		}
	} else if _, ok = p.autopub[name]; ok {
		p.public[name] = ""
		goName = gox.CPubName(name)
	} else {
		return
	}
	*pfnName = goName
	return goName != name
}

const (
	valistName = "__cgo_args"
	retName    = "_cgo_ret"
	tagName    = "_cgo_tag"
)

func compileVAArgExpr(ctx *blockCtx, expr *ast.Node) {
	pkg := ctx.pkg
	ap := expr.Inner[0]
	typ := toType(ctx, expr.Type, 0)
	args := pkg.NewParam(token.NoPos, valistName, ctypes.Valist)
	ret := pkg.NewParam(token.NoPos, retName, typ)
	cb := ctx.cb.NewClosure(types.NewTuple(args), types.NewTuple(ret), false).BodyStart(pkg)
	//
	// func(__cgo_args []any) (_cgo_ret T) {
	//    ...
	//    ap = __cgo_args[1:]
	//    return
	// }(ap)
	if isNilComparable(typ) {
		// _cgo_ret = typ((
		//	 (*[2]unsafe.Pointer)(unsafe.Pointer(&__cgo_args[0]))
		// )[1])
		cb.VarRef(ret).Typ(typ).Typ(tyPVPA).Typ(ctypes.UnsafePointer).
			Val(args).Val(0).IndexRef(1).UnaryOp(token.AND).
			Call(1).Call(1).Val(1).Index(1, false).
			Call(1).Assign(1)
	} else if t, ok := typ.(*types.Basic); ok && isNormalInteger(t) {
		// switch _cgo_tag := __cgo_args[0].(type) {
		// case typ:
		//   _go_ret = _cgo_tag
		// case typ2:
		//   _cgo_ret = typ(_cgo_tag)
		// }
		var typ2 types.Type
		if t.Kind() <= types.Int64 { // int
			typ2 = types.Typ[t.Kind()+(types.Uint-types.Int)]
		} else { // uint
			typ2 = types.Typ[t.Kind()-(types.Uint-types.Int)]
		}
		cb.TypeSwitch(tagName).Val(args).Val(0).Index(1, false).TypeAssertThen()

		cb.Typ(typ).TypeCase(1)
		cb.VarRef(ret).Val(cb.Scope().Lookup(tagName)).Assign(1).End()

		cb.Typ(typ2).TypeCase(1)
		cb.VarRef(ret).Typ(typ).Val(cb.Scope().Lookup(tagName)).Call(1).Assign(1).End()

		if t.Kind() == types.Int32 {
			cb.Typ(types.Typ[types.Uint64]).TypeCase(1)
			cb.VarRef(ret).Typ(typ).Val(cb.Scope().Lookup(tagName)).Call(1).Assign(1).End()
		}
		cb.End() // end switch
	} else {
		// _cgo_ret = __cgo_args[0].(typ)
		cb.VarRef(ret).Val(args).Val(0).Index(1, false).TypeAssert(typ, false).Assign(1)
	}
	ap = compileValistLHS(ctx, ap)
	cb.Val(args).Val(1).None().Slice(false).Assign(1).Return(0).End()
	compileExpr(ctx, ap)
	cb.Call(1)
}

var (
	tyPVPA = types.NewPointer(types.NewArray(ctypes.UnsafePointer, 2))
)

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
