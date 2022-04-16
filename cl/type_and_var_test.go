package cl

import (
	"go/token"
	"go/types"
	"testing"

	ctypes "github.com/goplus/c2go/clang/types"
	"github.com/goplus/gox"
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

func newBftype(typ types.Type, fldName, name string, off, bits int, first bool) *bfType {
	return &bfType{
		Type: typ,
		BitField: &gox.BitField{
			Name:    name,
			FldName: fldName,
			Off:     off,
			Bits:    bits,
		},
		first: first,
	}
}

var (
	tyInt = ctypes.Int
)

// -----------------------------------------------------------------------------

type caseStruct struct {
	name string
	code string
	flds []structFld
}

func TestVStruct(t *testing.T) {
	cases := []caseStruct{
		{name: "Basic", flds: newStruc("a", tyInt), code: `
struct foo {
	int a;
};
`},
		{name: "BitF1", flds: newStruc(
			"u", tyInt,
			"a", newBftype(tyInt, "Xbf_0", "a", 0, 1, true),
			"b", newBftype(tyInt, "Xbf_0", "b", 1, 2, false),
			"x", tyInt,
			"c", newBftype(tyInt, "Xbf_1", "c", 0, 3, true),
			"y", tyInt,
		), code: `
struct foo {
	int u;
	int a :1;
	int b :2;
	int x;
	int c :3;
	int y;
};
`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			e := newTestEnv(c.code)
			compileDeclStmt(e.ctx, e.doc, true)
			pkg := e.pkg.Types
			o := pkg.Scope().Lookup("struct_foo")
			if o == nil {
				t.Fatal("object not found")
			}
			named := o.Type().(*types.Named)
			t1 := e.ctx.getVStruct(named)
			t2 := newStrucT(pkg, c.flds)
			if identicalStruct(t1, t2) {
				return
			}
			t.Fatal("identicalStruct failed:", t1, t2)
		})
	}
}

// -----------------------------------------------------------------------------

func TestVarAndInit(t *testing.T) {
	testFunc(t, "testBasic", `
void test() {
	int a;
}
`, `func test() {
	var a int32
}`)
}

// -----------------------------------------------------------------------------

func TestEnum(t *testing.T) {
	testFunc(t, "testEnum", `
void test() {
	enum foo {
		a,
		b
	};
}
`, `func test() {
	const (
		a int32 = 0
		b int32 = 1
	)
}`)
	testFunc(t, "testTypedefEnum", `
void test() {
	typedef enum foo {
		a = 3,
		b
	} foo;
}
`, `func test() {
	const (
		a int32 = 3
		b int32 = 4
	)
	type foo = int32
}`)
}

// -----------------------------------------------------------------------------

func TestStructUnion(t *testing.T) {
	testFunc(t, "testStruct", `
void test() {
	struct foo {
		int a;
	};
}
`, `func test() {
	type struct_foo struct {
		a int32
	}
}`)
	testFunc(t, "testAnonymous", `
void test() {
	struct foo {
		int a;
		struct {
			double v;
		};
	};
}
`, `func test() {
	type _cgoa_1 struct {
		v float64
	}
	type struct_foo struct {
		a int32
		_cgoa_1
	}
}`)
	testFunc(t, "testAnonymousVar", `
void test() {
	struct foo {
		int a;
		struct {
			double v;
		} b;
	};
}
`, `func test() {
	type _cgoa_1 struct {
		v float64
	}
	type struct_foo struct {
		a int32
		b _cgoa_1
	}
}`)
	testFunc(t, "testNestStruct", `
void test() {
	struct foo {
		int a;
		struct bar {
			double v;
		} b;
	};
}
`, `func test() {
	type struct_bar struct {
		v float64
	}
	type struct_foo struct {
		a int32
		b struct_bar
	}
}`)
	testFunc(t, "testUnion", `
void test() {
	union foo {
		int a;
		double b;
	};
}
`, `func test() {
	type union_foo struct {
		b float64
	}
}`)
	testFunc(t, "testUnionNest", `
void test() {
	union foo {
		int a;
		double b;
		struct bar {
			int x;
			double y;
		} c;
	};
}
`, `func test() {
	type struct_bar struct {
		x int32
		y float64
	}
	type union_foo struct {
		c struct_bar
	}
}`)
	testFunc(t, "testUnionAnonymous", `
void test() {
	union foo {
		int a;
		double b;
		struct {
			int x;
			double y;
		};
	};
}
`, `func test() {
	type _cgoa_1 struct {
		x int32
		y float64
	}
	type union_foo struct {
		 _cgoa_1
	}
}`)
	testFunc(t, "testUnionAnonymousVar", `
void test() {
	union foo {
		int a;
		double b;
		struct {
			int x;
			double y;
		} c;
	};
}
`, `func test() {
	type _cgoa_1 struct {
		x int32
		y float64
	}
	type union_foo struct {
		c _cgoa_1
	}
}`)
	testFunc(t, "testTypedef", `
void test() {
	typedef struct foo {
		int a;
	} foo;
}
`, `func test() {
	type struct_foo struct {
		a int32
	}
	type foo = struct_foo
}`)
}

// -----------------------------------------------------------------------------
