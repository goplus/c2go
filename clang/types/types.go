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

type Func struct {
	*types.Signature
}

func NewFunc(sig *types.Signature) types.Type {
	return Func{sig}
}

func NewPointer(typ types.Type) types.Type {
	switch t := typ.(type) {
	case *types.Basic:
		if t == Void {
			return types.Typ[types.UnsafePointer]
		}
	case Func:
		return t.Signature
	}
	return types.NewPointer(typ)
}

func IsFunc(typ types.Type) bool {
	_, ok := typ.(Func)
	return ok
}

// -----------------------------------------------------------------------------
