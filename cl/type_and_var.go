package cl

import (
	"go/token"
	"go/types"
	"log"
	"strconv"
	"strings"

	ctypes "github.com/goplus/c2go/clang/types"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/types/parser"
	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

func isVariadicFn(typ *ast.Type) bool {
	return strings.HasSuffix(typ.QualType, "...)")
}

func toType(ctx *blockCtx, typ *ast.Type, flags int) types.Type {
	t, _ := toTypeEx(ctx, ctx.cb.Scope(), nil, typ, flags)
	return t
}

func toTypeEx(ctx *blockCtx, scope *types.Scope, tyAnonym types.Type, typ *ast.Type, flags int) (t types.Type, kind int) {
retry:
	t, kind, err := parser.ParseType(ctx.pkg.Types, scope, tyAnonym, typ.QualType, flags)
	if err != nil {
		if e, ok := err.(*parser.TypeNotFound); ok && e.StructOrUnion {
			ctx.uncompls[e.Literal] = ctx.cb.NewType(e.Literal)
			goto retry
		}
		log.Fatalln("toType:", err, "-", typ.QualType)
	}
	return
}

func toStructType(ctx *blockCtx, t *types.Named, struc *ast.Node, ns string) *types.Struct {
	b := newStructBuilder()
	scope := types.NewScope(ctx.cb.Scope(), token.NoPos, token.NoPos, "")
	n := len(struc.Inner)
	for i := 0; i < n; i++ {
		decl := struc.Inner[i]
		switch decl.Kind {
		case ast.FieldDecl:
			if debugCompileDecl {
				log.Println("  => field", decl.Name, "-", decl.Type.QualType)
			}
			typ, _ := toTypeEx(ctx, scope, nil, decl.Type, 0)
			if len(decl.Inner) > 0 {
				bits := toInt64(ctx, decl.Inner[0], "non-constant bit field")
				b.BitField(ctx, typ, decl.Name, int(bits))
			} else {
				b.Field(ctx, goNodePos(decl), typ, decl.Name, false)
			}
		case ast.RecordDecl:
			name, suKind := ctx.getSuName(decl, ns, decl.TagUsed)
			typ := compileStructOrUnion(ctx, name, decl)
			if suKind != suAnonymous {
				if suKind == suNested {
					mangledName := ctypes.MangledName(decl.TagUsed, decl.Name)
					alias := types.NewTypeName(token.NoPos, ctx.pkg.Types, mangledName, typ)
					scope.Insert(alias)
					break
				}
			}
			for i+1 < n {
				next := struc.Inner[i+1]
				if next.Kind == ast.FieldDecl {
					if next.IsImplicit {
						b.Field(ctx, goNodePos(decl), typ, name, true)
						i++
					} else if ret, ok := checkAnonymous(ctx, scope, typ, next); ok {
						b.Field(ctx, goNodePos(next), ret, next.Name, false)
						i++
						continue
					}
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
	scope := types.NewScope(ctx.cb.Scope(), token.NoPos, token.NoPos, "")
	n := len(unio.Inner)
	for i := 0; i < n; i++ {
		decl := unio.Inner[i]
		switch decl.Kind {
		case ast.FieldDecl:
			if debugCompileDecl {
				log.Println("  => field", decl.Name, "-", decl.Type.QualType)
			}
			typ, _ := toTypeEx(ctx, scope, nil, decl.Type, 0)
			b.Field(ctx, goNodePos(decl), typ, decl.Name, false)
		case ast.RecordDecl:
			name, suKind := ctx.getSuName(decl, ns, decl.TagUsed)
			typ := compileStructOrUnion(ctx, name, decl)
			if suKind != suAnonymous {
				if suKind == suNested {
					mangledName := ctypes.MangledName(decl.TagUsed, decl.Name)
					alias := types.NewTypeName(token.NoPos, ctx.pkg.Types, mangledName, typ)
					scope.Insert(alias)
				}
				break
			}
			for i+1 < n {
				next := unio.Inner[i+1]
				if next.Kind == ast.FieldDecl {
					if next.IsImplicit {
						b.Field(ctx, goNodePos(decl), typ, name, true)
						i++
					} else if ret, ok := checkAnonymous(ctx, scope, typ, next); ok {
						b.Field(ctx, goNodePos(next), ret, next.Name, false)
						i++
						continue
					}
				}
				break
			}
		case ast.IndirectFieldDecl:
		default:
			log.Fatalln("toUnionType: unknown field kind =", decl.Kind)
		}
	}
	return b.Type(ctx, t)
}

func checkAnonymous(ctx *blockCtx, scope *types.Scope, typ types.Type, v *ast.Node) (ret types.Type, ok bool) {
	ret, kind := toTypeEx(ctx, scope, typ, v.Type, 0)
	ok = (kind & parser.KindFAnonymous) != 0
	return
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
				if typ, ok := ctx.unnameds[id]; ok {
					aliasType(ctx.cb.Scope(), ctx.pkg.Types, name, typ)
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
	if typ.String() != name {
		ctx.cb.AliasType(name, typ, goNodePos(decl))
	}
}

func compileStructOrUnion(ctx *blockCtx, name string, decl *ast.Node) *types.Named {
	if debugCompileDecl {
		log.Println(decl.TagUsed, name, "-", decl.Loc.PresumedLine)
	}
	t, uncompl := ctx.uncompls[name]
	if !uncompl {
		t = ctx.cb.NewType(name, goNodePos(decl))
	}
	if decl.CompleteDefinition {
		if uncompl {
			delete(ctx.uncompls, name)
		}
		var inner types.Type
		switch decl.TagUsed {
		case "struct":
			inner = toStructType(ctx, t.Type(), decl, name)
		default:
			inner = toUnionType(ctx, t.Type(), decl, name)
		}
		return t.InitType(ctx.pkg, inner)
	} else {
		ctx.uncompls[name] = t
	}
	return t.Type()
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

func compileVarDecl(ctx *blockCtx, decl *ast.Node) {
	if debugCompileDecl {
		log.Println("varDecl", decl.Name, "-", decl.Loc.PresumedLine)
	}
	scope := ctx.cb.Scope()
	typ, kind := toTypeEx(ctx, scope, nil, decl.Type, 0)
	switch decl.StorageClass {
	case ast.Extern:
		scope.Insert(types.NewVar(goNodePos(decl), ctx.pkg.Types, decl.Name, typ))
	default:
		if (kind&parser.KindFConst) != 0 && isInteger(typ) && tryNewConstInteger(ctx, typ, decl) {
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
	if debugCompileDecl {
		log.Println("var", decl.Name, typ, "-", decl.Kind)
	}
	varDecl := ctx.pkg.NewVarEx(scope, goNodePos(decl), typ, decl.Name)
	if len(decl.Inner) > 0 {
		initExpr := decl.Inner[0]
		if ufs, ok := checkUnion(ctx, typ); ok {
			initUnionVar(ctx, decl.Name, ufs, initExpr)
			return
		}
		cb := varDecl.InitStart(ctx.pkg)
		varInit(ctx, typ, initExpr)
		cb.EndInit(1)
	}
}

func varInit(ctx *blockCtx, typ types.Type, initExpr *ast.Node) {
	if initExpr.Kind == ast.InitListExpr {
		initLit(ctx, typ, initExpr)
	} else if !initWithStringLiteral(ctx, typ, initExpr) {
		compileExpr(ctx, initExpr)
	}
}

func initLit(ctx *blockCtx, typ types.Type, initExpr *ast.Node) {
	if debugCompileDecl {
		log.Println("initLit", typ, "-", initExpr.Kind)
	}
	switch t := typ.(type) {
	case *types.Array:
		if !initWithStringLiteral(ctx, typ, initExpr) {
			arrayLit(ctx, t, initExpr)
		}
	case *types.Named:
		structLit(ctx, t, initExpr)
	default:
		compileExpr(ctx, initExpr)
	}
}

func arrayLit(ctx *blockCtx, t *types.Array, decl *ast.Node) {
	elem := t.Elem()
	for _, initExpr := range decl.Inner {
		initLit(ctx, elem, initExpr)
	}
	ctx.cb.ArrayLit(t, len(decl.Inner))
}

func structLit(ctx *blockCtx, typ *types.Named, decl *ast.Node) {
	t := typ.Underlying().(*types.Struct)
	for i, initExpr := range decl.Inner {
		initLit(ctx, t.Field(i).Type(), initExpr)
	}
	ctx.cb.StructLit(typ, len(decl.Inner), false)
}

func checkUnion(ctx *blockCtx, typ types.Type) (ufs *gox.UnionFields, is bool) {
	if t, ok := typ.(*types.Named); ok {
		if vft, ok := ctx.pkg.VFields(t); ok {
			ufs, is = vft.(*gox.UnionFields)
			return
		}
	}
	return nil, false
}

func initUnionVar(ctx *blockCtx, name string, ufs *gox.UnionFields, decl *ast.Node) {
	initExpr := decl.Inner[0]
	t := toType(ctx, initExpr.Type, 0)
	for i, n := 0, ufs.Len(); i < n; i++ {
		fld := ufs.At(i)
		if types.Identical(fld.Type, t) {
			pkg, cb := ctx.pkg, ctx.cb
			scope := cb.Scope()
			obj := scope.Lookup(name)
			global := scope == pkg.Types.Scope()
			if global {
				pkg.NewFunc(nil, "init", nil, nil, false).BodyStart(pkg)
			}
			cb.Val(obj).MemberRef(fld.Name)
			initLit(ctx, t, initExpr)
			cb.Assign(1)
			if global {
				cb.End()
			}
			return
		}
	}
	log.Fatalln("initUnion: init with unexpect type -", t)
}

func isInteger(typ types.Type) bool {
	return isKind(typ, types.IsInteger)
}

func isBool(typ types.Type) bool {
	return isKind(typ, types.IsBoolean)
}

func isKind(typ types.Type, mask types.BasicInfo) bool {
	if t, ok := typ.(*types.Basic); ok {
		return (t.Info() & mask) != 0
	}
	return false
}

func isNilComparable(typ types.Type) bool {
	switch typ.(type) {
	case *types.Pointer, *types.Signature:
		return true
	}
	return false
}

func isFunc(typ types.Type) bool {
	_, ok := typ.(*types.Signature)
	return ok
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
