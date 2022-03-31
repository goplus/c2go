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
		compileMemberExpr(ctx, expr, lhs)
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
	compileExprEx(ctx, expr, "compileExpr: unknown kind =", false)
}

func compileExprLHS(ctx *blockCtx, expr *ast.Node) {
	compileExprEx(ctx, expr, "compileExpr: unknown kind =", true)
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
	cb := ctx.cb.Typ(t, src)
	compileExpr(ctx, v.Inner[0])
	cb.Call(1)
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
		if op == token.NOT {
			castToBoolExpr(ctx.cb)
		}
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
