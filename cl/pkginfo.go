package cl

import (
	"go/token"
	"go/types"
	"io"
	"sort"

	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

func (p *blockCtx) genPkgInfo() *PkgInfo {
	return &p.PkgInfo
}

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
		if !tdecl.Inited() {
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
