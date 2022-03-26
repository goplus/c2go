package cl

import (
	"go/token"
	"go/types"
	"log"
	"syscall"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/gox"
)

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
	_ = ctx
	for _, decl := range file.Inner {
		if decl.IsImplicit {
			continue
		}
		switch decl.Kind {
		default:
			log.Fatalln("loadFile: unknown kind =", decl.Kind)
		}
	}
	return
}

// -----------------------------------------------------------------------------
