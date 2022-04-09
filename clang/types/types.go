package types

import (
	"go/types"
)

// -----------------------------------------------------------------------------

var (
	Void          = types.Typ[types.UntypedNil]
	UnsafePointer = types.Typ[types.UnsafePointer]

	Int     = types.Typ[types.Int32]
	Uint    = types.Typ[types.Uint32]
	Long    = types.Typ[types.Int]
	Ulong   = types.Typ[types.Uint]
	NotImpl = UnsafePointer

	Enum = types.Typ[types.Int32]

	LongDouble = types.Typ[types.Float64]

	Int128  = NotImpl
	Uint128 = NotImpl
)

func NotVoid(t types.Type) bool {
	return t != Void
}

func MangledName(tag, name string) string {
	return tag + "_" + name // TODO: use sth to replace _
}

// -----------------------------------------------------------------------------
