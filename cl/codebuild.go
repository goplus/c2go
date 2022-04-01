package cl

import (
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"log"
	"strconv"

	"github.com/goplus/gox"

	ctypes "github.com/goplus/c2go/clang/types"
)

// -----------------------------------------------------------------------------

func binaryOp(ctx *blockCtx, op token.Token, src ast.Node) {
	switch op {
	case token.SUB, token.ADD: // ptr-ptr, ptr-n, ptr+n
		cb := ctx.cb
		stk := cb.InternalStack()
		arg0 := stk.Get(-2)
		if t, ok := arg0.Type.(*types.Pointer); ok {
			elemSize := ctx.sizeof(t.Elem())
			arg1 := stk.Get(-1)
			stk.PopN(2)
			if t2 := arg1.Type; isInteger(t2) {
				castPtrType(cb, tyUintptr, arg0)
				if t2 != tyUintptr {
					castPtrType(cb, tyUintptr, arg1)
				} else {
					stk.Push(arg1)
				}
				if elemSize != 1 {
					cb.Val(elemSize).BinaryOp(token.MUL)
				}
				cb.BinaryOp(op, src)
				castPtrType(cb, t, stk.Pop())
				return
			} else if op == token.SUB && types.Identical(t, t2) {
				castPtrType(cb, tyUintptr, arg0)
				castPtrType(cb, tyUintptr, arg1)
				cb.BinaryOp(token.SUB, src)
				if elemSize != 1 {
					cb.Val(elemSize).BinaryOp(token.MUL)
				}
				return
			}
			log.Fatalln("binaryOp token.SUB - TODO: unexpected")
		}
	}
	ctx.cb.BinaryOp(op, src)
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
	cb.Typ(types.NewPointer(elem)).Typ(ctypes.UnsafePointer).
		Val(arr).UnaryOp(token.AND).Call(1).Call(1)
}

func castToBoolExpr(cb *gox.CodeBuilder) {
	elem := cb.InternalStack().Get(-1)
	if isInteger(elem.Type) {
		cb.Val(0).BinaryOp(token.NEQ)
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
	switch v.Type.(type) {
	case *types.Pointer:
		stk.Pop()
		if _, ok := typ.(*types.Pointer); ok || typ == tyUintptr { // ptr => ptr|uintptr
			cb.Typ(ctypes.UnsafePointer).Val(v).Call(1)
		} else { // ptr => int
			castPtrType(cb, tyUintptr, v)
		}
	case *types.Basic:
		switch tt := typ.(type) {
		case *types.Pointer: // int => ptr
			stk.Pop()
			negConst2Uint(ctx, v, tyUintptr)
			cb.Typ(ctypes.UnsafePointer).Typ(tyUintptr).Val(v).Call(1).Call(1)
		case *types.Basic: // int => int
			if (tt.Info() & types.IsUnsigned) != 0 {
				negConst2Uint(ctx, v, typ)
			}
		}
	}
	cb.Call(1)
}

func castPtrType(cb *gox.CodeBuilder, typ types.Type, v interface{}) {
	cb.Typ(typ).Typ(ctypes.UnsafePointer).Val(v).Call(1).Call(1)
}

var (
	tyInt        = types.Typ[types.Int]
	tyUintptr    = types.Typ[types.Uintptr]
	tyUintptrPtr = types.NewPointer(tyUintptr)
)

// -----------------------------------------------------------------------------
