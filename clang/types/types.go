package types

import (
	"go/types"
)

// -----------------------------------------------------------------------------

var (
	Void          = types.Typ[types.UntypedNil]
	UnsafePointer = types.Typ[types.UnsafePointer]

	Long    = types.Typ[types.Int]
	Ulong   = types.Typ[types.Uint]
	NotImpl = UnsafePointer

	LongDouble = types.Typ[types.Float64]

	Int128  = NotImpl
	Uint128 = NotImpl
)

func NotVoid(t types.Type) bool {
	return t != Void
}

// -----------------------------------------------------------------------------
