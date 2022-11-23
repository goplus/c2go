/*
 Copyright 2022 The GoPlus Authors (goplus.org)
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at
     http://www.apache.org/licenses/LICENSE-2.0
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package packages

import (
	"bytes"
	"errors"
	"go/token"
	"go/types"
	"os"
	"os/exec"

	"golang.org/x/tools/go/gcexportdata"
)

// ----------------------------------------------------------------------------

type Importer struct {
	loaded map[string]*types.Package
	fset   *token.FileSet
	tags   string
	dir    string
}

// NewImporter creates an Importer object that meets types.Importer interface.
func NewImporter(fset *token.FileSet, tags string, workDir ...string) *Importer {
	dir := ""
	if len(workDir) > 0 {
		dir = workDir[0]
	}
	if fset == nil {
		fset = token.NewFileSet()
	}
	loaded := make(map[string]*types.Package)
	loaded["unsafe"] = types.Unsafe
	return &Importer{loaded: loaded, fset: fset, tags: tags, dir: dir}
}

func (p *Importer) Import(pkgPath string) (pkg *types.Package, err error) {
	return p.ImportFrom(pkgPath, p.dir, 0)
}

// ImportFrom returns the imported package for the given import
// path when imported by a package file located in dir.
// If the import failed, besides returning an error, ImportFrom
// is encouraged to cache and return a package anyway, if one
// was created. This will reduce package inconsistencies and
// follow-on type checker errors due to the missing package.
// The mode value must be 0; it is reserved for future use.
// Two calls to ImportFrom with the same path and dir must
// return the same package.
func (p *Importer) ImportFrom(pkgPath, dir string, mode types.ImportMode) (*types.Package, error) {
	if ret, ok := p.loaded[pkgPath]; ok && ret.Complete() {
		return ret, nil
	}
	expfile, err := FindExport(dir, pkgPath, p.tags)
	if err != nil {
		return nil, err
	}
	return p.loadByExport(expfile, pkgPath)
}

func (p *Importer) loadByExport(expfile string, pkgPath string) (pkg *types.Package, err error) {
	f, err := os.Open(expfile)
	if err != nil {
		return
	}
	defer f.Close()

	r, err := gcexportdata.NewReader(f)
	if err == nil {
		pkg, err = gcexportdata.Read(r, p.fset, p.loaded, pkgPath)
	}
	return
}

// ----------------------------------------------------------------------------

// FindExport lookups export file (.a) of a package by its pkgPath.
// It returns empty if pkgPath not found.
func FindExport(dir, pkgPath string, tags string) (expfile string, err error) {
	data, err := golistExport(dir, pkgPath, tags)
	if err != nil {
		return
	}
	expfile = string(bytes.TrimSuffix(data, []byte{'\n'}))
	return
}

func golistExport(dir, pkgPath string, tags string) (ret []byte, err error) {
	var stdout, stderr bytes.Buffer
	var args []string = []string{"list"}
	if len(tags) > 0 {
		args = append(args, "--tags", tags)
	}
	args = append(args, "-f={{.Export}}", "-export", pkgPath)
	cmd := exec.Command("go", args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = dir
	err = cmd.Run()
	if err == nil {
		ret = stdout.Bytes()
	} else if stderr.Len() > 0 {
		err = errors.New(stderr.String())
	}
	return
}

// ----------------------------------------------------------------------------
