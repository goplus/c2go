package cl

import (
	"encoding/json"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/goplus/gox"

	cast "github.com/goplus/c2go/clang/ast"
	ctypes "github.com/goplus/c2go/clang/types"
)

// -----------------------------------------------------------------------------

const builtin_decls = `{
	"__sync_synchronize": "void ()",
	"__builtin_bswap32": "uint32 (uint32)",
	"__builtin_bswap64": "uint64 (uint64)",
	"__builtin___memset_chk": "void* (void*, int32, size_t, size_t)",
	"__builtin___memcpy_chk": "void* (void*, void*, size_t, size_t)",
	"__builtin___memmove_chk": "void* (void*, void*, size_t, size_t)",
	"__builtin___strlcpy_chk": "size_t (char*, char*, size_t, size_t)",
	"__builtin___strlcat_chk": "size_t (char*, char*, size_t, size_t)",
	"__builtin_object_size": "size_t (void*, int32)",
	"__builtin_fabsf": "float32 (float32)",
	"__builtin_fabsl": "float64 (float64)",
	"__builtin_fabs": "float64 (float64)",
	"__builtin_nanf": "float32 (char*)",
	"__builtin_nanl": "float64 (char*)",
	"__builtin_nan": "float64 (char*)",
	"__builtin_huge_valf": "float32 ()",
	"__builtin_inff": "float32 ()",
	"__builtin_infl": "float64 ()",
	"__builtin_inf": "float64 ()",
	"__atomic_store_n_u16": "void (uint16*, int, uint16)",
	"__atomic_store_n_i16": "void (int16*, int, int16)",
	"__atomic_store_n_u32": "void (uint32*, int, uint32)",
	"__atomic_store_n_i32": "void (int32*, int, int32)",
	"__atomic_store_n_u64": "void (uint64*, int, uint64)",
	"__atomic_store_n_i64": "void (int64*, int, int64)",
	"__atomic_load_n_u16": "uint16 (uint16*, int)",
	"__atomic_load_n_i16": "int16 (int16*, int)",
	"__atomic_load_n_u32": "uint32 (uint32*, int)",
	"__atomic_load_n_i32": "int32 (int32*, int)",
	"__atomic_load_n_u64": "uint64 (uint64*, int)",
	"__atomic_load_n_i64": "int64 (int64*, int)"
}`

type overloadFn struct {
	name      string
	overloads []string
}

var (
	builtin_overloads = []overloadFn{
		{name: "__atomic_store_n", overloads: []string{
			"__atomic_store_n_u16", "__atomic_store_n_u32", "__atomic_store_n_u64",
			"__atomic_store_n_i16", "__atomic_store_n_i32", "__atomic_store_n_i64",
		}},
		{name: "__atomic_load_n", overloads: []string{
			"__atomic_load_n_u16", "__atomic_load_n_u32", "__atomic_load_n_u64",
			"__atomic_load_n_i16", "__atomic_load_n_i32", "__atomic_load_n_i64",
		}},
	}
)

func decl_builtin(ctx *blockCtx) {
	var fns map[string]string
	err := json.NewDecoder(strings.NewReader(builtin_decls)).Decode(&fns)
	if err != nil {
		log.Panicln("decl_builtin decode error:", err)
	}
	bfm := ctx.bfm
	pkg := ctx.pkg.Types
	scope := pkg.Scope()
	if bfm != BFM_FromLibC {
		for fn, proto := range fns {
			t := toType(ctx, &cast.Type{QualType: strings.ReplaceAll(proto, "size_t", "unsigned long")}, 0)
			origFn := fn
			if bfm == BFM_InLibC {
				fn = "X" + fn
			}
			fnObj := types.NewFunc(token.NoPos, pkg, fn, t.(*types.Signature))
			scope.Insert(fnObj)
			if bfm == BFM_InLibC {
				substObj(pkg, scope, origFn, fnObj)
			}
		}
	}
	for _, o := range builtin_overloads {
		fns := make([]types.Object, len(o.overloads))
		for i, item := range o.overloads {
			switch bfm {
			case BFM_InLibC:
				item = "X" + item
			}
			fns[i] = pkg.Scope().Lookup(item)
		}
		scope.Insert(gox.NewOverloadFunc(token.NoPos, pkg, o.name, fns...))
	}
}

