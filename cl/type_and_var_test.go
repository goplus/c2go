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
	int a = (unsigned char)-1;
	static long b;
	if (b == -1) {
		a = 3;
	}
}
`, `func test() {
	var a int32 = int32(255)
	if _cgos_test_b == int64(-1) {
		a = int32(3)
	}
}`)
	testFunc(t, "testKeyword", `
void test() {
	int type;
	double import;
	type = (long)-1;
	import = type;
}
`, `func test() {
	var type_ int32
	var import_ float64
	type_ = int32(-1)
	import_ = float64(type_)
}`)
	testFunc(t, "testExtern", `
void test() {
	extern int a;
}
`, `func test() {
}`)
	testFunc(t, "testString", `
void test() {
	char a[] = "Hi";
	char *p = "Hi";
}
`, `func test() {
	var a [3]int8 = [3]int8{'H', 'i', '\x00'}
	var p *int8 = (*int8)(unsafe.Pointer(&[3]int8{'H', 'i', '\x00'}))
}`)
	testFunc(t, "testStringInArray", `
void test() {
	struct {
		char *a;
	} b = {"Hi"};
	struct {
		char a[6];
	} c = {"Hi"};
}
`, `func test() {
	type _cgoa_1 struct {
		a *int8
	}
	var b _cgoa_1 = _cgoa_1{(*int8)(unsafe.Pointer(&[3]int8{'H', 'i', '\x00'}))}
	type _cgoa_2 struct {
		a [6]int8
	}
	var c _cgoa_2 = _cgoa_2{[6]int8{'H', 'i', '\x00'}}
}`)
	testFunc(t, "testIntArray", `
void test() {
	struct {
		int a[6];
		int b;
	} x = {{1, 2, 3}};
}
`, `func test() {
	type _cgoa_1 struct {
		a [6]int32
		b int32
	}
	var x _cgoa_1 = _cgoa_1{[6]int32{int32(1), int32(2), int32(3)}, 0}
}`)
}

// -----------------------------------------------------------------------------

func TestBitField(t *testing.T) {
	testFunc(t, "testBasic", `
void test() {
	struct foo {
		int a;
		int b :1;
		int c :2;
		double x;
	};
}
`, `func test() {
	type struct_foo struct {
		a     int32
		Xbf_0 int32
		x     float64
	}
}`)
	testFunc(t, "testBFInit", `
void test() {
	struct foo {
		int a;
		int b :1;
		int c :2;
		double x;
	} a = {1};
}
`, `func test() {
	type struct_foo struct {
		a     int32
		Xbf_0 int32
		x     float64
	}
	var a struct_foo = struct_foo{int32(1), 0, 0}
}`)
}

// -----------------------------------------------------------------------------

func TestVoid(t *testing.T) {
	testFunc(t, "testVoidTypedef", `
void test() {
	typedef void foo;
	void* p = (void*)0;
}
`, `func test() {
	var p unsafe.Pointer = unsafe.Pointer(nil)
}`)
}

// -----------------------------------------------------------------------------

func TestConst(t *testing.T) {
	testFunc(t, "testEnum", `
void test() {
	const i = 100;
}
`, `func test() {
	const i int32 = int32(100)
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
		a int32 = int32(3)
		b int32 = 4
	)
	type foo = int32
}`)
}

// -----------------------------------------------------------------------------

func TestValist(t *testing.T) {
	testFunc(t, "testValistVar", `
void test() {
	__builtin_va_list a;
}
`, `func test() {
	var a []interface {
	}
}`)
	testFunc(t, "testValistTypedef", `
void test() {
	typedef __builtin_va_list foo;
}
`, `func test() {
	type foo = []interface {
	}
}`)
}

// -----------------------------------------------------------------------------

func TestArray(t *testing.T) {
	testFunc(t, "testExtern", `
void test() {
	extern int a[];
}
`, `func test() {
}`)
	testFunc(t, "testTypedef", `
void test() {
	typedef int foo[];
}
`, `func test() {
}`)
	testFunc(t, "testArray", `
void test() {
	struct foo {
		int a[3];
	};
}
`, `func test() {
	type struct_foo struct {
		a [3]int32
	}
}`)
	testFunc(t, "testDynArray", `
void test() {
	struct foo {
		int h;
		int a[];
	};
}
`, `func test() {
	type struct_foo struct {
		h int32
		a [0]int32
	}
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
	testFunc(t, "testKeyword", `
void test() {
	struct foo {
		int type;
	};
	struct foo a;
	a.type = 1;
}
`, `func test() {
	type struct_foo struct {
		type_ int32
	}
	var a struct_foo
	a.type_ = int32(1)
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

func TestFloat(t *testing.T) {
	testFunc(t, "testFloat", `
void f(float a, double b, ...){}
void test() {
	f(1.0, 1.0, 1.0);
}
`, `func test() {
	f(float32(1.0), 1.0, 1.0)
}`)
}

func TestPointer(t *testing.T) {
	testFunc(t, "testPointer", `
void test() {
	int *a;
	int *b;
	int c;
	a>b;
	a>=b;
	a<b;
	a<=b;
	a=b;
	a-b;
}
`, `func test() {
	var a *int32
	var b *int32
	var c int32
	uintptr(unsafe.Pointer(a)) > uintptr(unsafe.Pointer(b))
	uintptr(unsafe.Pointer(a)) >= uintptr(unsafe.Pointer(b))
	uintptr(unsafe.Pointer(a)) < uintptr(unsafe.Pointer(b))
	uintptr(unsafe.Pointer(a)) <= uintptr(unsafe.Pointer(b))
	a = b
	(uintptr(unsafe.Pointer(a)) - uintptr(unsafe.Pointer(b))) / 4
}`)
}

func TestWideString(t *testing.T) {
	testFunc(t, "testWideString", `
void test() {
	L"";
	L"\253\xab\\\u4E2D";
	L"\\\a\b\f\n\r\t\v\e\x1\x0104中A+-@\"\'123abcABC";
}
`, `func test() {
	(*int32)(unsafe.Pointer(&[1]int32{'\x00'}))
	(*int32)(unsafe.Pointer(&[5]int32{'«', '«', '\\', '中', '\x00'}))
	(*int32)(unsafe.Pointer(&[28]int32{'\\', '\a', '\b', '\f', '\n', '\r', '\t', '\v', '\x1b', '\x01', 'Ą', '中', 'A', '+', '-', '@', '"', '\'', '1', '2', '3', 'a', 'b', 'c', 'A', 'B', 'C', '\x00'}))
}`)
}

// -----------------------------------------------------------------------------
