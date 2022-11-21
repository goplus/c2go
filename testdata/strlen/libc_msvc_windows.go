//go:build windows_msvc
// +build windows_msvc

package main

import unsafe "unsafe"

type struct___crt_locale_data struct {
}
type struct___crt_multibyte_data struct {
}

func _errno() *int32 {
	panic("notimpl")
}
func _invalid_parameter_noinfo() {
	panic("notimpl")
}
func memcpy(_Dst unsafe.Pointer, _Src unsafe.Pointer, _Size uint64) unsafe.Pointer {
	panic("notimpl")
}
func memmove(_Dst unsafe.Pointer, _Src unsafe.Pointer, _Size uint64) unsafe.Pointer {
	panic("notimpl")
}
func memset(_Dst unsafe.Pointer, _Val int32, _Size uint64) unsafe.Pointer {
	panic("notimpl")
}
func strnlen(_String *int8, _MaxCount uint64) uint64 {
	panic("notimpl")
}
func wcsnlen(_Source *uint16, _MaxCount uint64) uint64 {
	panic("notimpl")
}
func wcstok(_String *uint16, _Delimiter *uint16, _Context **uint16) *uint16 {
	panic("notimpl")
}
