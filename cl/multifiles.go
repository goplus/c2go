package cl

import (
	"encoding/json"
	"go/token"
	"go/types"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	ctypes "github.com/goplus/c2go/clang/types"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/gox"
	"github.com/goplus/mod/gopmod"
)

// -----------------------------------------------------------------------------

type multiFileCtl struct {
	PkgInfo
	incs      map[string]int  // incPath => incInSelf/incInDeps (only valid on hasMulti)
	exists    map[string]none // only valid on hasMulti
	autopub   map[string]none
	base      *int   // for anonymous struct/union or static
	baseOF    string // basename of file
	baseDir   string
	hasMulti  bool
	inHeader  bool // in header file (only valid on hasMulti)
	inDepPkg  bool // in dependent package (only valid on hasMulti)
	skipLibcH bool // skip libc header
}

func (p *multiFileCtl) initMultiFileCtl(pkg *gox.Package, conf *Config) {
	reused := conf.Reused
	if reused != nil {
		pi := reused.pkg.pi
		if pi == nil {
			pi = new(PkgInfo)
			pi.typdecls = make(map[string]*gox.TypeDecl)
			pi.extfns = make(map[string]none)
			reused.pkg.pi = pi
			reused.pkg.Package = pkg
			reused.deps.init(conf)
			initDepPkgs(pkg, &reused.deps)
		}
		p.typdecls = pi.typdecls
		p.extfns = pi.extfns
		if reused.exists == nil {
			reused.exists = make(map[string]none)
		}
		if reused.autopub == nil {
			reused.autopub = make(map[string]none)
		}
		p.exists = reused.exists
		p.autopub = reused.autopub
		p.base = &reused.base
		p.hasMulti = true
		p.incs = reused.deps.incs
		p.skipLibcH = reused.deps.skipLibcH
	} else {
		p.typdecls = make(map[string]*gox.TypeDecl)
		p.extfns = make(map[string]none)
		p.base = new(int)
	}
	if file := conf.SrcFile; file != "" {
		p.baseOF = "_" + baseOfFile(file)
		if reused != nil {
			p.base = new(int)
		}
	}
}

func baseOfFile(file string) string {
	base := filepath.Base(file)
	pos := strings.IndexFunc(base, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_')
	})
	if pos > 0 {
		base = base[:pos]
	}
	return base
}

const (
	suNormal = iota
	suAnonymous
)

func (p *blockCtx) getSuName(v *ast.Node, tag string) (string, int) {
	if name := v.Name; name != "" {
		return ctypes.MangledName(tag, name), suNormal
	}
	*p.base++
	return "_cgoa_" + strconv.Itoa(*p.base) + p.baseOF, suAnonymous
}

func (p *blockCtx) getAnonyName() string {
	*p.base++
	return "_cgoz_" + strconv.Itoa(*p.base) + p.baseOF
}

func checkAnonyUnion(typ types.Type) (t *types.Named, ok bool) {
	if t, ok = typ.(*types.Named); ok {
		name := t.Obj().Name()
		ok = strings.HasPrefix(name, "_cgoz_")
	}
	return
}

func (p *blockCtx) autoStaticName(name string) string {
	return "_cgos_" + name + p.baseOF
}

func (p *blockCtx) logFile(node *ast.Node) {
	if f := node.Loc.PresumedFile; f != "" {
		if p.hasMulti {
			var fname string
			switch filepath.Ext(f) {
			case ".c":
				fname = filepath.Base(f) + ".i.go"
				p.inHeader = false
				p.inDepPkg = false
			default:
				fname = headerGoFile
				p.inHeader = true
				p.inDepPkg = p.skipLibcH
				f, _ = filepath.Abs(f) // f is related to cwd, not p.baseDir
				for dir, kind := range p.incs {
					if strings.HasPrefix(f, dir) {
						suffix := f[len(dir):]
						if suffix == "" || suffix[0] == '/' || suffix[0] == '\\' {
							p.inDepPkg = (kind == incInDeps)
							break
						}
					}
				}
			}
			p.pkg.SetCurFile(fname, true)
			if debugCompileDecl && !p.inDepPkg {
				log.Println("==>", f)
			}
		}
	}
	return
}

