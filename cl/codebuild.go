package cl

import (
	"go/ast"
	"go/token"
	"go/types"
	"log"

	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

func binaryOp(ctx *blockCtx, op token.Token, src ast.Node) {
	switch op {
	case token.SUB, token.ADD: // ptr-ptr, ptr-n, ptr+n
		cb := ctx.cb
		stk := cb.InternalStack()
		args := stk.GetArgs(2)
		if t, ok := args[0].Type.(*types.Pointer); ok {
			stk.PopN(2)
			elemSize := ctx.sizeof(t.Elem())
			t2 := args[1].Type
			if isInteger(t2) {
				castPtrType(cb, tyUintptr, args[0])
				if t2 != tyUintptr {
					castPtrType(cb, tyUintptr, args[1])
				} else {
					stk.Push(args[1])
				}
				if elemSize != 1 {
					cb.Val(elemSize).BinaryOp(token.MUL)
				}
				cb.BinaryOp(op, src)
				castPtrType(cb, t, stk.Pop())
				return
			} else if op == token.SUB && types.Identical(t, t2) {
				castPtrType(cb, tyUintptr, args[0])
				castPtrType(cb, tyUintptr, args[1])
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
	cb.Typ(types.NewPointer(elem)).Typ(types.Typ[types.UnsafePointer]).
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

func castPtrType(cb *gox.CodeBuilder, typ types.Type, v interface{}) {
	cb.Typ(typ).Typ(types.Typ[types.UnsafePointer]).Val(v).Call(1).Call(1)
}

var (
	tyUintptr    = types.Typ[types.Uintptr]
	tyUintptrPtr = types.NewPointer(tyUintptr)
)

// -----------------------------------------------------------------------------
