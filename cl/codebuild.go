package cl

import (
	"go/token"
	"go/types"

	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

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
	cb.Typ(types.NewPointer(elem)).Typ(types.Typ[types.UnsafePointer])
	cb.InternalStack().Push(arr)
	cb.UnaryOp(token.AND).Call(1).Call(1)
}

func valOfAddr(cb *gox.CodeBuilder, addr types.Object) (elemSize int) {
	typ := addr.Type()
	if t, ok := typ.(*types.Pointer); ok {
		typ = t.Elem()
		if t, ok = typ.(*types.Pointer); ok { // **type
			cb.Typ(tyUintptrPtr).Typ(types.Typ[types.UnsafePointer]).Val(addr).Call(1).Call(1)
			return sizeof(t.Elem())
		}
	}
	cb.Val(addr)
	return 1
}

var (
	tyUintptrPtr = types.NewPointer(types.Typ[types.Uintptr])
)

// -----------------------------------------------------------------------------
