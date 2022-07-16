package cl

import (
	"testing"

	"github.com/goplus/gox"
)

func TestInitDepPkgs(t *testing.T) {
	testPanic(t, "conflicted name `printf` in github.com/goplus/c2go/cl/internal/libc, previous definition is var printf substType{real: func github.com/goplus/c2go/cl/internal/libc.Printf(fmt *int8, args ...interface{})}\n", func() {
		pkg := gox.NewPackage("", "foo", nil)
		dep := depPkg{
			path: "github.com/goplus/c2go/cl/internal/libc",
			pubs: []pubName{{name: "printf", goName: "Printf"}},
		}
		deps := &depPkgs{
			pkgs: []depPkg{dep, dep},
		}
		initDepPkgs(pkg, deps)
	})
}

func TestBaseOfFile(t *testing.T) {
	if ret := baseOfFile("src/errno/strerror.c.i"); ret != "strerror" {
		t.Fatal("baseOfFile:", ret)
	}
}
