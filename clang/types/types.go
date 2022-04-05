package types

import (
	"go/types"
	"unsafe"
)

// -----------------------------------------------------------------------------

var (
	Void          = types.Typ[types.UntypedNil]
	UnsafePointer = types.Typ[types.UnsafePointer]

	Long    = types.Typ[types.Int64]
	Ulong   = types.Typ[types.Uint64]
	NotImpl = UnsafePointer

	LongDouble = types.Typ[types.Float64]

	Int128  = NotImpl
	Uint128 = NotImpl
)

func init() { // TODO: how to support cross-compiling?
	if unsafe.Sizeof(uintptr(0)) == 4 {
		Long = types.Typ[types.Int32]
		Ulong = types.Typ[types.Uint32]
	}
}

func NotVoid(t types.Type) bool {
	return t != Void
}

// -----------------------------------------------------------------------------
