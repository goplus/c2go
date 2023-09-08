//go:build windows_gnu
// +build windows_gnu

package main

import unsafe "unsafe"

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
func free(_Memory unsafe.Pointer) {
	panic("notimpl")
}
func malloc(_Size uint64) unsafe.Pointer {
	panic("notimpl")
}
func strnlen(_Str *int8, _MaxCount uint64) uint64 {
	panic("notimpl")
}
func wcsnlen(_Src *uint16, _MaxCount uint64) uint64 {
	panic("notimpl")
}