// -----------------------------------------------------------------------------

type unionBuilder struct {
	fields []*gox.UnionField
}

func newUnionBuilder() *unionBuilder {
	return &unionBuilder{}
}

func unionEmbeddedField(ctx *blockCtx, fields []*gox.UnionField, t *types.Named, off int) []*gox.UnionField {
	o := t.Underlying().(*types.Struct)
	for i, n := 0, o.NumFields(); i < n; i++ {
		fld := o.Field(i)
		fldType := fld.Type()
		if fld.Embedded() {
			fields = unionEmbeddedField(ctx, fields, fldType.(*types.Named), off)
		} else {
			fields = append(fields, &gox.UnionField{
				Name: fld.Name(),
				Off:  off,
				Type: fldType,
			})
		}
		off += ctx.sizeof(fldType)
	}
	return fields
}

func (p *unionBuilder) Type(ctx *blockCtx, t *types.Named) *types.Struct {
	var fldLargest *gox.UnionField
	var fields = p.fields
	var lenLargest, n = 0, len(fields)
	for i := 0; i < n; i++ {
		fld := fields[i]
		if len := ctx.sizeof(fld.Type); len > lenLargest {
			fldLargest, lenLargest = fld, len
		}
		if fld.Name == "" { // embedded
			fields = unionEmbeddedField(ctx, fields, fld.Type.(*types.Named), 0)
		}
	}
	flds := make([]*types.Var, 0, 1)
	if fldLargest != nil {
		pkg := ctx.pkg
		pkg.SetVFields(t, gox.NewUnionFields(fields))
		fld := types.NewField(fldLargest.Pos, pkg.Types, fldLargest.Name, fldLargest.Type, false)
		flds = append(flds, fld)
	}
	return types.NewStruct(flds, nil)
}

func (p *unionBuilder) Field(ctx *blockCtx, pos token.Pos, typ types.Type, name string, embedded bool) {
	if embedded {
		name = ""
	}
	fld := &gox.UnionField{
		Name: name,
		Type: typ,
		Pos:  pos,
	}
	p.fields = append(p.fields, fld)
}

// -----------------------------------------------------------------------------

type structBuilder struct {
	fields      []*types.Var
	bitFields   []*gox.BitField
	lastFldName string
	lastTy      types.Type
	totalBits   int
	leftBits    int
	idx         int
}

func newStructBuilder() *structBuilder {
	return &structBuilder{leftBits: -1}
}

func (p *structBuilder) Type(ctx *blockCtx, t *types.Named) *types.Struct {
	struc := types.NewStruct(p.fields, nil)
	if len(p.bitFields) > 0 {
		ctx.pkg.SetVFields(t, gox.NewBitFields(p.bitFields))
	}
	return struc
}

func (p *structBuilder) BitField(ctx *blockCtx, typ types.Type, name string, bits int) {
	if p.leftBits >= bits && ctypes.Identical(typ, p.lastTy) {
		if name != "" {
			p.bitFields = append(p.bitFields, &gox.BitField{
				Name:    name,
				FldName: p.lastFldName,
				Off:     p.totalBits - p.leftBits,
				Bits:    bits,
			})
		}
		p.leftBits -= bits
	} else if p.totalBits = ctx.sizeof(typ) << 3; p.totalBits >= bits {
		fldName := "Xbf_" + strconv.Itoa(p.idx)
		p.Field(ctx, token.NoPos, typ, fldName, false)
		p.idx++
		p.lastFldName = fldName
		p.lastTy = typ
		p.leftBits = p.totalBits - bits
		if name != "" {
			p.bitFields = append(p.bitFields, &gox.BitField{
				Name:    name,
				FldName: p.lastFldName,
				Bits:    bits,
			})
		}
	} else {
		p.leftBits = -1
		log.Panicln("BitField - too large bits:", bits)
	}
}

func (p *structBuilder) Field(ctx *blockCtx, pos token.Pos, typ types.Type, name string, embedded bool) {
	fld := types.NewField(pos, ctx.pkg.Types, name, typ, embedded)
	p.fields = append(p.fields, fld)
	p.leftBits = -1
}

