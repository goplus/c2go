package cl

import (
	"go/token"
	"go/types"
	"io"
	"sort"
	"strings"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

const (
	depsFile = "deps"
)

func (p Package) InitDependencies() {
	if p.pi == nil {
		panic("Please set conf.NeedPkgInfo = true")
	}
	if _, ok := p.File(depsFile); ok {
		return
	}

	var uds []string
	for name, tdecl := range p.pi.typdecls {
		if tdecl.State() == gox.TyStateUninited {
			uds = append(uds, name)
		}
	}
	sort.Strings(uds)
	extfns := make([]string, 0, len(p.pi.extfns))
	for name := range p.pi.extfns {
		extfns = append(extfns, name)
	}
	sort.Strings(extfns)

	pkg := p.Package
	old, _ := pkg.SetCurFile(depsFile, true)
	defer pkg.RestoreCurFile(old)

	scope := pkg.Types.Scope()
	empty := types.NewStruct(nil, nil)
	for _, us := range uds {
		p.pi.typdecls[us].InitType(pkg, empty)
	}
	vPanic := types.Universe.Lookup("panic")
	for _, uf := range extfns {
		sig := scope.Lookup(uf).Type().(*types.Signature)
		f, _ := pkg.NewFuncWith(token.NoPos, uf, sig, nil)
		f.BodyStart(pkg).
			Val(vPanic).Val("notimpl").Call(1).EndStmt().
			End()
	}
}

func (p Package) WriteDepTo(dst io.Writer) error {
	p.InitDependencies()
	return gox.WriteTo(dst, p.Package, depsFile)
}

func (p Package) WriteDepFile(file string) error {
	p.InitDependencies()
	return gox.WriteFile(file, p.Package, depsFile)
}

// -----------------------------------------------------------------------------

func initPublicFrom(conf *Config, node *ast.Node) {
	pubFrom := conf.PublicFrom
	if len(pubFrom) == 0 {
		return
	}
	if conf.Public == nil {
		conf.Public = make(map[string]string)
	}
	public := conf.Public
	isPub := false
	for _, decl := range node.Inner {
		if f := decl.Loc.PresumedFile; f != "" {
			isPub = isPublicFrom(f, pubFrom)
		}
		if isPub {
			switch decl.Kind {
			case ast.VarDecl, ast.TypedefDecl, ast.FunctionDecl:
				if canPub(decl.Name) {
					public[decl.Name] = ""
				}
			}
		}
	}
}

func isPublicFrom(f string, pubFrom []string) bool {
	for _, from := range pubFrom {
		if strings.HasSuffix(f, from) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
