package cl

import (
	"go/token"
	"go/types"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/goplus/c2go/clang/pathutil"
	ctypes "github.com/goplus/c2go/clang/types"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

const (
	depsFile = "deps"
)

type undefStruct struct {
	name string
	typeDecl
}

func (p Package) InitDependencies() {
	if p.pi == nil {
		panic("Please set conf.NeedPkgInfo = true")
	}
	if _, ok := p.File(depsFile); ok {
		return
	}

	var uds []undefStruct
	for name, tdecl := range p.pi.typdecls {
		if tdecl.State() == gox.TyStateUninited {
			uds = append(uds, undefStruct{name, tdecl})
		}
	}
	sort.Slice(uds, func(i, j int) bool {
		return uds[i].name < uds[j].name
	})
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
		us.defineHere()
		us.InitType(pkg, empty)
	}
	vPanic := types.Universe.Lookup("panic")
	for _, uf := range extfns {
		var sig *types.Signature
		switch t := scope.Lookup(uf).Type().(type) {
		case *types.Signature:
			sig = t
		case *gox.SubstType:
			real := t.Real
			uf, sig = real.Name(), real.Type().(*types.Signature)
		}
		f, _ := pkg.NewFuncWith(token.NoPos, uf, ensureSig(sig), nil)
		f.BodyStart(pkg).
			Val(vPanic).Val("notimpl").Call(1).EndStmt().
			End()
	}
}

const (
	sigParamWithoutName = 1 << iota
	sigParamWithName
	sigParamInvalid = sigParamWithoutName | sigParamWithName
)

func ensureSig(sig *types.Signature) *types.Signature {
	params := sig.Params()
	kind := 0
	for i, n := 0, params.Len(); i < n; i++ {
		param := params.At(i)
		if param.Name() != "" {
			kind |= sigParamWithName
		} else {
			kind |= sigParamWithoutName
		}
		if kind == sigParamInvalid {
			newParams := make([]*types.Var, n)
			for i = 0; i < n; i++ {
				param = params.At(i)
				if param.Name() != "" {
					param = types.NewParam(param.Pos(), param.Pkg(), "", param.Type())
				}
				newParams[i] = param
			}
			return types.NewSignature(nil, types.NewTuple(newParams...), sig.Results(), sig.Variadic())
		}
	}
	return sig
}

func (p Package) WriteDepTo(dst io.Writer) error {
	p.InitDependencies()
	return p.Package.WriteTo(dst, depsFile)
}

func (p Package) WriteDepFile(file string) error {
	p.InitDependencies()
	return p.Package.WriteFile(file, depsFile)
}

// -----------------------------------------------------------------------------

func loadPubFile(pubfile string) (pubs []pubName) {
	b, err := os.ReadFile(pubfile)
	if err != nil {
		log.Panicln("loadPubFile failed:", err)
	}

	text := string(b)
	lines := strings.Split(text, "\n")
	pubs = make([]pubName, 0, len(lines))
	for i, line := range lines {
		flds := strings.Fields(line)
		goName := ""
		switch len(flds) {
		case 1:
			goName = gox.CPubName(flds[0])
		case 2:
			goName = flds[1]
		case 0:
			continue
		default:
			log.Panicf("%s:%d: too many fields - %s\n", pubfile, i+1, line)
		}
		pubs = append(pubs, pubName{name: flds[0], goName: goName})
	}
	return
}

// baseDir should be absolute path
func (p *blockCtx) initPublicFrom(baseDir string, conf *Config, node *ast.Node) {
	pubFrom := conf.PublicFrom
	if len(pubFrom) == 0 {
		return
	}
	for i, from := range pubFrom {
		pubFrom[i] = pathutil.Canonical(baseDir, from)
	}
	isPub := false
	for _, decl := range node.Inner {
		if f := decl.Loc.PresumedFile; f != "" {
			isPub = isPublicFrom(filepath.Clean(f), pubFrom)
		}
		if isPub {
			switch decl.Kind {
			case ast.FunctionDecl, ast.TypedefDecl, ast.VarDecl:
				p.autopub[decl.Name] = none{}
			case ast.RecordDecl:
				if decl.Name != "" {
					suName := ctypes.MangledName(decl.TagUsed, decl.Name)
					p.autopub[suName] = none{}
				}
			}
		}
	}
}

// f, pubFrom are absolute paths
func isPublicFrom(f string, pubFrom []string) bool {
	for _, from := range pubFrom {
		if f == from {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
