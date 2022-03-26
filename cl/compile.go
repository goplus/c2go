package cl

import (
	"go/token"
	"go/types"
	"log"
	"syscall"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/gox"
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

func logFile(node *ast.Node) {
	if debugCompileDecl {
		if f := node.Loc.PresumedFile; f != "" {
			log.Println("==>", f)
		}
	}
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
	confGox := &gox.Config{
		Fset:            conf.Fset,
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

type blockCtx struct {
	pkg *gox.Package
	cb  *gox.CodeBuilder
}

func loadFile(p *gox.Package, file *ast.Node) (err error) {
	if file.Kind != ast.TranslationUnitDecl {
		return syscall.EINVAL
	}
	ctx := &blockCtx{pkg: p, cb: p.CB()}
	for _, decl := range file.Inner {
		logFile(decl)
		if decl.IsImplicit {
			continue
		}
		switch decl.Kind {
		case ast.FunctionDecl:
			compileFunc(ctx, decl)
		case ast.TypedefDecl:
			compileTypedef(ctx, decl)
		case ast.RecordDecl:
			compileStructOrUnion(ctx, decl)
		case ast.VarDecl:
			compileVar(ctx, decl)
		default:
			log.Fatalln("loadFile: unknown kind =", decl.Kind)
		}
	}
	return
}

func compileFunc(ctx *blockCtx, fn *ast.Node) {
	if debugCompileDecl {
		log.Println("func", fn.Name, "-", fn.Type.QualType, fn.Loc.PresumedLine)
	}
	for _, item := range fn.Inner {
		switch item.Kind {
		case ast.ParmVarDecl:
			if debugCompileDecl {
				log.Println("  => param", item.Name, "-", item.Type.QualType)
			}
		case ast.CompoundStmt:
		case ast.BuiltinAttr:
		case ast.FormatAttr:
		case ast.AsmLabelAttr:
		case ast.AvailabilityAttr:
		case ast.ColdAttr:
		case ast.DeprecatedAttr:
		case ast.AlwaysInlineAttr:
		default:
			log.Fatalln("compileFunc: unknown kind =", item.Kind)
		}
	}
}

// -----------------------------------------------------------------------------
