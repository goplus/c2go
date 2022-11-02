package cl

import (
	"go/constant"
	"go/token"
	"go/types"
	"log"
	"strconv"

	ctypes "github.com/goplus/c2go/clang/types"

	"github.com/goplus/c2go/clang/ast"
	"github.com/goplus/c2go/clang/types/parser"
	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

func toType(ctx *blockCtx, typ *ast.Type, flags int) types.Type {
	t, _ := toTypeEx(ctx, ctx.cb.Scope(), nil, typ, flags, false)
	return t
}

func toTypeEx(ctx *blockCtx, scope *types.Scope, tyAnonym types.Type, typ *ast.Type, flags int, pub bool) (t types.Type, kind int) {
	t, kind, err := parseType(ctx, scope, tyAnonym, typ, flags, pub)
	if err != nil {
		log.Panicln("toType:", err, "-", typ.QualType)
	}
	return
}

func parseType(ctx *blockCtx, scope *types.Scope, tyAnonym types.Type, typ *ast.Type, flags int, pub bool) (t types.Type, kind int, err error) {
	conf := &parser.Config{
		Scope: scope, Flags: flags, Anonym: tyAnonym, ParseEnv: ctx,
	}
retry:
	t, kind, err = parser.ParseType(typ.QualType, conf)
	if err != nil {
		if e, ok := err.(*parser.TypeNotFound); ok && e.StructOrUnion {
			name := e.Literal
			if pub {
				name = gox.CPubName(name)
			}
			newStructOrUnionType(ctx, token.NoPos, name)
			if pub {
				substObj(ctx.pkg.Types, scope, e.Literal, scope.Lookup(name))
			}
			goto retry
		}
	}
	return
}

func toAnonymType(ctx *blockCtx, pos token.Pos, decl *ast.Node) (ret *types.Named) {
	scope := types.NewScope(ctx.cb.Scope(), token.NoPos, token.NoPos, "")
	switch decl.Kind {
	case ast.FieldDecl:
		pkg := ctx.pkg
		typ, _ := toTypeEx(ctx, scope, nil, decl.Type, 0, false)
		fld := types.NewField(ctx.goNodePos(decl), pkg.Types, decl.Name, typ, false)
		struc := types.NewStruct([]*types.Var{fld}, nil)
		ret = pkg.NewType(ctx.getAnonyName(), pos).InitType(pkg, struc)
	default:
		log.Panicln("toAnonymType: unknown kind -", decl.Kind)
	}
	return
}

func checkFieldName(pname *string, pub bool) {
	if pub {
		*pname = gox.CPubName(*pname)
	} else {
		avoidKeyword(pname)
	}
}

func toStructType(ctx *blockCtx, t *types.Named, struc *ast.Node, pub bool) (ret *types.Struct, dels delfunc) {
	b := newStructBuilder()
	scope := types.NewScope(ctx.cb.Scope(), token.NoPos, token.NoPos, "")
	n := len(struc.Inner)
	for i := 0; i < n; i++ {
		decl := struc.Inner[i]
		switch decl.Kind {
		case ast.FieldDecl:
			name := decl.Name
			if debugCompileDecl {
				log.Println("  => field", name, "-", decl.Type.QualType)
			}
			if name != "" {
				checkFieldName(&name, pub)
			}
			typ, _ := toTypeEx(ctx, scope, nil, decl.Type, parser.FlagIsStructField, false)
			if decl.IsBitfield {
				bits := toInt64(ctx, decl.Inner[0], "non-constant bit field")
				b.BitField(ctx, typ, name, int(bits))
			} else {
				b.Field(ctx, ctx.goNodePos(decl), typ, name, false)
			}
		case ast.RecordDecl:
			name, suKind := ctx.getSuName(decl, decl.TagUsed)
			typ, del := compileStructOrUnion(ctx, name, decl, pub)
			if suKind != suAnonymous {
				break
			}
			dels = append(dels, name)
			dels = append(dels, del...)
			for i+1 < n {
				next := struc.Inner[i+1]
				if next.Kind == ast.FieldDecl {
					if next.IsImplicit {
						b.Field(ctx, ctx.goNodePos(decl), typ, name, true)
						i++
					} else if ret, ok := checkAnonymous(ctx, scope, typ, next); ok {
						checkFieldName(&next.Name, pub)
						b.Field(ctx, ctx.goNodePos(next), ret, next.Name, false)
						i++
						continue
					}
				}
				break
			}
		case ast.IndirectFieldDecl, ast.MaxFieldAlignmentAttr, ast.AlignedAttr, ast.PackedAttr:
		default:
			log.Panicln("toStructType: unknown field kind =", decl.Kind)
		}
	}
	ret = b.Type(ctx, t)
	return
}

func toUnionType(ctx *blockCtx, t *types.Named, unio *ast.Node, pub bool) (ret types.Type, dels delfunc) {
	b := newUnionBuilder()
	scope := types.NewScope(ctx.cb.Scope(), token.NoPos, token.NoPos, "")
	n := len(unio.Inner)
	for i := 0; i < n; i++ {
		decl := unio.Inner[i]
		switch decl.Kind {
		case ast.FieldDecl:
			name := decl.Name
			if debugCompileDecl {
				log.Println("  => field", name, "-", decl.Type.QualType)
			}
			checkFieldName(&name, pub)
			typ, _ := toTypeEx(ctx, scope, nil, decl.Type, 0, false)
			b.Field(ctx, ctx.goNodePos(decl), typ, name, false)
		case ast.RecordDecl:
			name, suKind := ctx.getSuName(decl, decl.TagUsed)
			typ, del := compileStructOrUnion(ctx, name, decl, pub)
			if suKind != suAnonymous {
				break
			}
			dels = append(dels, name)
			dels = append(dels, del...)
			for i+1 < n {
				next := unio.Inner[i+1]
				if next.Kind == ast.FieldDecl {
					if next.IsImplicit {
						b.Field(ctx, ctx.goNodePos(decl), typ, name, true)
						i++
					} else if ret, ok := checkAnonymous(ctx, scope, typ, next); ok {
						checkFieldName(&next.Name, pub)
						b.Field(ctx, ctx.goNodePos(next), ret, next.Name, false)
						i++
						continue
					}
				}
				break
			}
		case ast.IndirectFieldDecl, ast.MaxFieldAlignmentAttr, ast.AlignedAttr, ast.PackedAttr:
		default:
			log.Panicln("toUnionType: unknown field kind =", decl.Kind)
		}
	}
	ret = b.Type(ctx, t)
	return
}

func checkAnonymous(ctx *blockCtx, scope *types.Scope, typ types.Type, v *ast.Node) (ret types.Type, ok bool) {
	ret, kind := toTypeEx(ctx, scope, typ, v.Type, 0, false)
	ok = (kind & parser.KindFAnonymous) != 0
	return
}

// -----------------------------------------------------------------------------

func compileTypedef(ctx *blockCtx, decl *ast.Node, global, pub bool) types.Type {
	name, qualType := decl.Name, decl.Type.QualType
	if debugCompileDecl {
		log.Println("typedef", name, "-", qualType)
	}
	if global && ctx.checkExists(name) {
		if len(decl.Inner) > 0 { // check to delete unnamed types
			item := decl.Inner[0]
			if item.Kind == "ElaboratedType" {
				if owned := item.OwnedTagDecl; owned != nil && owned.Name == "" && owned.Kind != ast.EnumDecl {
					ctx.deleteUnnamed(owned.ID)
				}
			}
		}
		return nil
	}
	scope := ctx.cb.Scope()
	if len(decl.Inner) > 0 {
		item := decl.Inner[0]
		if item.Kind == "ElaboratedType" {
			if owned := item.OwnedTagDecl; owned != nil && owned.Name == "" {
				var typ types.Type
				if owned.Kind == ast.EnumDecl {
					typ = ctypes.Int
				} else if u, ok := ctx.unnameds[owned.ID]; ok {
					typ = u.typ
				} else {
					log.Panicf("compileTypedef %v: unknown id = %v\n", name, owned.ID)
				}
				ctx.cb.AliasType(name, typ, ctx.goNodePos(decl))
				return typ
			}
		}
	}
	typ, _ := toTypeEx(ctx, scope, nil, decl.Type, parser.FlagIsTypedef, pub)
	if isArrayUnknownLen(typ) || typ == ctypes.Void {
		aliasType(scope, ctx.pkg.Types, name, typ)
		return nil
	}
	if global {
		if old := scope.Lookup(name); old != nil {
			if types.Identical(typ, old.Type()) {
				return nil
			}
		}
	}
	ctx.cb.AliasType(name, typ, ctx.goNodePos(decl))
	return typ
}

func newStructOrUnionType(ctx *blockCtx, pos token.Pos, name string) (t *gox.TypeDecl) {
	t, decled := ctx.typdecls[name]
	if !decled {
		t = ctx.cb.NewType(name, pos)
		ctx.typdecls[name] = t
	}
	return
}

func compileStructOrUnion(ctx *blockCtx, name string, decl *ast.Node, pub bool) (*types.Named, delfunc) {
	if debugCompileDecl {
		log.Println(decl.TagUsed, name, "-", decl.Loc.PresumedLine)
	}
	var t *gox.TypeDecl
	pos := ctx.goNodePos(decl)
	pkg := ctx.pkg
	if ctx.inSrcFile() && decl.Name != "" {
		realName := ctx.autoStaticName(decl.Name)
		var scope = ctx.cb.Scope()
		t = ctx.cb.NewType(realName, pos)
		substObj(pkg.Types, scope, name, t.Type().Obj())
	} else {
		t = newStructOrUnionType(ctx, pos, name)
	}
	if decl.CompleteDefinition {
		var inner types.Type
		var del delfunc
		switch decl.TagUsed {
		case "struct":
			inner, del = toStructType(ctx, t.Type(), decl, pub)
		default:
			inner, del = toUnionType(ctx, t.Type(), decl, pub)
		}
		ret := t.InitType(pkg, inner)
		if pub {
			pkg.ExportFields(ret)
		}
		return ret, del
	}
	return t.Type(), nil
}

func compileEnum(ctx *blockCtx, decl *ast.Node, global bool) {
	inner := decl.Inner
	if global && len(inner) > 0 && ctx.checkExists(inner[0].Name) {
		return
	}
	scope := ctx.cb.Scope()
	cdecl := ctx.pkg.NewConstDefs(scope)
	iotav := 0
	for _, item := range inner {
		iotav = compileEnumConst(ctx, cdecl, item, iotav)
	}
}

func compileEnumConst(ctx *blockCtx, cdecl *gox.ConstDefs, v *ast.Node, iotav int) int {
	fn := func(cb *gox.CodeBuilder) int {
		if len(v.Inner) > 0 {
			compileExpr(ctx, v.Inner[0])
			cval := cb.Get(-1).CVal
			if cval == nil {
				log.Panicln("compileEnumConst: not a constant expression")
			}
			ival, ok := constant.Int64Val(cval)
			if !ok {
				log.Panicln("compileEnumConst: not a integer constant")
			}
			iotav = int(ival)
		} else {
			cb.Val(iotav)
		}
		return 1
	}
	name := v.Name
	if ctx.inSrcFile() {
		name = ctx.srcEnumName(v.Name)
		ctx.srcenums[v.ID] = name
	}
	cdecl.New(fn, iotav, ctx.goNodePos(v), ctypes.Int, name)
	return iotav + 1
}

func compileVarDecl(ctx *blockCtx, decl *ast.Node, global bool) {
	origName, rewritten := decl.Name, false
	if debugCompileDecl {
		log.Println("varDecl", origName, "-", decl.Loc.PresumedLine)
	}
	if global {
		rewritten = ctx.getPubName(&decl.Name)
	}
	origScope := ctx.cb.Scope()
	scope := origScope
	flags := 0
	gblStatic := false
	switch decl.StorageClass {
	case ast.Extern:
		flags = parser.FlagIsExtern
	case ast.Static:
		if global { // global static
			gblStatic = true
		} else { // local static variable
			scope = ctx.pkg.Types.Scope() // Go don't have local static variable, change to global
		}
		decl.Name, rewritten = ctx.autoStaticName(origName), true
	}
	typ, kind, err := parseType(ctx, scope, nil, decl.Type, flags, false)
	if err != nil {
		if gblStatic && parser.IsArrayWithoutLen(err) {
			return
		}
		log.Panicln("parseType:", err, "-", decl.Type)
	}
	avoidKeyword(&decl.Name)
	if flags == parser.FlagIsExtern {
		scope.Insert(types.NewVar(ctx.goNodePos(decl), ctx.pkg.Types, decl.Name, typ))
	} else {
		if (kind&parser.KindFConst) != 0 && isInteger(typ) && tryNewConstInteger(ctx, typ, decl) {
			if rewritten {
				substObj(ctx.pkg.Types, origScope, origName, origScope.Lookup(decl.Name))
			}
			return
		}
		newVarAndInit(ctx, scope, typ, decl, global)
		if rewritten {
			substObj(ctx.pkg.Types, ctx.cb.Scope(), origName, scope.Lookup(decl.Name))
		} else if kind == parser.KindFVolatile && !global {
			addr := gox.Lookup(scope, decl.Name)
			ctx.cb.VarRef(nil).Val(addr).Assign(1) // musl: use volatile to mark unused
		}
	}
}

func substObj(pkg *types.Package, scope *types.Scope, origName string, real types.Object) {
	old := scope.Insert(gox.NewSubst(token.NoPos, pkg, origName, real))
	if old != nil {
		if t, ok := old.Type().(*gox.SubstType); ok {
			t.Real = real
		} else {
			log.Panicln(origName, "redefined")
		}
	}
}

func avoidKeyword(name *string) {
	switch *name {
	case "map", "type", "range", "chan", "var", "func", "go", "select",
		"defer", "package", "import", "interface", "fallthrough":
		*name += "_"
	}
}

func compileVarWith(ctx *blockCtx, typ types.Type, decl *ast.Node) {
	scope := ctx.cb.Scope()
	newVarAndInit(ctx, scope, typ, decl, false)
}

func compileVarDef(ctx *blockCtx, decl *ast.Node) {
	if debugCompileDecl {
		log.Println("varDef", decl.Name, "-", decl.Loc.PresumedLine)
	}
	typ := toType(ctx, decl.Type, 0)
	avoidKeyword(&decl.Name)
	cb := ctx.cb.DefineVarStart(ctx.goNodePos(decl), decl.Name).Typ(typ)
	if inner := decl.Inner; len(inner) > 0 {
		initExpr := inner[0]
		varInit(ctx, typ, initExpr)
	} else {
		cb.ZeroLit(typ)
	}
	cb.Call(1).EndInit(1)
}

func newVarAndInit(ctx *blockCtx, scope *types.Scope, typ types.Type, decl *ast.Node, global bool) {
	if debugCompileDecl {
		log.Println("var", decl.Name, typ, "-", decl.Kind)
	}
	varDecl, inVBlock := ctx.newVar(scope, ctx.goNodePos(decl), typ, decl.Name)
	inner := decl.Inner
	if len(inner) == 1 && inner[0].Kind == ast.VisibilityAttr {
		inner = inner[1:]
	}
	if len(inner) > 0 {
		initExpr := inner[0]
		if ufs, ok := checkUnion(ctx, typ); ok {
			if inVBlock {
				log.Panicln("TODO: initUnionVar inVBlock")
			}
			initUnionVar(ctx, decl.Name, ufs, initExpr)
			return
		}
		if inVBlock {
			varAssign(ctx, scope, typ, decl.Name, initExpr)
		} else if global && hasFnPtrMember(typ) {
			pkg := ctx.pkg
			cb := pkg.NewFunc(nil, "init", nil, nil, false).BodyStart(pkg)
			varAssign(ctx, scope, typ, decl.Name, initExpr)
			cb.End()
		} else {
			cb := varDecl.InitStart(ctx.pkg)
			varInit(ctx, typ, initExpr)
			cb.EndInit(1)
		}
	} else if inVBlock {
		addr := gox.Lookup(scope, decl.Name)
		ctx.cb.VarRef(addr).ZeroLit(typ).Assign(1)
	}
}

func hasFnPtrMember(typ types.Type) bool {
retry:
	switch t := typ.Underlying().(type) {
	case *types.Struct:
		for i, n := 0, t.NumFields(); i < n; i++ {
			if isFunc(t.Field(i).Type()) {
				return true
			}
		}
	case *types.Array:
		typ = t.Elem()
		goto retry
	}
	return false
}

func varAssign(ctx *blockCtx, scope *types.Scope, typ types.Type, name string, initExpr *ast.Node) {
	addr := gox.Lookup(scope, name)
	cb := ctx.cb.VarRef(addr)
	varInit(ctx, typ, initExpr)
	cb.Assign(1)
}

func varInit(ctx *blockCtx, typ types.Type, initExpr *ast.Node) {
	if initExpr.Kind == ast.InitListExpr {
		initLit(ctx, typ, initExpr)
	} else if !initWithStringLiteral(ctx, typ, initExpr) {
		compileExpr(ctx, initExpr)
	}
}

func initLit(ctx *blockCtx, typ types.Type, initExpr *ast.Node) int {
	switch t := typ.(type) {
	case *types.Array:
		if !initWithStringLiteral(ctx, typ, initExpr) {
			arrayLit(ctx, t, initExpr)
		}
	case *types.Named:
		structLit(ctx, t, initExpr)
	case *bfType:
		if initExpr.Kind != ast.ImplicitValueInitExpr {
			log.Panicln("initLit bfType: TODO")
		}
		if !t.first {
			return 0
		}
		ctx.cb.ZeroLit(t.Type)
	default:
		compileExpr(ctx, initExpr)
	}
	return 1
}

func arrayLit(ctx *blockCtx, t *types.Array, decl *ast.Node) {
	var inits []*ast.Node
	if len(decl.ArrayFiller) > 0 {
		idx := 0
		if decl.ArrayFiller[idx].Kind == ast.ImplicitValueInitExpr {
			idx = 1
		}
		inits = decl.ArrayFiller[idx:]
	} else {
		inits = decl.Inner
	}
	elem := t.Elem()
	for _, initExpr := range inits {
		initLit(ctx, elem, initExpr)
	}
	ctx.cb.ArrayLit(t, len(inits))
}

func structLit(ctx *blockCtx, typ *types.Named, decl *ast.Node) {
	t := ctx.getVStruct(typ)
	n := 0
	for i, initExpr := range decl.Inner {
		n += initLit(ctx, t.Field(i).Type(), initExpr)
	}
	ctx.cb.StructLit(typ, n, false)
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
		if ctypes.Identical(fld.Type, t) {
			pkg, cb := ctx.pkg, ctx.cb
			scope := cb.Scope()
			obj := gox.Lookup(scope, name)
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
	log.Panicln("initUnion: init with unexpect type -", t)
}

const (
	ncKindInvalid = iota
	ncKindPointer
	ncKindUnsafePointer
	ncKindSignature
)

func checkNilComparable(v *gox.Element) int {
	switch t := v.Type.(type) {
	case *types.Pointer:
		return ncKindPointer
	case *types.Basic:
		switch t.Kind() {
		case types.UnsafePointer:
			return ncKindUnsafePointer
		}
	case *types.Signature:
		return ncKindSignature
	}
	return ncKindInvalid
}

func isNilComparable(typ types.Type) bool {
	switch t := typ.(type) {
	case *types.Pointer, *types.Signature:
		return true
	case *types.Basic:
		if t.Kind() == types.UnsafePointer {
			return true
		}
	}
	return false
}

func isIntegerOrBool(typ types.Type) bool {
	return isKind(typ, types.IsInteger|types.IsBoolean)
}

func isNumber(typ types.Type) bool {
	return isKind(typ, types.IsInteger|types.IsFloat)
}

func isUnsigned(typ types.Type) bool {
	return isKind(typ, types.IsUnsigned)
}

func isInteger(typ types.Type) bool {
	return isKind(typ, types.IsInteger)
}

func isUntyped(typ types.Type) bool {
	return isKind(typ, types.IsUntyped)
}

func isFloat(typ types.Type) bool {
	return isKind(typ, types.IsFloat)
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

func isArrayUnknownLen(typ types.Type) bool {
	if t, ok := typ.(*types.Array); ok {
		return t.Len() < 0
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
		case ast.InitListExpr:
			inner := decl.Inner
			if len(inner) != 1 || inner[0].Kind != ast.StringLiteral {
				break
			}
			decl = inner[0]
			fallthrough
		case ast.StringLiteral:
			s, err := strconv.Unquote(decl.Value.(string))
			if err != nil {
				log.Panicln("initWithStringLiteral:", err)
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
