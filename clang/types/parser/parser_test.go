package parser

import (
	"go/token"
	"go/types"
	"syscall"
	"testing"
)

// -----------------------------------------------------------------------------

type clangTypeSys struct {
}

func (p *clangTypeSys) Pkg() *types.Package {
	return pkg
}

func (p *clangTypeSys) LookupType(typ string, unsigned bool) (t types.Type, err error) {
	if typs, ok := intTypes[typ]; ok {
		idx := 0
		if unsigned {
			idx++
		}
		return typs[idx], nil
	}
	switch typ {
	case "void":
		return tyVoid, nil
	case "string":
		return tyString, nil
	case "ConstantString":
		return tyConstantString, nil
	}
	return nil, syscall.ENOENT
}

var intTypes = map[string][2]types.Type{
	"int":  {tyInt, tyUint},
	"char": {tyChar, tyUchar},
}

var (
	pkg = types.NewPackage("", "foo")
)

var (
	tnameConstantString = types.NewTypeName(token.NoPos, pkg, "ConstantString", nil)
)

var (
	tyChar           = types.Typ[types.Int8]
	tyUchar          = types.Typ[types.Uint8]
	tyInt            = types.Typ[types.Int]
	tyUint           = types.Typ[types.Uint]
	tyString         = types.Typ[types.String]
	tyVoid           = types.Typ[types.UntypedNil]
	tyCharPtr        = types.NewPointer(tyChar)
	tyConstantString = types.NewNamed(tnameConstantString, tyString, nil)
)

var (
	paramInt = types.NewParam(token.NoPos, pkg, "", tyInt)
)

// -----------------------------------------------------------------------------

type testCase struct {
	qualType string
	isParam  bool
	typ      types.Type
	err      string
}

var cases = []testCase{
	{qualType: "unsigned int", typ: tyUint},
	{qualType: "struct ConstantString", typ: tyConstantString},
	{qualType: "volatile signed int", typ: tyInt},
	{qualType: "int (*)(void)", typ: types.NewSignature(nil, nil, types.NewTuple(paramInt), false)},
	{qualType: "int (*)()", typ: types.NewSignature(nil, nil, types.NewTuple(paramInt), false)},
	{qualType: "const char *restrict", typ: tyCharPtr},
	{qualType: "const char [7]", typ: types.NewArray(tyChar, 7)},
	{qualType: "const char [7]", isParam: true, typ: tyCharPtr},
	{qualType: "char *", typ: tyCharPtr},
	{qualType: "void", typ: tyVoid},
}

func TestCases(t *testing.T) {
	ts := new(clangTypeSys)
	fset := token.NewFileSet()
	for _, c := range cases {
		t.Run(c.qualType, func(t *testing.T) {
			typ, err := ParseType(ts, fset, c.qualType, c.isParam)
			if err != nil {
				if errMsgOf(err) != c.err {
					t.Fatal("ParseType:", err, "expected:", c.err)
				}
			} else if !types.Identical(typ, c.typ) {
				t.Fatal("ParseType:", typ, "expected:", c.typ)
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
