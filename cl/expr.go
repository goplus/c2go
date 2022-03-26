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
		compileUnaryOperator(ctx, expr)
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
	compileExpr(ctx, v.Inner[0])
	compileExpr(ctx, v.Inner[1])
	op, ok := binaryOps[v.OpCode]
	if !ok {
		log.Fatalln("compileBinaryExpr: unknown operator =", v.OpCode)
	}
	ctx.cb.BinaryOp(op, goNode(v))
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
		cb.VarRef(addr).Star().VarRef(ret).Assign(1).
			VarRef(addr).Star().IncDec(op)
	} else {
		cb.VarRef(addr).Star().IncDec(op)
		cb.VarRef(addr).Star()
		n = 1
	}
	cb.Return(n).End()
}

func compileUnaryOperator(ctx *blockCtx, v *ast.Node) {
	compileExpr(ctx, v.Inner[0])
	op, ok := unaryOps[v.OpCode]
	if !ok {
		switch v.OpCode {
		case "++":
			compileIncDec(ctx, token.INC, v)
		case "--":
			compileIncDec(ctx, token.DEC, v)
		default:
			log.Fatalln("compileUnaryOperator: unknown operator =", v.OpCode)
		}
	}
	ctx.cb.UnaryOp(op)
}

var (
	unaryOps = map[ast.OpCode]token.Token{
		"-": token.SUB,
		"*": token.MUL,
		"&": token.AND,
		"~": token.XOR,
		"!": token.NOT,
	}
)

// -----------------------------------------------------------------------------
