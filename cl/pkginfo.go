package cl

import (
	"go/types"
	"io"
	"sort"

	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

type PkgInfo struct {
	UndefinedStructs []string

	confGox *gox.Config
}

// -----------------------------------------------------------------------------

func (p *blockCtx) genPkgInfo(confGox *gox.Config) *PkgInfo {
	var undefinedStructs []string
	for name, tdecl := range p.typdecls {
		if !tdecl.Inited() {
			undefinedStructs = append(undefinedStructs, name)
		}
	}
	sort.Strings(undefinedStructs)
	return &PkgInfo{UndefinedStructs: undefinedStructs, confGox: confGox}
}

// -----------------------------------------------------------------------------

func (p Package) Dependencies() *gox.Package {
	pkg := gox.NewPackage("", p.Types.Name(), p.confGox)
	empty := types.NewStruct(nil, nil)
	for _, us := range p.UndefinedStructs {
		pkg.NewType(us).InitType(pkg, empty)
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
