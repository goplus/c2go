package cl

import (
	"go/types"
	"log"
	"strconv"
	"strings"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/types/parser"
	"github.com/goplus/gox"

	ctypes "github.com/goplus/c2go/clang/types"
)

// -----------------------------------------------------------------------------

func isVariadicFn(typ *ast.Type) bool {
	return strings.HasSuffix(typ.QualType, "...)")
}

func toType(ctx *blockCtx, typ *ast.Type, flags int) types.Type {
	t, _ := toTypeEx(ctx, typ, flags)
	return t
}

func toTypeEx(ctx *blockCtx, typ *ast.Type, flags int) (t types.Type, isConst bool) {
	t, isConst, err := parser.ParseType(ctx, ctx.fset, typ.QualType, flags)
	if err != nil {
		log.Fatalln("toType:", err, "-", typ.QualType)
	}
	return
}

func toStructType(ctx *blockCtx, t *types.Named, struc *ast.Node) *types.Struct {
	b := newStructBuilder()
	for _, item := range struc.Inner {
		switch item.Kind {
		case ast.FieldDecl:
			if debugCompileDecl {
				log.Println("  => field", item.Name, "-", item.Type.QualType)
			}
			typ := toType(ctx, item.Type, 0)
			if len(item.Inner) > 0 {
				bits := toInt64(ctx, item.Inner[0], "non-constant bit field")
				b.BitField(ctx, typ, item.Name, int(bits))
			} else {
				b.Field(ctx, goNodePos(item), typ, item.Name)
			}
		default:
			log.Fatalln("toStructType: unknown field kind =", item.Kind)
		}
	}
	return b.Type(ctx, t)
}

func toUnionType(ctx *blockCtx, decl *ast.Node) types.Type {
	// TODO: union
	return ctypes.NotImpl
}

// -----------------------------------------------------------------------------

func compileTypedef(ctx *blockCtx, decl *ast.Node) {
	name, qualType := decl.Name, decl.Type.QualType
	if debugCompileDecl {
		log.Println("typedef", name, "-", qualType, decl.Loc.PresumedLine)
	}
	if len(decl.Inner) > 0 {
		item := decl.Inner[0]
		if item.Kind == "ElaboratedType" {
			if owned := item.OwnedTagDecl; owned != nil && owned.Name == "" {
				if owned.Kind == ast.EnumDecl {
					ctx.cb.AliasType(name, tyInt, goNodePos(decl))
					return
				}
				id := owned.ID
				if detail, ok := ctx.unnameds[id]; ok {
					compileStructOrUnion(ctx, name, detail)
					return
				}
				log.Fatalln("compileTypedef: unknown id =", id)
			}
		}
	}
	typ := toType(ctx, decl.Type, 0)
	if types.Identical(typ, ctx.tyValist) {
		aliasType(ctx.cb.Scope(), ctx.pkg.Types, name, typ)
		return
	}
	ctx.cb.AliasType(name, typ, goNodePos(decl))
}

func compileStructOrUnion(ctx *blockCtx, name string, decl *ast.Node) {
	if debugCompileDecl {
		log.Println(decl.TagUsed, name, "-", decl.Loc.PresumedLine)
	}
	if name == "" {
		ctx.unnameds[decl.ID] = decl
		return
	}
	var inner types.Type
	var pkg = ctx.pkg
	var t = pkg.NewType(name, goNodePos(decl))
	switch decl.TagUsed {
	case "struct":
		inner = toStructType(ctx, t.Type(), decl)
	default:
		inner = toUnionType(ctx, decl)
	}
	t.InitType(pkg, inner)
}

func compileEnum(ctx *blockCtx, decl *ast.Node) {
	scope := ctx.cb.Scope()
	cdecl := ctx.pkg.NewConstDecl(scope)
	iotav := 0
	for _, item := range decl.Inner {
		iotav = compileEnumConst(ctx, cdecl, item, iotav)
	}
}

func compileEnumConst(ctx *blockCtx, cdecl *gox.ConstDecl, v *ast.Node, iotav int) int {
	fn := func(cb *gox.CodeBuilder) int {
		if v.Value != nil {
			log.Fatalln("compileEnumConst: TODO -", v.Name)
		}
		cb.Val(iotav)
		return 1
	}
	cdecl.New(fn, iotav, goNodePos(v), tyInt, v.Name)
	return iotav + 1
}

func compileVar(ctx *blockCtx, decl *ast.Node) {
	if debugCompileDecl {
		log.Println("var", decl.Name, "-", decl.Loc.PresumedLine)
	}
	scope := ctx.cb.Scope()
	typ, isConst := toTypeEx(ctx, decl.Type, 0)
	switch decl.StorageClass {
	case ast.Extern:
		scope.Insert(types.NewVar(goNodePos(decl), ctx.pkg.Types, decl.Name, typ))
	default:
		if isConst && isInteger(typ) && tryNewConstInteger(ctx, typ, decl) {
			return
		}
		if isValistType(ctx, typ) { // skip valist variable
			return
		}
		varDecl := ctx.pkg.NewVarEx(scope, goNodePos(decl), typ, decl.Name)
		if len(decl.Inner) > 0 {
			cb := varDecl.InitStart(ctx.pkg)
			initExpr := decl.Inner[0]
			if !initWithStringLiteral(ctx, typ, initExpr) {
				compileExpr(ctx, initExpr)
			}
			cb.EndInit(1)
		}
	}
}

func isInteger(typ types.Type) bool {
	if t, ok := typ.(*types.Basic); ok {
		return (t.Info() & types.IsInteger) != 0
	}
	return false
}

// char[N], char[], unsigned char[N], unsigned char[]
func isCharArray(typ types.Type) bool {
	if t, ok := typ.(*types.Array); ok {
		switch t.Elem() {
		case types.Typ[types.Int8], types.Typ[types.Uint8]:
			return true
		}
	}
	return false
}

func initWithStringLiteral(ctx *blockCtx, typ types.Type, decl *ast.Node) bool {
	if isCharArray(typ) {
		switch decl.Kind {
		case ast.StringLiteral:
			s, err := strconv.Unquote(decl.Value.(string))
			if err != nil {
				log.Fatalln("initWithStringLiteral:", err)
			}
			stringLit(ctx.cb, s, typ)
			return true
		}
	}
	return false
}

func tryNewConstInteger(ctx *blockCtx, typ types.Type, decl *ast.Node) bool {
	if len(decl.Inner) > 0 {
		initExpr := decl.Inner[0]
		switch initExpr.Kind {
		case ast.IntegerLiteral:
			cb := ctx.cb.NewConstStart(typ, decl.Name)
			compileExpr(ctx, initExpr)
			cb.EndInit(1)
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
