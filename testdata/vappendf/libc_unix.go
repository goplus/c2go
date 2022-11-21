//go:build !windows
// +build !windows

package main

import (
	"unsafe"

	c "github.com/goplus/c2go/clang"
)

func strlen(s *int8) int32 {
	n, arr := int32(0), (*[1 << 20]byte)(unsafe.Pointer(s))
	for arr[n] != 0 {
		n++
	}
	return n
}

func sliceOf(v unsafe.Pointer, bytes c.SizeT) []byte {
	return (*[1 << 20]byte)(v)[:bytes]
}

func memcpy(dst unsafe.Pointer, src unsafe.Pointer, n c.SizeT) unsafe.Pointer {
	copy(sliceOf(dst, n), sliceOf(src, n))
	return dst
}

func __builtin___memcpy_chk(dst unsafe.Pointer, src unsafe.Pointer, n c.SizeT, elem c.SizeT) unsafe.Pointer {
	copy(sliceOf(dst, n), sliceOf(src, n))
	return dst
}

func __builtin_object_size(unsafe.Pointer, int32) c.SizeT {
	return 1
}

type struct___locale_data struct{}