// -----------------------------------------------------------------------------

func toInt64(ctx *blockCtx, v *cast.Node, emsg string) int64 {
	cb := ctx.pkg.ConstStart()
	compileExpr(ctx, v)
	tv := cb.EndConst()
	if val := tv.CVal; val != nil {
		if val.Kind() == constant.Float {
			if v, ok := constant.Val(val).(*big.Rat); ok && v.IsInt() {
				return v.Num().Int64()
			}
		} else if v, ok := constant.Int64Val(val); ok {
			return v
		}
	}
	log.Panicln(emsg)
	return -1
}

// -----------------------------------------------------------------------------

func typeCast(ctx *blockCtx, typ types.Type, arg *gox.Element) {
	if !ctypes.Identical(typ, arg.Type) {
		adjustIntConst(ctx, arg, typ)
		*arg = *ctx.cb.Typ(typ).Val(arg).Call(1).InternalStack().Pop()
	}
}

func assign(ctx *blockCtx, src ast.Node) {
	cb := ctx.cb
	arg1 := cb.Get(-2)
	arg2 := cb.Get(-1)
	arg1Type, _ := gox.DerefType(arg1.Type)
	typeCast(ctx, arg1Type, arg2)
	cb.AssignWith(1, 1, src)
}

func assignOp(ctx *blockCtx, op token.Token, src ast.Node) {
	cb := ctx.cb
	stk := cb.InternalStack()
	arg1 := stk.Get(-2)
	arg1Type, _ := gox.DerefType(arg1.Type)
	switch op {
	case token.ADD_ASSIGN, token.SUB_ASSIGN: // ptr+=n, ptr-=n
		if t1, ok := arg1Type.(*types.Pointer); ok {
			elemSize := ctx.sizeof(t1.Elem())
			arg2 := stk.Pop()

			cb.UnaryOp(token.AND)
			arg1 = stk.Pop()

			castPtrType(cb, tyUintptrPtr, arg1)
			cb.ElemRef()
			if arg2.Type != tyUintptr {
				cb.Typ(tyUintptr).Val(arg2).Call(1)
			} else {
				stk.Push(arg2)
			}
			if elemSize != 1 {
				cb.Val(elemSize).BinaryOp(token.MUL)
			}
			goto done
		}
		fallthrough
	default:
		arg2 := stk.Get(-1)
		typeCast(ctx, arg1Type, arg2)
	case token.SHL_ASSIGN, token.SHR_ASSIGN:
		// noop
	}
done:
	cb.AssignOp(op, src)
}

func isNegConst(v *gox.Element) bool {
	if cval := v.CVal; cval != nil && cval.Kind() == constant.Int {
		if v, ok := constant.Int64Val(cval); ok {
			return v < 0
		}
	}
	return false
}

func isNilConst(v *gox.Element) bool {
	if cval := v.CVal; cval != nil && cval.Kind() == constant.Int {
		if v, ok := constant.Int64Val(cval); ok {
			return v == 0
		}
	}
	return false
}

func isZeroNumber(v *gox.Element) bool {
	if cval := v.CVal; cval != nil {
		return constant.Sign(cval) == 0
	}
	return false
}

func unaryOp(ctx *blockCtx, op token.Token, v *cast.Node) {
	switch op {
	case token.NOT:
		castToBoolExpr(ctx.cb)
	case token.AND:
		arg := ctx.cb.Get(-1)
		if ctypes.IsFunc(arg.Type) {
			arg.Type = ctypes.NewPointer(arg.Type)
			return
		}
	}
	cb := ctx.cb.UnaryOp(op)
	ret := cb.Get(-1)
	adjustIntConst(ctx, ret, ret.Type)
}

