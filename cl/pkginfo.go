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

	typdecls map[string]*gox.TypeDecl
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
	return &PkgInfo{UndefinedStructs: uds, UsedFuncs: extfns, typdecls: p.typdecls}
}

// -----------------------------------------------------------------------------

const (
	depsFile = "deps"
)

func (p Package) InitDependencies() {
	if p.PkgInfo == nil {
		panic("Please set conf.NeedPkgInfo = true")
	}
	if _, ok := p.File(depsFile); ok {
		return
	}

	pkg := p.Package
	old, _ := pkg.SetCurFile(depsFile, true)
	defer pkg.RestoreCurFile(old)

	scope := pkg.Types.Scope()
	empty := types.NewStruct(nil, nil)
	for _, us := range p.UndefinedStructs {
		p.typdecls[us].InitType(pkg, empty)
	}
	vPanic := types.Universe.Lookup("panic")
	for _, uf := range p.UsedFuncs {
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
