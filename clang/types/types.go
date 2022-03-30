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
	NotImpl = types.Typ[types.UnsafePointer]

	Int128  = NotImpl
	Uint128 = NotImpl

	Void  = types.Typ[types.UntypedNil]
	Long  = types.Typ[types.Int64]
	Ulong = types.Typ[types.Uint64]
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