func binaryOp(ctx *blockCtx, op token.Token, v *cast.Node) {
	src := ctx.goNode(v)
	cb := ctx.cb
	stk := cb.InternalStack()
	arg1 := stk.Get(-2)
	arg2 := stk.Get(-1)
	switch op {
	case token.SUB, token.ADD: // ptr-ptr, ptr-n, ptr+n, n+ptr
		if op == token.ADD && isIntegerOrBool(arg1.Type) { // n+ptr
			if _, ok := arg2.Type.(*types.Pointer); ok {
				*arg1, *arg2 = *arg2, *arg1 // => ptr+n
			}
		}
		if t1, ok := arg1.Type.(*types.Pointer); ok {
			elemSize := ctx.sizeof(t1.Elem())
			if isNegConst(arg2) { // fix: can't convert -1 to uintptr
				cb.UnaryOp(token.SUB)
				arg2 = stk.Get(-1)
				adjustIntConst(ctx, arg2, arg2.Type)
				op = (token.SUB + token.ADD) - op
			}
			stk.PopN(2)
			if t2 := arg2.Type; isIntegerOrBool(t2) {
				castPtrType(cb, tyUintptr, arg1)
				if t2 != tyUintptr {
					cb.Typ(tyUintptr).Val(arg2).Call(1)
				} else {
					stk.Push(arg2)
				}
				if elemSize != 1 {
					cb.Val(elemSize).BinaryOp(token.MUL)
				}
				cb.BinaryOp(op, src)
				castPtrType(cb, t1, stk.Pop())
				return
			} else if op == token.SUB && ctypes.Identical(t1, t2) {
				castPtrType(cb, tyUintptr, arg1)
				castPtrType(cb, tyUintptr, arg2)
				cb.BinaryOp(token.SUB, src)
				if elemSize != 1 {
					cb.Val(elemSize).BinaryOp(token.QUO)
				}
				return
			}
			log.Panicln("binaryOp token.ADD/SUB - TODO: unexpected")
		}
	case token.QUO:
		if isZeroNumber(arg2) && arg1.CVal != nil && isFloat(arg1.Type) {
			pkg := ctx.pkg
			ret := pkg.NewParam(token.NoPos, "", arg1.Type)
			arg1.Val = cb.NewClosure(nil, types.NewTuple(ret), false).BodyStart(pkg).
				Val(arg1).Return(1).
				End().Call(0).InternalStack().Pop().Val
			arg1.CVal = nil
		}
	}
	if !isShiftOpertor(op) {
		isUnt1 := isUntyped(arg1.Type)
		isUnt2 := isUntyped(arg2.Type)
		if isUnt1 && !isUnt2 {
			adjustIntConst(ctx, arg1, arg2.Type)
		} else if isUnt2 && !isUnt1 {
			adjustIntConst(ctx, arg2, arg1.Type)
		}
	}
	if isCmpOperator(op) {
		kind1 := checkNilComparable(arg1)
		kind2 := checkNilComparable(arg2)
		if kind1 != 0 || kind2 != 0 { // ptr <cmp> ptr|nil
			isNil1 := isNilConst(arg1)
			isNil2 := isNilConst(arg2)
			if isNil1 || isNil2 { // ptr <cmp> nil
				if isNil1 {
					untypedZeroToNil(arg1)
				}
				if isNil2 {
					untypedZeroToNil(arg2)
				}
				cb.BinaryOp(op, src)
				return
			}
			stk.PopN(2)
			castPtrOrFnPtrType(cb, kind1, arg1.Type, tyUintptr, arg1)
			castPtrOrFnPtrType(cb, kind2, arg2.Type, tyUintptr, arg2)
			cb.BinaryOp(op, src)
			return
		}
	}
	t := toType(ctx, v.Type, 0)
	if isInteger(t) { // bool => int
		args := stk.GetArgs(2)
		if v, ok := gox.CastFromBool(cb, t, args[0]); ok {
			args[0] = v
		}
		if v, ok := gox.CastFromBool(cb, t, args[1]); ok {
			args[1] = v
		}
	}
	cb.BinaryOp(op, src)
	ret := cb.Get(-1)
	adjustIntConst(ctx, ret, ret.Type)
}

func compareOp(ctx *blockCtx, op token.Token, src ast.Node) {
	ctx.cb.BinaryOp(op, src)
}

func untypedZeroToNil(v *gox.Element) {
	v.Type = types.Typ[types.UntypedNil]
	v.Val = &ast.Ident{Name: "nil"}
	v.CVal = nil
}

