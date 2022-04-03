package cl

import (
	"go/types"
	"log"
	"strconv"
	"strings"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/types/parser"
	"github.com/goplus/gox"
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

func toStructType(ctx *blockCtx, t *types.Named, struc *ast.Node, ns string) *types.Struct {
	b := newStructBuilder()
	n := len(struc.Inner)
	for i := 0; i < n; i++ {
		decl := struc.Inner[i]
		switch decl.Kind {
		case ast.FieldDecl:
			if debugCompileDecl {
				log.Println("  => field", decl.Name, "-", decl.Type.QualType)
			}
			typ := toType(ctx, decl.Type, 0)
			if len(decl.Inner) > 0 {
				bits := toInt64(ctx, decl.Inner[0], "non-constant bit field")
				b.BitField(ctx, typ, decl.Name, int(bits))
			} else {
				b.Field(ctx, goNodePos(decl), typ, decl.Name, false)
			}
		case ast.RecordDecl:
			name, _ := ctx.getAsuName(decl, ns)
			typ := compileStructOrUnion(ctx, name, decl)
			for i+1 < n {
				next := struc.Inner[i+1]
				if next.IsImplicit {
					b.Field(ctx, goNodePos(decl), typ, name, true)
					i++
				}
				break
			}
		case ast.IndirectFieldDecl:
		default:
			log.Fatalln("toStructType: unknown field kind =", decl.Kind)
		}
	}
	return b.Type(ctx, t)
}

func toUnionType(ctx *blockCtx, t *types.Named, unio *ast.Node, ns string) types.Type {
	b := newUnionBuilder()
	for _, decl := range unio.Inner {
		switch decl.Kind {
		case ast.FieldDecl:
			if debugCompileDecl {
				log.Println("  => field", decl.Name, "-", decl.Type.QualType)
			}
			typ := toType(ctx, decl.Type, 0)
			b.Field(ctx, goNodePos(decl), typ, decl.Name)
		case ast.RecordDecl:
			name, _ := ctx.getAsuName(decl, ns)
			compileStructOrUnion(ctx, name, decl)
		default:
			log.Fatalln("toUnionType: unknown field kind =", decl.Kind)
		}
	}
	return b.Type(ctx, t)
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

func compileStructOrUnion(ctx *blockCtx, name string, decl *ast.Node) *types.Named {
	if debugCompileDecl {
		log.Println(decl.TagUsed, name, "-", decl.Loc.PresumedLine)
	}
	if name == "" {
		ctx.unnameds[decl.ID] = decl
		return nil
	}
	var inner types.Type
	var t = ctx.cb.NewType(name, goNodePos(decl))
	switch decl.TagUsed {
	case "struct":
		inner = toStructType(ctx, t.Type(), decl, name)
	default:
		inner = toUnionType(ctx, t.Type(), decl, name)
	}
	return t.InitType(ctx.pkg, inner)
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
		newVarAndInit(ctx, scope, typ, decl)
	}
}

func compileVarWith(ctx *blockCtx, typ types.Type, decl *ast.Node) {
	scope := ctx.cb.Scope()
	newVarAndInit(ctx, scope, typ, decl)
}

func newVarAndInit(ctx *blockCtx, scope *types.Scope, typ types.Type, decl *ast.Node) {
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
