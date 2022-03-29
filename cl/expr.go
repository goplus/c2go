package cl

import (
	"go/token"
	"go/types"
	"log"

	"github.com/goplus/c2go/clang/ast"
)

// -----------------------------------------------------------------------------

func compileExprEx(ctx *blockCtx, expr *ast.Node, prompt string, lhs bool) {
	switch expr.Kind {
	case ast.BinaryOperator:
		compileBinaryExpr(ctx, expr)
	case ast.UnaryOperator:
		compileUnaryOperator(ctx, expr, lhs)
	case ast.CallExpr:
		compileCallExpr(ctx, expr)
	case ast.ImplicitCastExpr:
		compileImplicitCastExpr(ctx, expr)
	case ast.DeclRefExpr:
		compileDeclRefExpr(ctx, expr, lhs)
	case ast.MemberExpr:
		compileMemberExpr(ctx, expr)
	case ast.ParenExpr:
		compileExpr(ctx, expr.Inner[0])
	default:
		log.Fatalln(prompt, expr.Kind)
	}
}

func compileExpr(ctx *blockCtx, expr *ast.Node) {
	compileExprEx(ctx, expr, "compileExpr: unknown kind =", false)
}

// -----------------------------------------------------------------------------

func compileImplicitCastExpr(ctx *blockCtx, v *ast.Node) {
	switch v.CastKind {
	case ast.LValueToRValue:
		compileExpr(ctx, v.Inner[0])
	// case ast.FunctionToPointerDecay:
	default:
		log.Fatalln("compileImplicitCastExpr: unknown castKind =", v.CastKind)
	}
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
		for _, arg := range v.Inner {
			compileExpr(ctx, arg)
		}
		ctx.cb.CallWith(n-1, false, goNode(v))
	}
}

// -----------------------------------------------------------------------------

func compileMemberExpr(ctx *blockCtx, v *ast.Node) {
	compileExpr(ctx, v.Inner[0])
	ctx.cb.MemberRef(v.Name, goNode(v))
}

// -----------------------------------------------------------------------------

func compileBinaryExpr(ctx *blockCtx, v *ast.Node) {
	if op, ok := binaryOps[v.OpCode]; ok {
		compileExpr(ctx, v.Inner[0])
		compileExpr(ctx, v.Inner[1])
		ctx.cb.BinaryOp(op, goNode(v))
		return
	}
	log.Fatalln("compileBinaryExpr: unknown operator =", v.OpCode)
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

func compileIncDec(ctx *blockCtx, op token.Token, v *ast.Node) {
	const (
		addrVarName = "_cgo_addr"
	)
	pkg := ctx.pkg
	ret := pkg.NewAutoParam("_cgo_ret")
	cb := ctx.cb.NewClosure(nil, types.NewTuple(ret), false).BodyStart(pkg)

	cb.DefineVarStart(token.NoPos, addrVarName)
	compileExpr(ctx, v.Inner[0])
	cb.UnaryOp(token.AND).EndInit(1)

	n := 0
	addr := cb.Scope().Lookup(addrVarName)
	if v.IsPostfix {
		cb.VarRef(ret).Val(addr).Elem().Assign(1).
			Val(addr).ElemRef().IncDec(op)
	} else {
		cb.Val(addr).ElemRef().IncDec(op)
		cb.Val(addr).Elem()
		n = 1
	}
	cb.Return(n).End()
}

func compileStarExpr(ctx *blockCtx, v *ast.Node, lhs bool) {
	compileExpr(ctx, v.Inner[0])
	src := goNode(v)
	if lhs {
		ctx.cb.ElemRef(src)
	} else {
		ctx.cb.Elem(src)
	}
}

func compileUnaryOperator(ctx *blockCtx, v *ast.Node, lhs bool) {
	if v.OpCode == "*" {
		compileStarExpr(ctx, v, lhs)
		return
	}
	if lhs {
		log.Fatalln("compileUnaryOperator: not a lhs expression -", v.OpCode)
	}
	if op, ok := unaryOps[v.OpCode]; ok {
		compileExpr(ctx, v.Inner[0])
		ctx.cb.UnaryOp(op)
		return
	}
	switch v.OpCode {
	case "++":
		compileIncDec(ctx, token.INC, v)
	case "--":
		compileIncDec(ctx, token.DEC, v)
	default:
		log.Fatalln("compileUnaryOperator: unknown operator -", v.OpCode)
	}
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
