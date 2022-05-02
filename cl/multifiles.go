package cl

import (
	"log"
	"path/filepath"

	"github.com/goplus/c2go/clang/ast"
)

type fileCtlKey struct {
	presumedFile string
	presumedLine int
}

type multiFileCtl struct {
	exists   map[fileCtlKey]none
	exist    bool
	hasMulti bool
}

func (p *multiFileCtl) initMultiFileCtl(multiCFiles bool) {
	p.hasMulti = multiCFiles
	if multiCFiles {
		p.exists = make(map[fileCtlKey]none)
	}
}

func shouldSkipFile(ctx *blockCtx, node *ast.Node) (skip bool) {
	if f := node.Loc.PresumedFile; f != "" {
		line := node.Loc.PresumedLine
		if debugCompileDecl {
			log.Println("==>", f, "line:", line)
		}
		if ctx.hasMulti {
			key := fileCtlKey{presumedFile: f, presumedLine: line}
			if _, skip = ctx.exists[key]; !skip {
				var fname string
				switch filepath.Ext(f) {
				case ".c":
					fname = filepath.Base(f) + ".i.go"
				default:
					fname = headerGoFile
				}
				ctx.pkg.SetCurFile(fname, true)
			}
			ctx.exist = skip
		}
	} else if ctx.hasMulti {
		skip = ctx.exist
	}
	return
}
