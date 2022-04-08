package cl

import (
	"go/token"
	"go/types"
	"log"
	"strings"
	"syscall"

	goast "go/ast"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/types/parser"
	"github.com/goplus/gox"

	ctypes "github.com/goplus/c2go/clang/types"
)

const (
	DbgFlagCompileDecl = 1 << iota
	DbgFlagAll         = DbgFlagCompileDecl
)

var (
	debugCompileDecl bool
)

func SetDebug(flags int) {
	debugCompileDecl = (flags & DbgFlagCompileDecl) != 0
}

func logFile(ctx *blockCtx, node *ast.Node) {
	if f := node.Loc.PresumedFile; f != "" {
		ctx.curfile = f
		if strings.HasSuffix(f, ".c") {
			ctx.curfile += ".i"
		}
		if debugCompileDecl {
			log.Println("==>", f)
		}
	}
}

func goNode(node *ast.Node) goast.Node {
	return nil // TODO:
}

func goNodePos(node *ast.Node) token.Pos {
	return token.NoPos // TODO:
}

// -----------------------------------------------------------------------------

type Config struct {
	// Fset provides source position information for syntax trees and types.
	// If Fset is nil, Load will use a new fileset, but preserve Fset's value.
	Fset *token.FileSet

	// An Importer resolves import paths to Packages.
	Importer types.Importer
}

func NewPackage(pkgPath, pkgName string, file *ast.Node, conf *Config) (p *gox.Package, err error) {
	fset := conf.Fset
	if fset == nil {
		fset = token.NewFileSet()
	}
	confGox := &gox.Config{
		Fset:            fset,
		Importer:        conf.Importer,
		LoadNamed:       nil,
		HandleErr:       nil,
		NodeInterpreter: nil,
		NewBuiltin:      nil,
	}
	p = gox.NewPackage(pkgPath, pkgName, confGox)
	err = loadFile(p, file)
	return
}

// -----------------------------------------------------------------------------

func loadFile(p *gox.Package, file *ast.Node) error {
	if file.Kind != ast.TranslationUnitDecl {
		return syscall.EINVAL
	}
	ctx := &blockCtx{
		pkg: p, cb: p.CB(), fset: p.Fset,
		unnameds: make(map[ast.ID]*types.Named),
		uncompls: make(map[string]*gox.TypeDecl),
		files:    make(map[string]source),
	}
	ctx.initCTypes()
	compileDeclStmt(ctx, file, true)
	return nil
}

func compileDeclStmt(ctx *blockCtx, node *ast.Node, global bool) {
	scope := ctx.cb.Scope()
	n := len(node.Inner)
	for i := 0; i < n; i++ {
		decl := node.Inner[i]
		if global {
			logFile(ctx, decl)
			if decl.IsImplicit {
				continue
			}
		}
		switch decl.Kind {
		case ast.VarDecl:
			compileVarDecl(ctx, decl)
		case ast.TypedefDecl:
			compileTypedef(ctx, decl)
		case ast.RecordDecl:
			name, anonymous := ctx.getAsuName(decl, "")
			typ := compileStructOrUnion(ctx, name, decl)
			if !anonymous {
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
			log.Fatalln("compileDeclStmt: unknown kind =", decl.Kind)
		}
	}
}

func compileFunc(ctx *blockCtx, fn *ast.Node) {
	fnType := fn.Type
	if debugCompileDecl {
		log.Println("func", fn.Name, "-", fnType.QualType, fn.Loc.PresumedLine)
	}
	var variadic, hasName bool
	var params []*types.Var
	var results *types.Tuple
	var body *ast.Node
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
			ast.ConstAttr:
		default:
			log.Fatalln("compileFunc: unknown kind =", item.Kind)
		}
	}
	if variadic = isVariadicFn(fnType); variadic {
		params = append(params, newVariadicParam(ctx, hasName))
	} else {
		variadic = checkVariadic(ctx, params, hasName)
	}
	pkg := ctx.pkg
	if t := toType(ctx, fnType, parser.FlagGetRetType); ctypes.NotVoid(t) {
		ret := types.NewParam(token.NoPos, pkg.Types, "", t)
		results = types.NewTuple(ret)
	}
	sig := gox.NewSignature(nil, types.NewTuple(params...), results, variadic)
	if body != nil {
		fnName, isMain := fn.Name, false
		if fnName == "main" && (results != nil || params != nil) {
			fnName, isMain = "_cgo_main", true
		}
		f, err := pkg.NewFuncWith(goNodePos(fn), fnName, sig, nil)
		if err != nil {
			log.Fatalln("compileFunc:", err)
		}
		cb := f.BodyStart(pkg)
		ctx.curfn = newFuncCtx()
		compileCompoundStmt(ctx, body)
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
		}
	} else {
		f := types.NewFunc(goNodePos(fn), pkg.Types, fn.Name, sig)
		pkg.Types.Scope().Insert(f)
	}
}

const (
	valistName = "__cgo_args"
)

func newVariadicParam(ctx *blockCtx, hasName bool) *types.Var {
	name := ""
	if hasName {
		name = valistName
	}
	return types.NewParam(token.NoPos, ctx.pkg.Types, name, types.NewSlice(gox.TyEmptyInterface))
}

func newParam(ctx *blockCtx, decl *ast.Node) *types.Var {
	typ := toType(ctx, decl.Type, parser.FlagIsParam)
	return types.NewParam(goNodePos(decl), ctx.pkg.Types, decl.Name, typ)
}

func checkVariadic(ctx *blockCtx, params []*types.Var, hasName bool) bool {
	n := len(params)
	if n > 0 {
		if last := params[n-1]; isValistType(ctx, last.Type()) {
			params[n-1] = newVariadicParam(ctx, hasName)
			return true
		}
	}
	return false
}

func isValistType(ctx *blockCtx, t types.Type) bool {
	return types.Identical(t, ctx.tyValist)
}

// -----------------------------------------------------------------------------
