package cl

import (
	"bytes"
	"go/types"
	"os"
	"testing"
)

// -----------------------------------------------------------------------------

func TestEnsureSig(t *testing.T) {
	a := types.NewParam(0, nil, "", types.Typ[types.Int])
	b := types.NewParam(0, nil, "b", types.Typ[types.Bool])
	params := types.NewTuple(a, b)
	sig := types.NewSignature(nil, params, nil, false)
	if tsig := ensureSig(sig); tsig == sig {
		t.Fatal("ensureSig 1:", sig, tsig)
	}
	params = types.NewTuple(a)
	sig = types.NewSignature(nil, params, nil, false)
	if tsig := ensureSig(sig); tsig != sig {
		t.Fatal("ensureSig 2:", sig, tsig)
	}
}

func TestPkgInfo_Basic(t *testing.T) {
	pkg := testFunc(t, "Basic", `

void f();
void test(struct foo* in, struct bar* out) {
	f();
}
`, `func test(in *struct_foo, out *struct_bar) {
	f()
}`)
	var out bytes.Buffer
	pkg.WriteDepTo(&out)
	deps := out.String()
	if deps != `package main

type struct_bar struct {
}
type struct_foo struct {
}

func f() {
	panic("notimpl")
}
` {
		t.Fatalf("WriteDepTo:\n%s\n", deps)
	}
	genfile := tmpDir + "c2go_autogen.go"
	if err := pkg.WriteDepFile(genfile); err != nil {
		t.Fatal("WriteDepFile failed:", err)
	}
	os.Remove(genfile)
}

func TestPkgInfo_BuiltinFn(t *testing.T) {
	pkg := testFuncEx(t, "Basic", `

void test(struct foo* in) {
	__builtin_inf();
}
`, `func test(in *struct_foo) {
	X__builtin_inf()
}`, func(c *Config) {
		c.Ignored = []string{"f", "g"}
		c.Public = map[string]string{"f": ""}
		c.BuiltinFuncMode = BFM_InLibC
	})
	var out bytes.Buffer
	pkg.WriteDepTo(&out)
	deps := out.String()
	if deps != `package main

type struct_foo struct {
}

func X__builtin_inf() float64 {
	panic("notimpl")
}
` {
		t.Fatalf("WriteDepTo:\n%s\n", deps)
	}
	genfile := tmpDir + "c2go_autogen.go"
	if err := pkg.WriteDepFile(genfile); err != nil {
		t.Fatal("WriteDepFile failed:", err)
	}
	os.Remove(genfile)
}

// -----------------------------------------------------------------------------
