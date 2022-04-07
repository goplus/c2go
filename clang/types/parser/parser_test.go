package parser

import (
	"go/token"
	"go/types"
	"testing"

	ctypes "github.com/goplus/c2go/clang/types"
)

// -----------------------------------------------------------------------------

var (
	pkg   = types.NewPackage("", "foo")
	scope = pkg.Scope()
)

func init() {
	aliasType(scope, pkg, "char", types.Typ[types.Int8])
	aliasType(scope, pkg, "void", ctypes.Void)
	aliasType(scope, pkg, "float", types.Typ[types.Float32])
	aliasType(scope, pkg, "double", types.Typ[types.Float64])
	aliasType(scope, pkg, "__int128", ctypes.Int128)
	aliasType(scope, pkg, "ConstantString", tyConstantString)
}

func aliasType(scope *types.Scope, pkg *types.Package, name string, typ types.Type) {
	o := types.NewTypeName(token.NoPos, pkg, name, typ)
	scope.Insert(o)
}

var (
	tnameConstantString = types.NewTypeName(token.NoPos, pkg, "ConstantString", nil)
)

var (
	tyChar           = types.Typ[types.Int8]
	tyUchar          = types.Typ[types.Uint8]
	tyInt16          = types.Typ[types.Int16]
	tyUint16         = types.Typ[types.Uint16]
	tyInt32          = types.Typ[types.Int32]
	tyUint32         = types.Typ[types.Uint32]
	tyInt64          = types.Typ[types.Int64]
	tyUint64         = types.Typ[types.Uint64]
	tyInt            = ctypes.Int
	tyInt100         = types.NewArray(tyInt, 100)
	tyInt100_3       = types.NewArray(tyInt100, 3)
	tyPInt100        = types.NewPointer(tyInt100)
	tyPInt100_3      = types.NewPointer(tyInt100_3)
	tyUint           = ctypes.Uint
	tyString         = types.Typ[types.String]
	tyCharPtr        = types.NewPointer(tyChar)
	tyCharPtrPtr     = types.NewPointer(tyCharPtr)
	tyConstantString = types.NewNamed(tnameConstantString, tyString, nil)
)

var (
	paramInt        = types.NewParam(token.NoPos, pkg, "", tyInt)
	paramVoidPtr    = types.NewParam(token.NoPos, pkg, "", ctypes.UnsafePointer)
	paramCharPtrPtr = types.NewParam(token.NoPos, pkg, "", tyCharPtrPtr)
)

var (
	typesInt     = types.NewTuple(paramInt)
	typesVoidPtr = types.NewTuple(paramVoidPtr)
	typesPICC    = types.NewTuple(paramVoidPtr, paramInt, paramCharPtrPtr, paramCharPtrPtr)
)

func newFn(in, out *types.Tuple) types.Type {
	return types.NewSignature(nil, in, out, false)
}

var (
	tyFnHandle    = newFn(typesInt, nil)
	paramFnHandle = types.NewParam(token.NoPos, pkg, "", tyFnHandle)
	typesIF       = types.NewTuple(paramInt, paramFnHandle)
	typesF        = types.NewTuple(paramFnHandle)
)

// -----------------------------------------------------------------------------

type testCase struct {
	qualType string
	flags    int
	typ      types.Type
	err      string
}

var cases = []testCase{
	{qualType: "int", typ: tyInt},
	{qualType: "unsigned int", typ: tyUint},
	{qualType: "struct ConstantString", typ: tyConstantString},
	{qualType: "volatile signed int", typ: tyInt},
	{qualType: "__int128", typ: ctypes.Int128},
	{qualType: "signed", typ: tyInt},
	{qualType: "signed short", typ: tyInt16},
	{qualType: "signed long", typ: ctypes.Long},
	{qualType: "unsigned", typ: tyUint},
	{qualType: "unsigned char", typ: tyUchar},
	{qualType: "unsigned __int128", typ: ctypes.Uint128},
	{qualType: "unsigned long", typ: ctypes.Ulong},
	{qualType: "unsigned long long", typ: tyUint64},
	{qualType: "long double", typ: ctypes.LongDouble},
	{qualType: "_Complex float", typ: types.Typ[types.Complex64]},
	{qualType: "_Complex double", typ: types.Typ[types.Complex128]},
	{qualType: "_Complex long double", typ: types.Typ[types.Complex128]},
	{qualType: "int (*)(void)", typ: newFn(nil, typesInt)},
	{qualType: "void (*)(void *)", typ: newFn(typesVoidPtr, nil)},
	{qualType: "void (^ _Nonnull)(void)", typ: newFn(nil, nil)},
	{qualType: "int (*)()", typ: newFn(nil, typesInt)},
	{qualType: "int (const char *, const char *, unsigned int)", flags: FlagGetRetType, typ: tyInt},
	{qualType: "const char *restrict", typ: tyCharPtr},
	{qualType: "const char [7]", typ: types.NewArray(tyChar, 7)},
	{qualType: "const char [7]", flags: FlagIsParam, typ: tyCharPtr},
	{qualType: "int (*)[100]", typ: tyPInt100},
	{qualType: "int (*)[100][3]", typ: tyPInt100_3},
	{qualType: "char *", typ: tyCharPtr},
	{qualType: "void", typ: ctypes.Void},
	{qualType: "void *", typ: ctypes.UnsafePointer},
	{qualType: "int (*)(void *, int, char **, char **)", typ: newFn(typesPICC, typesInt)},
	{qualType: "void (*(*)(int, void (*)(int)))(int)", typ: newFn(typesIF, typesF)},
	{qualType: "void (*(int, void (*)(int)))(int)", typ: newFn(typesIF, typesF)},
	{qualType: "void (*(int, void (*)(int)))(int)", flags: FlagGetRetType, typ: tyFnHandle},
	{qualType: "int (*)(void *, int, const char *, void (**)(void *, int, void **), void **)"},
}

func TestCases(t *testing.T) {
	for _, c := range cases {
		t.Run(c.qualType, func(t *testing.T) {
			typ, _, err := ParseType(pkg, scope, c.qualType, c.flags)
			if err != nil {
				if errMsgOf(err) != c.err {
					t.Fatal("ParseType:", err, ", expected:", c.err)
				}
			} else if c.typ != nil && !types.Identical(typ, c.typ) {
				t.Fatal("ParseType:", typ, ", expected:", c.typ)
			}
		})
	}
}

func errMsgOf(err error) string {
	if e, ok := err.(*ParseTypeError); ok {
		return e.ErrMsg
	}
	return err.Error()
}

// -----------------------------------------------------------------------------
