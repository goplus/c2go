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
	"__builtin___memset_chk": "void* (void*, int32, uint, uint)",
	"__builtin___memcpy_chk": "void* (void*, void*, uint, uint)",
	"__builtin_object_size": "uint (void*, int32)"
}`

func decl_builtin(ctx *blockCtx) {
	var fns map[string]string
	err := json.NewDecoder(strings.NewReader(builtin_decls)).Decode(&fns)
	if err != nil {
		log.Panicln("decl_builtin decode error:", err)
	}
	pkg := ctx.pkg.Types
	scope := pkg.Scope()
	for fn, proto := range fns {
		t := toType(ctx, &cast.Type{QualType: proto}, 0)
		scope.Insert(types.NewFunc(token.NoPos, pkg, fn, t.(ctypes.Func).Signature))
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

func typeCast(cb *gox.CodeBuilder, typ types.Type, arg *gox.Element) {
	if !ctypes.Identical(typ, arg.Type) {
		*arg = *cb.Typ(typ).Val(arg).Call(1).InternalStack().Pop()
	}
}

func assign(ctx *blockCtx, src ast.Node) {
	cb := ctx.cb
	arg1 := cb.Get(-2)
	arg2 := cb.Get(-1)
	arg1Type, _ := gox.DerefType(arg1.Type)
	typeCast(cb, arg1Type, arg2)
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
		typeCast(cb, arg1Type, arg2)
	case token.SHL_ASSIGN, token.SHR_ASSIGN:
		// noop
	}
done:
	cb.AssignOp(op, src)
}

func binaryOp(ctx *blockCtx, op token.Token, v *cast.Node) {
	src := goNode(v)
	cb := ctx.cb
	stk := cb.InternalStack()
	switch op {
	case token.SUB, token.ADD: // ptr-ptr, ptr-n, ptr+n
		arg1 := stk.Get(-2)
		if t1, ok := arg1.Type.(*types.Pointer); ok {
			elemSize := ctx.sizeof(t1.Elem())
			arg2 := stk.Get(-1)
			stk.PopN(2)
			if t2 := arg2.Type; isInteger(t2) {
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
					cb.Val(elemSize).BinaryOp(token.MUL)
				}
				return
			}
			log.Panicln("binaryOp token.SUB - TODO: unexpected")
		}
	}
	if isCmpOperator(op) {
		arg1 := stk.Get(-2)
		if isNilComparable(arg1.Type) { // ptr <cmp> ptr
			arg2 := stk.Get(-1)
			stk.PopN(2)
			castPtrType(cb, tyUintptr, arg1)
			castPtrType(cb, tyUintptr, arg2)
			cb.BinaryOp(op, src)
			return
		}
	}
	t := toType(ctx, v.Type, 0)
	if isInteger(t) { // bool => int
		args := stk.GetArgs(2)
		if isBool(args[0].Type) {
			args[0] = castFromBoolExpr(ctx, t, args[0])
		}
		if isBool(args[1].Type) {
			args[1] = castFromBoolExpr(ctx, t, args[1])
		}
	}
	cb.BinaryOp(op, src)
}

func stringLit(cb *gox.CodeBuilder, s string, typ types.Type) {
	n := len(s)
	if typ == nil {
		typ = types.NewArray(types.Typ[types.Int8], int64(n+1))
	}
	for i := 0; i < n; i++ {
		cb.Val(rune(s[i]))
	}
	cb.Val(rune(0)).ArrayLit(typ, n+1)
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
	if t := elem.Type; isInteger(t) {
		cb.Val(0).BinaryOp(token.NEQ)
	} else if isNilComparable(t) {
		cb.Val(nil).BinaryOp(token.NEQ)
	}
}

func castFromBoolExpr(ctx *blockCtx, typ types.Type, v *gox.Element) *gox.Element {
	pkg := ctx.pkg
	results := types.NewTuple(types.NewParam(token.NoPos, pkg.Types, "", typ))
	return ctx.cb.NewClosure(nil, results, false).BodyStart(pkg).
		If().Val(v).Then().Val(1).Return(1).
		Else().Val(0).Return(1).
		End().
		End().Call(0).
		InternalStack().Pop()
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

func negConst2Uint(ctx *blockCtx, v *gox.Element, typ types.Type) {
	if v.CVal == nil {
		return
	}
	if val, ok := constant.Val(v.CVal).(int64); ok && val < 0 {
		nval := (uint64(1) << (8 * ctx.sizeof(typ))) + uint64(val)
		v.Val = &ast.BasicLit{Kind: token.INT, Value: strconv.FormatUint(nval, 10)}
		v.CVal = constant.MakeUint64(nval)
	}
}

func typeCastCall(ctx *blockCtx, typ types.Type) {
	cb := ctx.cb
	stk := cb.InternalStack()
	v := stk.Get(-1)
	switch vt := v.Type.(type) {
	case *types.Pointer:
		if typ == ctypes.UnsafePointer { // ptr => voidptr
			break
		}
		stk.Pop()
		if _, ok := typ.(*types.Pointer); ok || typ == tyUintptr { // ptr => ptr|uintptr
			cb.Typ(ctypes.UnsafePointer).Val(v).Call(1)
		} else { // ptr => int
			castPtrType(cb, tyUintptr, v)
		}
	case *types.Basic:
		switch tt := typ.(type) {
		case *types.Pointer:
			if vt == ctypes.UnsafePointer { // voidptr => ptr
				break
			}
			stk.Pop()
			negConst2Uint(ctx, v, tyUintptr)
			// int => ptr
			cb.Typ(ctypes.UnsafePointer).Typ(tyUintptr).Val(v).Call(1).Call(1)
		case *types.Basic: // int => int
			if (tt.Info() & types.IsUnsigned) != 0 {
				negConst2Uint(ctx, v, typ)
			}
		}
	}
	cb.Call(1)
}

func typeCastIndex(ctx *blockCtx, lhs bool) {
	cb := ctx.cb
	stk := cb.InternalStack()
	v := stk.Get(-2)
	switch t := v.Type.(type) {
	case *types.Pointer: // *T => *[N]T
		arrt := arrayPtrOf(t.Elem())
		idx := stk.Get(-1)
		stk.PopN(2)
		castPtrType(cb, arrt, v)
		stk.Push(idx)
	}
	if lhs {
		cb.IndexRef(1)
	} else {
		cb.Index(1, false)
	}
}

func arrayPtrOf(elem types.Type) types.Type {
	const (
		arrayLen = 1 << 20 // TODO: may need change
	)
	return types.NewPointer(types.NewArray(elem, arrayLen))
}

func castPtrType(cb *gox.CodeBuilder, typ types.Type, v interface{}) {
	cb.Typ(typ).Typ(ctypes.UnsafePointer).Val(v).Call(1).Call(1)
}

var (
	tyUintptr    = types.Typ[types.Uintptr]
	tyUintptrPtr = types.NewPointer(tyUintptr)
)

// -----------------------------------------------------------------------------
