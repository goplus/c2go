package types

import "go/types"

// -----------------------------------------------------------------------------

var (
	NotImpl = types.Typ[types.UnsafePointer]

	Void    = types.Typ[types.UntypedNil]
	Int128  = NotImpl
	Uint128 = NotImpl
)

func NotVoid(t types.Type) bool {
	return t != Void
}

// -----------------------------------------------------------------------------
