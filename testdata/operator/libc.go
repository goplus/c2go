package main

import (
	"fmt"
	"unsafe"
)

func gostring(s *int8) string {
	n, arr := 0, (*[1 << 20]byte)(unsafe.Pointer(s))
	for arr[n] != 0 {
		n++
	}
	return string(arr[:n])
}

func printf(format *int8, args ...interface{}) int32 {
	goformat := gostring(format)
	for i, arg := range args {
		if v, ok := arg.(*int8); ok {
			args[i] = gostring(v)
		}
	}
	fmt.Printf(goformat, args...)
	return 0
}

func sliceOf(v unsafe.Pointer, bytes uint) []byte {
	return (*[1 << 20]byte)(v)[:bytes]
}

func memcpy(dst unsafe.Pointer, src unsafe.Pointer, n uint) unsafe.Pointer {
	copy(sliceOf(dst, n), sliceOf(src, n))
	return dst
}

func __builtin___memcpy_chk(dst unsafe.Pointer, src unsafe.Pointer, n uint, elem uint) unsafe.Pointer {
	copy(sliceOf(dst, n), sliceOf(src, n))
	return dst
}

func __builtin_object_size(unsafe.Pointer, int32) uint {
	return 0
}

func __swbuf(_c int32, _p *FILE) int32 {
	return _c
}

type struct___sFILEX struct{}

type struct__IO_marker struct{}
type struct__IO_codecvt struct{}
type struct__IO_wide_data struct{}

type struct___locale_data struct{}
