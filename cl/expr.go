package cl

import (
	goast "go/ast"
	"go/token"
	"go/types"
	"log"
	"strconv"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

const (
	flagLHS = 1 << iota
	flagIgnoreResult
)

func compileExprEx(ctx *blockCtx, expr *ast.Node, prompt string, flags int) {
	switch expr.Kind {
	case ast.BinaryOperator:
		compileBinaryExpr(ctx, expr)
	case ast.UnaryOperator:
		compileUnaryOperator(ctx, expr, flags)
	case ast.CallExpr:
		compileCallExpr(ctx, expr)
	case ast.ImplicitCastExpr:
		compileImplicitCastExpr(ctx, expr)
	case ast.DeclRefExpr:
		compileDeclRefExpr(ctx, expr, (flags&flagLHS) != 0)
	case ast.MemberExpr:
		compileMemberExpr(ctx, expr, (flags&flagLHS) != 0)
	case ast.IntegerLiteral:
		compileLiteral(ctx, token.INT, expr)
	case ast.StringLiteral:
		compileStringLiteral(ctx, expr)
	case ast.CharacterLiteral:
		compileCharacterLiteral(ctx, expr)
	case ast.ParenExpr:
		compileExpr(ctx, expr.Inner[0])
	case ast.CStyleCastExpr:
		compileTypeCast(ctx, expr, goNode(expr))
	case ast.UnaryExprOrTypeTraitExpr:
		compileUnaryExprOrTypeTraitExpr(ctx, expr)
	default:
		log.Fatalln(prompt, expr.Kind)
	}
}

func compileExpr(ctx *blockCtx, expr *ast.Node) {
	compileExprEx(ctx, expr, "compileExpr: unknown kind =", 0)
}

func compileExprLHS(ctx *blockCtx, expr *ast.Node) {
	compileExprEx(ctx, expr, "compileExpr: unknown kind =", flagLHS)
}

func compileLiteral(ctx *blockCtx, kind token.Token, expr *ast.Node) {
	ctx.cb.Val(&goast.BasicLit{Kind: kind, Value: expr.Value.(string)}, goNode(expr))
}

func compileCharacterLiteral(ctx *blockCtx, expr *ast.Node) {
	ctx.cb.Val(rune(expr.Value.(float64)), goNode(expr))
}

func compileStringLiteral(ctx *blockCtx, expr *ast.Node) {
	s, err := strconv.Unquote(expr.Value.(string))
	if err != nil {
		log.Fatalln("compileStringLiteral:", err)
	}
	stringLit(ctx.cb, s, nil)
}

// -----------------------------------------------------------------------------

func compileSizeof(ctx *blockCtx, v *ast.Node) {
	if v.Type != nil {
		t := toType(ctx, v.Type, 0)
		ctx.cb.Val(ctx.sizeof(t))
		return
	}
	log.Fatalln("compileSizeof: TODO")
}

func compileUnaryExprOrTypeTraitExpr(ctx *blockCtx, v *ast.Node) {
	switch v.Name {
	case "sizeof":
		compileSizeof(ctx, v)
	default:
		log.Fatalln("unaryExprOrTypeTraitExpr unknown:", v.Name)
	}
}

// -----------------------------------------------------------------------------

func compileImplicitCastExpr(ctx *blockCtx, v *ast.Node) {
	switch v.CastKind {
	case ast.LValueToRValue, ast.FunctionToPointerDecay, ast.NoOp:
		compileExpr(ctx, v.Inner[0])
	case ast.ArrayToPointerDecay:
		compileExpr(ctx, v.Inner[0])
		arrayToElemPtr(ctx.cb)
	case ast.IntegralCast, ast.BitCast:
		compileTypeCast(ctx, v, nil)
	default:
		log.Fatalln("compileImplicitCastExpr: unknown castKind =", v.CastKind)
	}
}

func compileTypeCast(ctx *blockCtx, v *ast.Node, src goast.Node) {
	t := toType(ctx, v.Type, 0)
	ctx.cb.Typ(t, src)
	compileExpr(ctx, v.Inner[0])
	typeCastCall(ctx, t)
}

// -----------------------------------------------------------------------------

func compileDeclRefExpr(ctx *blockCtx, v *ast.Node, lhs bool) {
	cb := ctx.cb
	name := v.ReferencedDecl.Name
	_, obj := cb.Scope().LookupParent(name, token.NoPos)
	if obj == nil {
		log.Fatalln("compileDeclRefExpr: not found -", name)
	}
	if lhs {
		cb.VarRef(obj)
	} else {
		cb.Val(obj)
	}
}

// -----------------------------------------------------------------------------

func compileCallExpr(ctx *blockCtx, v *ast.Node) {
	if n := len(v.Inner); n > 0 {
		if fn := v.Inner[0]; isBuiltinFn(fn) {
			item := fn.Inner[0]
			switch name := item.ReferencedDecl.Name; name {
			case "__builtin_va_start", "__builtin_va_end":
				return
			default:
				log.Fatalln("compileCallExpr - unknown builtin func:", name)
			}
		}
		cb := ctx.cb
		ellipsis := n > 2 && isValist(ctx, v.Inner[n-1])
		log.Println("==> compileCallExpr:",
			v.Inner[0].Inner[0].ReferencedDecl.Name, n, ellipsis)
		if ellipsis {
			n--
		}
		for i := 0; i < n; i++ {
			compileExpr(ctx, v.Inner[i])
		}
		if ellipsis {
			_, o := cb.Scope().LookupParent(valistName, token.NoPos)
			cb.Val(o)
		} else {
			n--
		}
		cb.CallWith(n, ellipsis, goNode(v))
	}
}

func isBuiltinFn(fn *ast.Node) bool {
	return fn.CastKind == ast.BuiltinFnToFnPtr
}

func isValist(ctx *blockCtx, v *ast.Node) bool {
	return v.CastKind == ast.ArrayToPointerDecay && isValistType(ctx, toType(ctx, v.Type, 0))
}

// -----------------------------------------------------------------------------

func compileMemberExpr(ctx *blockCtx, v *ast.Node, lhs bool) {
	compileExpr(ctx, v.Inner[0])
	src := goNode(v)
	if lhs {
		ctx.cb.MemberRef(v.Name, src)
	} else {
		ctx.cb.MemberVal(v.Name, src)
	}
}

// -----------------------------------------------------------------------------

func compileBinaryExpr(ctx *blockCtx, v *ast.Node) {
	if op, ok := binaryOps[v.OpCode]; ok {
		compileExpr(ctx, v.Inner[0])
		compileExpr(ctx, v.Inner[1])
		binaryOp(ctx, op, goNode(v))
		return
	}
	switch v.OpCode {
	case "=":
		compileAssignExpr(ctx, v)
	default:
		log.Fatalln("compileBinaryExpr unknown operator:", v.OpCode)
	}
}

var (
	binaryOps = map[ast.OpCode]token.Token{
		"+": token.ADD,
		"-": token.SUB,
		"*": token.MUL,
		"/": token.QUO,
		"%": token.REM,

		"&":  token.AND,
		"|":  token.OR,
		"^":  token.XOR,
		"<<": token.SHL,
		">>": token.SHR,

		"==": token.EQL,
		"<":  token.LSS,
		">":  token.GTR,
		"!=": token.NEQ,
		"<=": token.LEQ,
		">=": token.GEQ,

		"||": token.LOR,
		"&&": token.LAND,
	}
)

// -----------------------------------------------------------------------------

const (
	addrVarName = "_cgo_addr"
)

func compileAssignExpr(ctx *blockCtx, v *ast.Node) {
	cb, _ := closureStartInitAddr(ctx, v)

	addr := cb.Scope().Lookup(addrVarName)
	cb.Val(addr).ElemRef()
	compileExpr(ctx, v.Inner[1])
	cb.AssignWith(1, 1, goNode(v.Inner[1]))

	cb.Val(addr).Elem().Return(1).End().Call(0)
}

func compileSimpleIncDec(ctx *blockCtx, op token.Token, v *ast.Node) {
	cb := ctx.cb
	stk := cb.InternalStack()
	compileExprLHS(ctx, v.Inner[0])
	typ, _ := gox.DerefType(stk.Get(-1).Type)
	if t, ok := typ.(*types.Pointer); ok { // *type
		cb.UnaryOp(token.AND)
		castPtrType(cb, tyUintptrPtr, stk.Pop())
		cb.ElemRef()
		if elemSize := ctx.sizeof(t.Elem()); elemSize != 1 {
			cb.Val(elemSize).AssignOp(op + (token.ADD_ASSIGN - token.INC))
			return
		}
	}
	cb.IncDec(op)
}

func compileIncDec(ctx *blockCtx, op token.Token, v *ast.Node) {
	cb, ret := closureStartInitAddr(ctx, v)
	n := 0
	addr := cb.Scope().Lookup(addrVarName)
	if v.IsPostfix {
		cb.VarRef(ret).Val(addr).Elem().Assign(1)
	}
	elemSize := valOfAddr(cb, addr, ctx)
	cb.ElemRef()
	if elemSize == 1 {
		cb.IncDec(op)
	} else {
		cb.Val(elemSize).AssignOp(op + (token.ADD_ASSIGN - token.INC))
	}
	if !v.IsPostfix {
		cb.Val(addr).Elem()
		n = 1
	}
	cb.Return(n).End().Call(0)
}

func closureStartInitAddr(ctx *blockCtx, v *ast.Node) (*gox.CodeBuilder, *types.Var) {
	pkg := ctx.pkg
	ret := pkg.NewAutoParam("_cgo_ret")
	cb := ctx.cb.NewClosure(nil, types.NewTuple(ret), false).BodyStart(pkg)

	cb.DefineVarStart(token.NoPos, addrVarName)
	compileExprLHS(ctx, v.Inner[0])
	cb.UnaryOp(token.AND).EndInit(1)
	return cb, ret
}

// -----------------------------------------------------------------------------

func compileStarExpr(ctx *blockCtx, v *ast.Node, lhs bool) {
	compileExpr(ctx, v.Inner[0])
	src := goNode(v)
	if lhs {
		ctx.cb.ElemRef(src)
	} else {
		ctx.cb.Elem(src)
	}
}

func compileUnaryOperator(ctx *blockCtx, v *ast.Node, flags int) {
	lhs := (flags & flagLHS) != 0
	if v.OpCode == "*" {
		compileStarExpr(ctx, v, lhs)
		return
	}
	if lhs {
		log.Fatalln("compileUnaryOperator: not a lhs expression -", v.OpCode)
	}
	if op, ok := unaryOps[v.OpCode]; ok {
		compileExpr(ctx, v.Inner[0])
		if op == token.NOT {
			castToBoolExpr(ctx.cb)
		}
		ctx.cb.UnaryOp(op)
		return
	}

	var tok token.Token
	switch v.OpCode {
	case "++":
		tok = token.INC
	case "--":
		tok = token.DEC
	default:
		log.Fatalln("compileUnaryOperator: unknown operator -", v.OpCode)
	}
	if (flags & flagIgnoreResult) != 0 {
		compileSimpleIncDec(ctx, tok, v)
		return
	}
	compileIncDec(ctx, tok, v)
}

var (
	unaryOps = map[ast.OpCode]token.Token{
		"-": token.SUB,
		"&": token.AND,
		"~": token.XOR,
		"!": token.NOT,
	}
)

// -----------------------------------------------------------------------------
