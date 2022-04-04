# c2go - Convert C to Go

![Screen Shot1](https://user-images.githubusercontent.com/396972/160951673-30ec62ae-2981-4cdf-a1ab-bc7fcb6f7475.png)

## Supported C Syntax

### Data structures

- [x] Void: void
- [x] Boolean: bool
- [x] Integer: [signed/unsigned] char, [signed/unsigned] [short/long/long long] int
- [x] Enum: enum
- [x] Float: float, double
- [ ] Wide Character: wchar_t
- [ ] Large Integer/Float: [signed/unsigned] __int128, long double
- [ ] Complex: [float/double/long double] complex
- [x] Typedef: typedef
- [x] Pointer: *T, T[]
- [x] Array: T[N], T[]
- [ ] Array Pointer: T(*)[]
- [x] Function Pointer: T (*)(T1, T2, ...)
- [x] Struct: struct
- [x] Union: union
- [x] Bit Fields

### Operators

- [x] Arithmetic: a+b, a-b, a*b, a/b, a%b, -a, +a
- [x] Increment/Decrement: a++, a--, ++a, --a
- [x] Comparison: a<b, a<=b, a>b, a>=b, a==b, a!=b
- [x] Logical: a&&b, a||b, !a
- [x] Bitwise: a|b, a&b, a^b, ~a, a<<n, a>>n
- [x] Assignment: `=`
- [ ] Operator Assignment: a`<op>=`b
- [x] Pointer: *a, &a
- [x] Member: a.b
- [ ] Pointer Member: a->b
- [ ] Array Member: a[n]
- [ ] Comma: a,b
- [ ] Ternary Conditional: cond?a:b
- [x] Function Call: f(a1, a2, ...)
- [x] Conversion: (T)a
- [ ] Sizeof: sizeof(T), sizeof(a)

### Literals

- [x] Boolean, Integer, Float, String
- [ ] Character
- [ ] Array: `(T[]){ expr1, expr2, ... }`
- [ ] Array Pointer: `&(T[]){ expr1, expr2, ... }`
- [ ] Struct: `struct T{ expr1, expr2, ... }`

### Initialization

- [x] Basic: `T a = expr`
- [ ] Array: `T[] a = { expr1, expr2, ... }`, `T[N] a = { expr1, expr2, ... }`
- [ ] Struct: `struct T a = { expr1, expr2, ... }`, `struct T a = { .a = expr1, .b = expr2, ... }`
- [ ] Union: `union T a = { expr }, union T a = { .a = expr }`
- [ ] Array in Struct: `struct { T[N] a; ... } v = { { expr1, expr2, ... }, ... }`, `struct { T[N] a; ... } v = { { [0].a = expr1, [1].a = expr2, ... }, ... }`

### Control structures

- [x] If: `if (cond) stmt1 [else stmt2]`
- [ ] Switch: `switch (tag) { case expr1: stmt1 case expr2: stmt2 default: stmtN }`
- [x] For: `for (init; cond; post) stmt`
- [ ] While: `while (cond) stmt`
- [x] Do While: `do stmt while (cond)`
- [ ] Goto: `goto label`
- [ ] Break: `break`
- [ ] Continue: `continue`

### Functions

- [x] Parameters
- [x] Variadic Parameters
- [ ] Variadic Parameter Access
- [x] Return

