package types

import (
	"go/token"
	"go/types"
	"unsafe"

	"github.com/goplus/gox"
)

// -----------------------------------------------------------------------------

var (
	Void          = types.Typ[types.UntypedNil]
	UnsafePointer = types.Typ[types.UnsafePointer]

	Int     = types.Typ[types.Int32]
	Uint    = types.Typ[types.Uint32]
	Long    = types.Typ[uintptr(types.Int32)+unsafe.Sizeof(0)>>3]  // int32/int64
	Ulong   = types.Typ[uintptr(types.Uint32)+unsafe.Sizeof(0)>>3] // uint32/uint64
	NotImpl = UnsafePointer
	Enum    = types.Typ[types.Int32]

	LongDouble = types.Typ[types.Float64]
)

func NotVoid(t types.Type) bool {
	return t != Void
}

func MangledName(tag, name string) string {
	return tag + "_" + name // TODO: use sth to replace _
}

// -----------------------------------------------------------------------------

var (
	ValistTag types.Type
	Valist    types.Type = types.NewSlice(gox.TyEmptyInterface)
)

func init() {
	vaTag := types.NewTypeName(token.NoPos, types.Unsafe, MangledName("struct", "__va_list_tag"), nil)
	ValistTag = types.NewNamed(vaTag, types.NewStruct(nil, nil), nil)
	types.Universe.Insert(vaTag)
}

// -----------------------------------------------------------------------------

func NewFunc(params, results *types.Tuple, variadic bool) *types.Signature {
	return gox.NewCSignature(params, results, variadic)
}

func NewPointer(typ types.Type) types.Type {
	switch t := typ.(type) {
	case *types.Basic:
		if t == Void {
			return types.Typ[types.UnsafePointer]
		}
	case *types.Signature:
		if gox.IsCSignature(t) {
			return types.NewSignature(nil, t.Params(), t.Results(), t.Variadic())
		}
	case *types.Named:
		if typ == ValistTag {
			return Valist
		}
	}
	return types.NewPointer(typ)
}

func IsFunc(typ types.Type) bool {
	sig, ok := typ.(*types.Signature)
	if ok {
		ok = gox.IsCSignature(sig)
	}
	return ok
}

func Identical(typ1, typ2 types.Type) bool {
	return types.Identical(typ1, typ2)
}

// -----------------------------------------------------------------------------
