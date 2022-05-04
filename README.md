c2go - Convert C to Go
========

[![Build Status](https://github.com/goplus/c2go/actions/workflows/go.yml/badge.svg)](https://github.com/goplus/c2go/actions/workflows/go.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/goplus/c2go)](https://goreportcard.com/report/github.com/goplus/c2go)
[![GitHub release](https://img.shields.io/github/v/tag/goplus/c2go.svg?label=release)](https://github.com/goplus/c2go/releases)
[![Coverage Status](https://codecov.io/gh/goplus/c2go/branch/main/graph/badge.svg)](https://codecov.io/gh/goplus/c2go)
[![GoDoc](https://pkg.go.dev/badge/github.com/goplus/c2go.svg)](https://pkg.go.dev/mod/github.com/goplus/c2go)

![Screen Shot1](https://user-images.githubusercontent.com/396972/160951673-30ec62ae-2981-4cdf-a1ab-bc7fcb6f7475.png)


## How to run examples?

Run an example:

- Build c2go tools: `go install -v ./...`
- Go `testdata/xxx` directory, and run `c2go .`

Run/Test multiple examples:

- Run examples: `c2go ./...`
- Test examples: `c2go -test ./...`


## What's our plan?

- First, support most of the syntax of C. Stage: `Done Almost`, see [supported C syntax](#supported-c-syntax).
- Second, compile `sqlite3` to fix c2go bugs and get list of its dependent C standard libary fuctions. Stage: `Done`, see [github.com/goplus/sqlite](https://github.com/goplus/sqlite) and [its dependent fuctions](https://github.com/goplus/sqlite/blob/main/c2go_autogen.go).
- Third, support most of C standard libaries (especially that used by `sqlite3`), and can import them by Go+. Stage: `Doing`, see [detailed progress here](https://github.com/goplus/libc#whats-our-plan).
- Last, support all custom libraries, especially those well-known open source libraries. Stage: `Planning`.


## Tested Platform

- [x] MacOS: 1.17.x
- [x] Linux: ubuntu-20.04 (temporarily skip `testdata/qsort`)
- [ ] Windows


## Supported C syntax

### Data structures

- [x] Void: `void`
- [x] Boolean: `_Bool`, `bool`
- [x] Integer: [`signed`/`unsigned`] [`short`/`long`/`long long`] `int`
- [x] Enum: `enum`
- [x] Float: `float`, `double`, `long double`
- [x] Character: [`signed`/`unsigned`] `char`
- [ ] Wide Character: `wchar_t`
- [ ] Large Integer: [`signed`/`unsigned`] `__int128`
- [x] Complex: `_Complex` `float`/`double`/`long double`
- [x] Typedef: `typedef`
- [x] Pointer: *T, T[]
- [x] Array: T[N], T[]
- [x] Array Pointer: T(*)[N]
- [x] Function Pointer: T (*)(T1, T2, ...)
- [x] Struct: `struct`
- [x] Union: `union`
- [x] BitField: `intType :N`

### Operators

- [x] Arithmetic: a+b, a-b, a*b, a/b, a%b, -a, +a
- [x] Increment/Decrement: a++, a--, ++a, --a
- [x] Comparison: a<b, a<=b, a>b, a>=b, a==b, a!=b
- [x] Logical: a&&b, a||b, !a
- [x] Bitwise: a|b, a&b, a^b, ~a, a<<n, a>>n
- [x] Pointer Arithmetic: p+n, p-n, p-q, p++, p--
- [x] Assignment: `=`
- [x] Operator Assignment: a`<op>=`b
- [x] BitField Assignment: `=`
- [ ] BitField Operator Assignment: a`<op>=`b
- [x] Struct/Union/BitField Member: a.b
- [x] Array Member: a[n]
- [x] Pointer Member: &a, *p, p[n], p->b
- [x] Comma: `a,b`
- [x] Ternary Conditional: cond?a:b
- [x] Function Call: f(a1, a2, ...)
- [x] Conversion: (T)a
- [x] Sizeof: sizeof(T), sizeof(a)
- [x] Offsetof: __builtin_offsetof(T, member)

### Literals

- [x] Boolean, Integer
- [x] Float, Complex Imaginary
- [x] Character, String
- [ ] Array: `(T[]){ expr1, expr2, ... }`
- [ ] Array Pointer: `&(T[]){ expr1, expr2, ... }`
- [ ] Struct: `struct T{ expr1, expr2, ... }`

### Initialization

- [x] Basic: `T a = expr`
- [x] Array: `T a[] = { expr1, expr2, ... }`, `T a[N] = { expr1, expr2, ... }`
- [x] Struct: `struct T a = { expr1, expr2, ... }`, `struct T a = { .a = expr1, .b = expr2, ... }`
- [x] Union: `union T a = { expr }, union T a = { .a = expr }`
- [x] Array in Struct: `struct { T a[N]; ... } v = { { expr1, expr2, ... }, ... }`, `struct { T a[N]; ... } v = { { [0].a = expr1, [1].a = expr2, ... }, ... }`

### Control structures

- [x] If: `if (cond) stmt1 [else stmt2]`
- [x] Switch: `switch (tag) { case expr1: stmt1 case expr2: stmt2 default: stmtN }`
- [x] For: `for (init; cond; post) stmt`
- [x] While: `while (cond) stmt`
- [x] Do While: `do stmt while (cond)`
- [x] Break/Continue: `break`, `continue`
- [x] Goto: `goto label`

### Functions

- [x] Parameters
- [x] Variadic Parameters
- [x] Variadic Parameter Access
- [x] Return
