package cl

import (
	"go/token"
	"go/types"
	"io"
	"sort"

	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

type PkgInfo struct {
	UndefinedStructs []string
	UsedFuncs        []string

	confGox *gox.Config
}

// -----------------------------------------------------------------------------

func (p *blockCtx) genPkgInfo(confGox *gox.Config) *PkgInfo {
	var uds []string
	for name, tdecl := range p.typdecls {
		if !tdecl.Inited() {
			uds = append(uds, name)
		}
	}
	sort.Strings(uds)
	extfns := make([]string, 0, len(p.extfns))
	for name := range p.extfns {
		extfns = append(extfns, name)
	}
	sort.Strings(extfns)
	return &PkgInfo{UndefinedStructs: uds, UsedFuncs: extfns, confGox: confGox}
}

// -----------------------------------------------------------------------------

func (p Package) Dependencies() *gox.Package {
	pkg := gox.NewPackage("", p.Types.Name(), p.confGox)
	scope := p.Types.Scope()
	me, old := *pkg.Types, *p.Types
	pkg.Types = p.Types
	*p.Types = me
	defer func() {
		*p.Types = old
	}()
	empty := types.NewStruct(nil, nil)
	for _, us := range p.UndefinedStructs {
		pkg.NewType(us).InitType(pkg, empty)
	}
	vPanic := types.Universe.Lookup("panic")
	for _, uf := range p.UsedFuncs {
		sig := scope.Lookup(uf).Type().(*types.Signature)
		f, _ := pkg.NewFuncWith(token.NoPos, uf, sig, nil)
		f.BodyStart(pkg).
			Val(vPanic).Val("notimpl").Call(1).EndStmt().
			End()
	}
	return pkg
}

func (p Package) WriteDepTo(dst io.Writer) error {
	pkg := p.Dependencies()
	return gox.WriteTo(dst, pkg, false)
}

func (p Package) WriteDepFile(file string) error {
	pkg := p.Dependencies()
	return gox.WriteFile(file, pkg, false)
}

// -----------------------------------------------------------------------------