func (p *blockCtx) checkExists(name string) (exist bool) {
	if p.inHeader {
		if _, exist = p.exists[name]; !exist {
			p.exists[name] = none{}
		}
	}
	return
}

func (p *blockCtx) inSrcFile() bool {
	return p.hasMulti && !p.inHeader
}

func canonical(baseDir string, uri string) string {
	if filepath.IsAbs(uri) {
		return uri
	}
	return filepath.Join(baseDir, uri)
}

// -----------------------------------------------------------------------------

type pubName struct {
	name   string
	goName string
}

type depPkg struct {
	path string
	pubs []pubName
}

const (
	incInSelf = iota
	incInDeps
)

type depPkgs struct {
	pkgs      []depPkg
	incs      map[string]int // incPath => incInSelf/incInDeps
	loaded    bool
	skipLibcH bool // skip libc header
}

func initDepPkgs(pkg *gox.Package, deps *depPkgs) {
	scope := pkg.Types.Scope()
	for _, dep := range deps.pkgs {
		depPkg := pkg.Import(dep.path)
		for _, pub := range dep.pubs {
			obj := depPkg.Ref(pub.goName)
			if old := scope.Insert(gox.NewSubst(token.NoPos, pkg.Types, pub.name, obj)); old != nil {
				log.Panicf("conflicted name `%v` in %v, previous definition is %v\n", pub.name, dep.path, old)
			}
		}
	}
}

func (p *depPkgs) init(conf *Config) {
	if p.loaded {
		return
	}
	p.loaded = true
	deps := conf.Deps
	if len(deps) == 0 {
		return
	}
	base, err := filepath.Abs(conf.Dir)
	if err != nil {
		log.Panicln("filepath.Abs failed:", err)
	}
	p.incs = make(map[string]int)
	for _, dir := range conf.Include {
		dir = canonical(base, dir)
		p.incs[dir] = incInSelf
	}
	procDepPkg := conf.ProcDepPkg
	gomod, _ := gopmod.Load(base, 0)
	for _, dep := range deps {
		if dep == "C" {
			p.skipLibcH = true
			continue
		}
		depPkgDir := findPkgDir(gomod, dep)
		if procDepPkg != nil {
			procDepPkg(depPkgDir)
		}
		depPkgIncs, err := findIncludeDirs(depPkgDir)
		if err != nil {
			log.Panicln("findIncludeDirs:", err)
		}
		for _, dir := range depPkgIncs {
			dir = canonical(depPkgDir, dir)
			p.incs[dir] = incInDeps
		}
		pubfile := filepath.Join(depPkgDir, "c2go.a.pub")
		p.loadPubFile(dep, pubfile)
	}
}

func (p *depPkgs) loadPubFile(path string, pubfile string) {
	if debugLoadDeps {
		log.Println("==> loadPubFile:", path, pubfile)
	}
	pubs := loadPubFile(pubfile)
	p.pkgs = append(p.pkgs, depPkg{path: path, pubs: pubs})
}

func findPkgDir(gomod *gopmod.Module, pkgPath string) (pkgDir string) {
	if gomod == nil {
		log.Panicln("findPkgDir TODO: no go.mod found")
	}
	pkg, err := gomod.Lookup(pkgPath)
	if err != nil {
		log.Panicln("gomod.Lookup:", err)
	}
	pkgDir, err = filepath.Abs(pkg.Dir)
	if err != nil {
		log.Panicln("filepath.Abs:", err)
	}
	return
}

func findIncludeDirs(pkgDir string) (incs []string, err error) {
	var conf struct {
		Include []string `json:"include"`
	}
	file := filepath.Join(pkgDir, "c2go.cfg")
	b, err := os.ReadFile(file)
	if err != nil {
		return
	}
	err = json.Unmarshal(b, &conf)
	if err != nil {
		return
	}
	return conf.Include, nil
}

// -----------------------------------------------------------------------------
