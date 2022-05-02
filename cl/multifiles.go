package cl

import (
	"log"
	"path/filepath"
	"strconv"

	"github.com/goplus/c2go/clang/ast"
	ctypes "github.com/goplus/c2go/clang/types"
)

type multiFileCtl struct {
	exists   map[string]none
	base     *int // anonymous struct/union
	hasMulti bool
	inHeader bool
}

func (p *multiFileCtl) initMultiFileCtl(conf *Config) {
	if reused := conf.Reused; reused != nil {
		if reused.exists == nil {
			reused.exists = make(map[string]none)
		}
		p.exists = reused.exists
		p.base = &reused.base
		p.hasMulti = true
	} else {
		p.base = new(int)
	}
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
	return "_cgoa_" + strconv.Itoa(*p.base), suAnonymous
}

func (p *blockCtx) logFile(node *ast.Node) {
	if f := node.Loc.PresumedFile; f != "" {
		if debugCompileDecl {
			log.Println("==>", f)
		}
		if p.hasMulti {
			var fname string
			switch filepath.Ext(f) {
			case ".c":
				fname = filepath.Base(f) + ".i.go"
				p.inHeader = false
			default:
				fname = headerGoFile
				p.inHeader = true
			}
			p.pkg.SetCurFile(fname, true)
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
