package cmod

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/goplus/c2go/clang/pathutil"
	"github.com/goplus/mod/gopmod"
	"github.com/qiniu/x/errors"
)

var (
	ErrGoModNotFound = errors.New("go.mod not found")
)

func LoadDeps(dir string, deps []string) (pkgs []*Package, err error) {
	mod, err := gopmod.Load(dir)
	if err != nil {
		err = errors.NewWith(err, `gopmod.Load(dir, 0)`, -2, "gopmod.Load", dir, 0)
		return
	}
	return Imports(mod, deps)
}

type Module = gopmod.Module

func Imports(mod *Module, pkgPaths []string) (pkgs []*Package, err error) {
	pkgs = make([]*Package, len(pkgPaths))
	for i, pkgPath := range pkgPaths {
		pkgs[i], err = Import(mod, pkgPath)
		if err != nil {
			err = errors.NewWith(err, `Import(mod, pkgPath)`, -2, "cmod.Import", mod, pkgPath)
			return
		}
	}
	return
}

type Package struct {
	*gopmod.Package
	Path    string   // package path
	Dir     string   // absolue local path of the package
	Include []string // absolute include paths
}

func Import(mod *Module, pkgPath string) (p *Package, err error) {
	if mod == nil {
		return nil, ErrGoModNotFound
	}
	pkg, err := mod.Lookup(pkgPath)
	if err != nil {
		err = errors.NewWith(err, `mod.Lookup(pkgPath)`, -2, "(*gopmod.Module).Lookup", pkgPath)
		return
	}
	pkgDir, err := filepath.Abs(pkg.Dir)
	if err != nil {
		err = errors.NewWith(err, `filepath.Abs(pkg.Dir)`, -2, "filepath.Abs", pkg.Dir)
		return
	}
	pkgIncs, err := findIncludeDirs(pkgDir)
	if err != nil {
		err = errors.NewWith(err, `findIncludeDirs(pkgDir)`, -2, "cmod.findIncludeDirs", pkgDir)
		return
	}
	for i, dir := range pkgIncs {
		pkgIncs[i] = pathutil.Canonical(pkgDir, dir)
	}
	return &Package{Package: pkg, Path: pkgPath, Dir: pkgDir, Include: pkgIncs}, nil
}

func findIncludeDirs(pkgDir string) (incs []string, err error) {
	var conf struct {
		Include []string `json:"include"`
	}
	file := filepath.Join(pkgDir, "c2go.cfg")
	b, err := os.ReadFile(file)
	if err != nil {
		err = errors.NewWith(err, `os.ReadFile(file)`, -2, "os.ReadFile", file)
		return
	}
	err = json.Unmarshal(b, &conf)
	if err != nil {
		err = errors.NewWith(err, `json.Unmarshal(b, &conf)`, -2, "json.Unmarshal", b, &conf)
		return
	}
	return conf.Include, nil
}
