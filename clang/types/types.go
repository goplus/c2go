package types

import (
	"errors"
	"go/types"
	"unsafe"
)

var (
	ErrNotFound = errors.New("type not found")
)

// -----------------------------------------------------------------------------

var (
	Void          = types.Typ[types.UntypedNil]
	UnsafePointer = types.Typ[types.UnsafePointer]

	Long    = types.Typ[types.Int64]
	Ulong   = types.Typ[types.Uint64]
	NotImpl = UnsafePointer

	LongDouble = NotImpl

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
