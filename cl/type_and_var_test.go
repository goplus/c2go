package cl

import (
	"go/token"
	"go/types"
	"testing"

	ctypes "github.com/goplus/c2go/clang/types"
)

// -----------------------------------------------------------------------------

func identicalBfType(t1, t2 *bfType) bool {
	if *t1.BitField != *t2.BitField {
		return false
	}
	if t1.first != t2.first {
		return false
	}
	return types.Identical(t1.Type, t2.Type)
}

func identicalStruct(t1, t2 *types.Struct) bool {
	n1 := t1.NumFields()
	n2 := t2.NumFields()
	if n1 != n2 {
		return false
	}
	for i := 0; i < n1; i++ {
		f1 := t1.Field(i)
		f2 := t2.Field(i)
		if f1.Name() != f2.Name() {
			return false
		}
		switch t1 := f1.Type().(type) {
		case *bfType:
			t2, ok := f2.Type().(*bfType)
			if !ok || !identicalBfType(t1, t2) {
				return false
			}
		default:
			_, ok := f2.Type().(*bfType)
			if ok || !types.Identical(t1, f2.Type()) {
				return false
			}
		}
	}
	return true
}

// -----------------------------------------------------------------------------

type structFld struct {
	name string
	typ  types.Type
}

type caseStruct struct {
	name string
	code string
	flds []structFld
}

func newStruc(pair ...interface{}) []structFld {
	n := len(pair)
	flds := make([]structFld, 0, n/2)
	for i := 0; i < n; i += 2 {
		name := pair[i].(string)
		typ := pair[i+1].(types.Type)
		flds = append(flds, structFld{name, typ})
	}
	return flds
}

func newStrucT(pkg *types.Package, flds []structFld) *types.Struct {
	items := make([]*types.Var, len(flds))
	for i, fld := range flds {
		items[i] = types.NewField(token.NoPos, pkg, fld.name, fld.typ, false)
	}
	return types.NewStruct(items, nil)
}

var (
	tyInt = ctypes.Int
)

func TestStruct(t *testing.T) {
	cases := []caseStruct{
		{name: "Basic", flds: newStruc("a", tyInt), code: `
struct foo {
	int a;
};
`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := newTestEnv(c.code)
			compileDeclStmt(e.ctx, e.doc, true)
			o := e.pkg.Scope().Lookup("struct_foo")
			if o == nil {
				t.Fatal("object not found")
			}
			if t1, ok := o.Type().Underlying().(*types.Struct); ok {
				t2 := newStrucT(e.pkg, c.flds)
				if identicalStruct(t1, t2) {
					return
				}
				t.Fatal("identicalStruct failed:", t1, t2)
			}
			t.Fatal("unexpected type:", o.Type())
		})
	}
}

// -----------------------------------------------------------------------------
