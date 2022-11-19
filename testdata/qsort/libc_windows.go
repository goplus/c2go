package main

import (
	"fmt"
	unsafe "unsafe"

	c "github.com/goplus/c2go/clang"
)

func C(s string) *int8 {
	n := len(s)
	ret := make([]byte, n+1)
	copy(ret, s)
	ret[n] = '\x00'
	return (*int8)(unsafe.Pointer(&ret[0]))
}

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

type struct___lc_time_data struct {
}
type struct_lconv struct {
}
type struct_threadmbcinfostruct struct {
}

func __builtin_llabs(int64) int64 {
	panic("notimpl")
}
func __mingw_aligned_free(_Memory unsafe.Pointer) {
	panic("notimpl")
}
func __mingw_aligned_malloc(_Size uint64, _Alignment uint64) unsafe.Pointer {
	panic("notimpl")
}
func __mingw_strtod(*int8, **int8) float64 {
	panic("notimpl")
}
func __mingw_strtof(*int8, **int8) float32 {
	panic("notimpl")
}
func __mingw_wcstod(_Str *uint16, _EndPtr **uint16) float64 {
	panic("notimpl")
}
func __mingw_wcstof(nptr *uint16, endptr **uint16) float32 {
	panic("notimpl")
}
func a_cas(p *int32, t int32, s int32) int32 {
	panic("notimpl")
}
func free(_Memory unsafe.Pointer) {
	panic("notimpl")
}
func malloc(_Size uint64) unsafe.Pointer {
	panic("notimpl")
}

func sliceOf(v unsafe.Pointer, bytes c.SizeT) []byte {
	return (*[1 << 20]byte)(v)[:bytes]
}

func memcpy(dst unsafe.Pointer, src unsafe.Pointer, n c.SizeT) unsafe.Pointer {
	copy(sliceOf(dst, n), sliceOf(src, n))
	return dst
}

func strnlen(_Str *int8, _MaxCount uint64) uint64 {
	panic("notimpl")
}
func wcsnlen(_Src *uint16, _MaxCount uint64) uint64 {
	panic("notimpl")
}