func castPtrOrFnPtrType(cb *gox.CodeBuilder, kind int, from, to types.Type, v *gox.Element) {
	switch kind {
	case ncKindSignature:
		castFnPtrType(cb, from, to, v)
	default:
		castPtrType(cb, to, v)
	}
}

func stringLit(cb *gox.CodeBuilder, s string, typ types.Type) {
	n := len(s)
	eos := true
	if typ == nil {
		typ = types.NewArray(types.Typ[types.Int8], int64(n+1))
	} else if t, ok := typ.(*types.Array); ok {
		eos = int(t.Len()) > n
	}
	for i := 0; i < n; i++ {
		if c := s[i]; c <= 0x7f {
			cb.Val(rune(c))
		} else {
			cb.Val(int(int8(c)))
		}
	}
	if eos {
		cb.Val(rune(0)).ArrayLit(typ, n+1)
	} else {
		cb.ArrayLit(typ, n)
	}
}

func wstringLit(cb *gox.CodeBuilder, s string, typ types.Type) {
	var n int
	for _, c := range s {
		n++
		cb.Val(c)
	}
	eos := true
	if typ == nil {
		typ = types.NewArray(types.Typ[types.Int32], int64(n+1))
	} else if t, ok := typ.(*types.Array); ok {
		eos = int(t.Len()) > n
	}
	if eos {
		cb.Val(rune(0)).ArrayLit(typ, n+1)
	} else {
		cb.ArrayLit(typ, n)
	}
}

func arrayToElemPtr(cb *gox.CodeBuilder) {
	arr := cb.InternalStack().Pop()
	t, _ := gox.DerefType(arr.Type)
	elem := t.(*types.Array).Elem()
	cb.Typ(ctypes.NewPointer(elem)).Typ(ctypes.UnsafePointer).
		Val(arr).UnaryOp(token.AND).Call(1).Call(1)
}

func castToBoolExpr(cb *gox.CodeBuilder) {
	elem := cb.InternalStack().Get(-1)
	if t := elem.Type; isNumber(t) {
		cb.Val(0).BinaryOp(token.NEQ)
	} else if isNilComparable(t) {
		cb.Val(nil).BinaryOp(token.NEQ)
	}
}

func valOfAddr(cb *gox.CodeBuilder, addr types.Object, ctx *blockCtx) (elemSize int) {
	typ := addr.Type()
	if t, ok := typ.(*types.Pointer); ok {
		typ = t.Elem()
		if t, ok = typ.(*types.Pointer); ok { // **type
			castPtrType(cb, tyUintptrPtr, addr)
			return ctx.sizeof(t.Elem())
		}
	}
	cb.Val(addr)
	return 1
}

func isBasicLit(v *gox.Element) (ok bool) {
	_, ok = v.Val.(*ast.BasicLit)
	return
}

func adjustBigIntConst(ctx *blockCtx, v *gox.Element, t *types.Basic) {
	var bits = 8 * ctx.sizeof(t)
	var mask = (uint64(1) << bits) - 1
	var val = constant.BinaryOp(v.CVal, token.AND, constant.MakeUint64(mask))
	var ival, iok = constant.Uint64Val(val)
	if iok && (t.Info()&types.IsUnsigned) == 0 { // int
		max := (uint64(1) << (bits - 1)) - 1
		if ival > max {
			maskAdd1 := constant.BinaryOp(constant.MakeUint64(mask), token.ADD, constant.MakeInt64(1))
			v.CVal = constant.BinaryOp(val, token.SUB, maskAdd1)
			v.Val = &ast.BasicLit{Kind: token.INT, Value: v.CVal.String()}
			return
		}
	}
	if !isBasicLit(v) || constant.Compare(val, token.NEQ, v.CVal) {
		v.CVal = val
		v.Val = &ast.BasicLit{Kind: token.INT, Value: val.String()}
	}
}

func adjustIntConst(ctx *blockCtx, v *gox.Element, typ types.Type) {
	if e := v.CVal; e == nil || e.Kind() != constant.Int {
		return
	}
	if t, ok := typ.(*types.Basic); ok && isNormalInteger(t) { // integer
		adjustBigIntConst(ctx, v, t)
	}
}

