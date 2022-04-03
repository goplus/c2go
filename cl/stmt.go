package cl

import (
	"go/token"
	"log"

	"github.com/goplus/c2go/clang/ast"
)

// -----------------------------------------------------------------------------

func compileStmt(ctx *blockCtx, stmt *ast.Node) {
	switch stmt.Kind {
	case ast.IfStmt:
		compileIfStmt(ctx, stmt)
	case ast.ForStmt:
		compileForStmt(ctx, stmt)
	case ast.DoStmt:
		compileDoStmt(ctx, stmt)
	case ast.ReturnStmt:
		compileReturnStmt(ctx, stmt)
	case ast.DeclStmt:
		compileDeclStmt(ctx, stmt)
	case ast.NullStmt:
	default:
		compileExprEx(ctx, stmt, "compileStmt: unknown kind =", flagIgnoreResult)
		ctx.cb.EndStmt()
	}
}

func compileDeclStmt(ctx *blockCtx, node *ast.Node) {
	n := len(node.Inner)
	for i := 0; i < n; i++ {
		decl := node.Inner[i]
		switch decl.Kind {
		case ast.VarDecl:
			compileVar(ctx, decl)
		case ast.TypedefDecl:
			compileTypedef(ctx, decl)
		case ast.RecordDecl:
			name, anonymous := ctx.getAsuName(decl, "")
			typ := compileStructOrUnion(ctx, name, decl)
			if !anonymous {
				break
			}
			for i+1 < n {
				next := node.Inner[i+1]
				if next.Kind == ast.VarDecl && isAnonymousType(next) {
					compileVarWith(ctx, typ, next)
					i++
					continue
				}
				break
			}
		default:
			log.Fatalln("compileDeclStmt: unknown kind =", decl.Kind)
		}
	}
}

func compileDoStmt(ctx *blockCtx, stmt *ast.Node) {
	cb := ctx.cb.For().None().Then()
	{
		compileStmt(ctx, stmt.Inner[0])
		cb.If()
		compileExpr(ctx, stmt.Inner[1])
		castToBoolExpr(cb)
		cb.UnaryOp(token.NOT).Then().
			Break(nil).
			End().
			End()
	}
}

func compileForStmt(ctx *blockCtx, stmt *ast.Node) {
	cb := ctx.cb.For()
	if initStmt := stmt.Inner[0]; initStmt.Kind != "" {
		compileStmt(ctx, initStmt)
	}
	if stmt := stmt.Inner[1]; stmt.Kind != "" {
		log.Fatalln("compileForStmt: unexpected -", stmt.Kind)
	}
	if cond := stmt.Inner[2]; cond.Kind != "" {
		compileExpr(ctx, cond)
		castToBoolExpr(cb)
	} else {
		cb.None()
	}
	cb.Then()
	compileStmt(ctx, stmt.Inner[4])
	if postStmt := stmt.Inner[3]; postStmt.Kind != "" {
		cb.Post()
		compileStmt(ctx, postStmt)
	}
	cb.End()
}

func compileIfStmt(ctx *blockCtx, stmt *ast.Node) {
	cb := ctx.cb
	cb.If()
	compileExpr(ctx, stmt.Inner[0])
	castToBoolExpr(cb)
	cb.Then()
	compileStmt(ctx, stmt.Inner[1])
	if stmt.HasElse {
		cb.Else()
		compileStmt(ctx, stmt.Inner[2])
	}
	cb.End()
}

func compileReturnStmt(ctx *blockCtx, stmt *ast.Node) {
	n := len(stmt.Inner)
	if n > 0 {
		n = 1
		compileExpr(ctx, stmt.Inner[0])
	}
	ctx.cb.Return(n, goNode(stmt))
}

func compileCompoundStmt(ctx *blockCtx, stmts *ast.Node) {
	for _, stmt := range stmts.Inner {
		compileStmt(ctx, stmt)
	}
}

// -----------------------------------------------------------------------------
