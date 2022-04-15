package cl

import (
	"testing"

	"github.com/goplus/c2go/clang/ast"
)

// -----------------------------------------------------------------------------

type caseSimpleSwitch struct {
	name     string
	code     string
	simpleSw bool
}

func TestIsSimpleSwitch(t *testing.T) {
	cases := []caseSimpleSwitch{
		{name: "Basic", simpleSw: true, code: `
void foo(int a) {
	switch (a) {
	case 1:
	case 2:
		;
	}
}
`},
		{name: "GotoInner", simpleSw: true, code: `
void foo(int a, int b) {
	switch (a) {
	default:
		a = b;
		retry: if (b) {
			goto retry;
		} else {
			a++;
		}
	case 2:
		;
	}
}
`},
		{name: "GotoAcross", simpleSw: false, code: `
void foo(int a, int b) {
	switch (a) {
	case 1:
		if (b) {
			goto next;
		}
	default:
		next: ;
	}
}
`},
		{name: "CaseAcross", simpleSw: false, code: `
void foo(int a, int b) {
	switch (a) {
	case 1:
		if (b) {
		case 2:
			a++;
		}
	default:
		;
	}
}
`},
		{name: "FirstStmtNotCase", simpleSw: false, code: `
#include <stdio.h>

int main() {
	int c = 9;
	switch (c&3) while((c-=4)>=0) {
		printf("=> %d\n", c);
		case 3: printf("3\n");
		case 2: printf("2\n");
			break;
		default: printf("default\n");
		case 0: printf("0\n");
	}
	return 0;
}
`},
		{name: "SqliteN1", simpleSw: true, code: `
void foo(int i, int r) {
	switch( i ){
		case 4: {
		  int x;
		  ((void)x);
		  break;
		}
		case 5: {
		  int y = (int)r;
		  ((void)y);
		  break;
		}
	}
}
`},
	}
	sel := ""
	for _, c := range cases {
		if sel != "" && c.name != sel {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			f, src := parse(c.code)
			ctx := &blockCtx{cursrc: src}
			check := checkSimpleSwitch(ctx, f)
			if check != c.simpleSw {
				t.Fatal("TestSimpleSwitch:", check, ", expect:", c.simpleSw, "code:", c.code)
			}
		})
	}
}

func checkSimpleSwitch(ctx *blockCtx, node *ast.Node) bool {
	if node.Kind == ast.SwitchStmt {
		return isSimpleSwitch(ctx, node)
	}
	for _, item := range node.Inner {
		if !checkSimpleSwitch(ctx, item) {
			return false
		}
	}
	return true
}

// -----------------------------------------------------------------------------