func isNormalInteger(t *types.Basic) bool {
	return t.Kind() <= types.Uintptr && t.Kind() >= types.Int
}

func convertibleTo(V, T types.Type) bool {
	if _, ok := V.(*types.Pointer); ok {
		if _, ok := T.(*types.Pointer); ok {
			return false
		}
		if T == ctypes.UnsafePointer {
			return true
		}
	} else if V == ctypes.UnsafePointer {
		if _, ok := T.(*types.Pointer); ok {
			return true
		}
	}
	return types.ConvertibleTo(V, T)
}

func typeCastCall(ctx *blockCtx, typ types.Type) {
	cb := ctx.cb
	stk := cb.InternalStack()
	v := stk.Get(-1)
	if convertibleTo(v.Type, typ) {
		adjustIntConst(ctx, v, typ)
		cb.Call(1)
		return
	}
	switch vt := v.Type.(type) {
	case *types.Pointer:
		stk.Pop()
		if _, ok := typ.(*types.Pointer); ok || typ == tyUintptr { // ptr => ptr|uintptr
			cb.Typ(ctypes.UnsafePointer).Val(v).Call(1)
		} else { // ptr => int
			castPtrType(cb, tyUintptr, v)
		}
	case *types.Basic:
		switch tt := typ.(type) {
		case *types.Pointer:
			stk.Pop()
			adjustIntConst(ctx, v, tyUintptr)
			cb.Typ(ctypes.UnsafePointer).Typ(tyUintptr).Val(v).Call(1).Call(1)
		case *types.Basic:
			if tt.Kind() == types.UnsafePointer { // int => voidptr
				typeCast(ctx, tyUintptr, v)
				break
			}
			if vt == ctypes.UnsafePointer { // voidptr => int
				stk.Pop()
				cb.Typ(tyUintptr).Val(v).Call(1)
			} else { // int => int
				adjustIntConst(ctx, v, typ)
			}
		case *types.Signature:
			switch vt {
			case types.Typ[types.UntypedInt], ctypes.Int: // untyped_int => fnptr
				vt = tyUintptr
				adjustIntConst(ctx, v, vt)
				fallthrough
			case ctypes.UnsafePointer: // voidptr => fnptr
				stk.PopN(2)
				castFnPtrType(cb, vt, typ, v)
				return
			}
		}
	case *types.Signature: // fnptr => fnptr/voidptr/uintptr
		stk.PopN(2)
		castFnPtrType(cb, vt, typ, v)
		return
	}
	cb.Call(1)
}

func typeCastIndex(ctx *blockCtx, lhs bool) {
	cb := ctx.cb
	v := cb.Get(-2)
	switch v.Type.(type) {
	case *types.Pointer, *types.Basic: // p[n] = *(p+n), n[p] = *(n+p)
		binaryOp(ctx, token.ADD, &cast.Node{})
		if lhs {
			cb.ElemRef()
		} else {
			cb.Elem()
		}
		return
	}
	if lhs {
		cb.IndexRef(1)
	} else {
		cb.Index(1, false)
	}
}

func castFnPtrType(cb *gox.CodeBuilder, from, to types.Type, v *gox.Element) {
	pkg := cb.Pkg()
	fn := types.NewParam(token.NoPos, pkg.Types, "_cgo_fn", from)
	ret := types.NewParam(token.NoPos, pkg.Types, "", to)
	cb.NewClosure(types.NewTuple(fn), types.NewTuple(ret), false).BodyStart(pkg).
		Typ(types.NewPointer(to)).
		Typ(ctypes.UnsafePointer).VarRef(fn).UnaryOp(token.AND).Call(1).
		Call(1).Elem().Return(1).
		End().Val(v).Call(1)
}

func castPtrType(cb *gox.CodeBuilder, typ types.Type, v interface{}) {
	cb.Typ(typ).Typ(ctypes.UnsafePointer).Val(v).Call(1).Call(1)
}

var (
	tyUintptr    = types.Typ[types.Uintptr]
	tyUintptrPtr = types.NewPointer(tyUintptr)
)

// -----------------------------------------------------------------------------
