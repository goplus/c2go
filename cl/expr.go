package cl

import (
	"go/token"
	"log"

	"github.com/goplus/c2go/clang/ast"
)

// -----------------------------------------------------------------------------

func compileExprEx(ctx *blockCtx, expr *ast.Node, prompt string) {
	switch expr.Kind {
	case ast.BinaryOperator:
		compileBinaryExpr(ctx, expr)
	case ast.UnaryOperator:
		compileUnaryOperator(ctx, expr)
	case ast.CallExpr:
		compileCallExpr(ctx, expr)
	case ast.MemberExpr:
		compileMemberExpr(ctx, expr)
	default:
		log.Fatalln(prompt, expr.Kind)
	}
}

func compileExpr(ctx *blockCtx, expr *ast.Node) {
	compileExprEx(ctx, expr, "compileExpr: unknown kind =")
}

// -----------------------------------------------------------------------------

func compileCallExpr(ctx *blockCtx, v *ast.Node) {
}

// -----------------------------------------------------------------------------

func compileMemberExpr(ctx *blockCtx, v *ast.Node) {
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

func compileUnaryOperator(ctx *blockCtx, v *ast.Node) {
	compileExpr(ctx, v.Inner[0])
	op, ok := unaryOps[v.OpCode]
	if !ok {
		log.Fatalln("compileUnaryOperator: unknown operator =", v.OpCode)
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
