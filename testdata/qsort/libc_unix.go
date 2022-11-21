//go:build !windows
// +build !windows

package main

import (
	"log"
	"unsafe"

	c "github.com/goplus/c2go/clang"
)

func printf(format *int8, args ...interface{}) int32 {
	return goprintf(format, args...)
}

func __swbuf(_c int32, _p *FILE) int32 {
	return _c
}

type struct___sFILEX struct{}

type struct__IO_marker struct{} // Linux
type struct__IO_codecvt struct{}
type struct__IO_wide_data struct{}

func __builtin___memcpy_chk(dst unsafe.Pointer, src unsafe.Pointer, n c.SizeT, elem c.SizeT) unsafe.Pointer {
	copy(sliceOf(dst, n), sliceOf(src, n))
	return dst
}

func __builtin_object_size(unsafe.Pointer, int32) c.SizeT {
	return 1
}

func __builtin_bswap32(v uint32) uint32 {
	log.Panicln("__builtin_bswap32: notimpl")
	return v
}

func __builtin_bswap64(v uint64) uint64 {
	log.Panicln("__builtin_bswap32: notimpl")
	return v
}
