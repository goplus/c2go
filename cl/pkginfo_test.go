package cl

import (
	"bytes"
	"os"
	"testing"
)

// -----------------------------------------------------------------------------

func TestPkgInfo_Basic(t *testing.T) {
	pkg := testFunc(t, "Basic", `

void f();
void test(struct foo* in) {
	f();
}
`, `func test(in *struct_foo) {
	f()
}`)
	var out bytes.Buffer
	pkg.WriteDepTo(&out)
	deps := out.String()
	if deps != `package main

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

func TestPkgInfo_Public(t *testing.T) {
	pkg := testFuncEx(t, "Basic", `

void f();
void test(struct foo* in) {
	f();
}
`, `func test(in *struct_foo) {
	F()
}`, func(c *Config) {
		c.Public = map[string]string{"f": ""}
	})
	var out bytes.Buffer
	pkg.WriteDepTo(&out)
	deps := out.String()
	if deps != `package main

type struct_foo struct {
}

func F() {
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
