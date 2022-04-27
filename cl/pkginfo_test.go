package cl

import (
	"bytes"
	"os"
	"testing"
)

// -----------------------------------------------------------------------------

func TestPkgInfo(t *testing.T) {
	pkg := testFunc(t, "Basic", `
void test(struct foo* in) {
}
`, `func test(in *struct_foo) {
}`)
	var out bytes.Buffer
	pkg.WriteDepTo(&out)
	deps := out.String()
	if deps != `package main

type struct_foo struct {
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
