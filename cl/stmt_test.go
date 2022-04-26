package cl

import (
	"testing"

	"github.com/goplus/c2go/clang/ast"
)

// -----------------------------------------------------------------------------

type caseSimpleStmt struct {
	name   string
	code   string
	simple bool
}

func TestIsSimpleSwitch(t *testing.T) {
	cases := []caseSimpleStmt{
		{name: "Basic", simple: true, code: `
void foo(int a) {
	switch (a) {
	case 1:
	case 2:
		;
	}
}
`},
		{name: "GotoInner", simple: true, code: `
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
		{name: "GotoAcross", simple: false, code: `
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
		{name: "CaseAcross", simple: false, code: `
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
		{name: "FirstStmtNotCase", simple: false, code: `
#include <stdio.h>

int foo() {
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
		{name: "SqliteN1", simple: true, code: `
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
		{name: "SqliteN2", simple: true, code: `
void foo(int op, int iA) {
  switch( op ){
	case 106: if( iA ) goto fp_math; break;
	case 107: if( iA ) goto fp_math; break;
	case 108: if( iA ) goto fp_math; break;
	case 109: {
	  if( iA==0 ) goto arithmetic_result_is_null;
	  break;
	}
	default: {
	  if( iA==0 ) goto arithmetic_result_is_null;
	  if( iA==-1 ) iA = 1;
	  break;
	}
  }
fp_math:
arithmetic_result_is_null:
  ;
}
`},
	}
	sel := ""
	for _, c := range cases {
		if sel != "" && c.name != sel {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			f, src := parse(c.code, nil)
			ctx := &blockCtx{src: src}
			check := checkSimpleStmt(ctx, f, ast.SwitchStmt)
			if check != c.simple {
				t.Fatal("TestSimpleSwitch:", check, ", expect:", c.simple, "code:", c.code)
			}
		})
	}
}

func checkSimpleStmt(ctx *blockCtx, node *ast.Node, kind ast.Kind) bool {
	fn := findNode(node, ast.FunctionDecl, "foo")
	for _, item := range fn.Inner {
		if item.Kind == ast.CompoundStmt {
			ctx.markComplicated(string(kind), item)
			return !hasComplicated(node, kind)
		}
	}
	panic("unexpected")
}

func hasComplicated(node *ast.Node, kind ast.Kind) bool {
	if node.Kind == kind {
		return node.Complicated
	}
	for _, item := range node.Inner {
		if hasComplicated(item, kind) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------

func TestIsSimpleDoStmt(t *testing.T) {
	cases := []caseSimpleStmt{
		{name: "Basic", simple: true, code: `
#include <stdio.h>

void foo(int a) {
    do {
		printf("one shot\n");
		break;
	} while (1);
}
`},
		{name: "GotoLoop", simple: false, code: `
#include <stdio.h>

void foo(int a) {
    goto one;
    do {
one:
        printf("one shot\n");
        break;
    } while (1);
}
`},
	}
	sel := "GotoLoop"
	for _, c := range cases {
		if sel != "" && c.name != sel {
			continue
		}
		t.Run(c.name, func(t *testing.T) {
			f, src := parse(c.code, nil)
			ctx := &blockCtx{src: src}
			check := checkSimpleStmt(ctx, f, ast.DoStmt)
			if check != c.simple {
				t.Fatal("TestIsSimpleDoStmt:", check, ", expect:", c.simple, "code:", c.code)
			}
		})
	}
}

// -----------------------------------------------------------------------------
