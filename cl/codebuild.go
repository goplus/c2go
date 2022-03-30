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

// -----------------------------------------------------------------------------
