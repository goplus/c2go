//go:build windows_msvc
// +build windows_msvc

package main

import unsafe "unsafe"

type struct___crt_locale_data struct {
}
type struct___crt_multibyte_data struct {
}

func __acrt_iob_func(_Ix uint32) *struct__iobuf {
	return nil
}
func __stdio_common_vfprintf(_Options uint64, _Stream *struct__iobuf, _Format *int8, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	return goprintf(_Format, _ArgList...)
}
func __stdio_common_vfprintf_p(_Options uint64, _Stream *struct__iobuf, _Format *int8, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vfprintf_s(_Options uint64, _Stream *struct__iobuf, _Format *int8, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vfscanf(_Options uint64, _Stream *struct__iobuf, _Format *int8, _Locale *struct___crt_locale_pointers, _Arglist []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vfwprintf(_Options uint64, _Stream *struct__iobuf, _Format *uint16, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vfwprintf_p(_Options uint64, _Stream *struct__iobuf, _Format *uint16, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vfwprintf_s(_Options uint64, _Stream *struct__iobuf, _Format *uint16, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vfwscanf(_Options uint64, _Stream *struct__iobuf, _Format *uint16, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vsnprintf_s(_Options uint64, _Buffer *int8, _BufferCount uint64, _MaxCount uint64, _Format *int8, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vsnwprintf_s(_Options uint64, _Buffer *uint16, _BufferCount uint64, _MaxCount uint64, _Format *uint16, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vsprintf(_Options uint64, _Buffer *int8, _BufferCount uint64, _Format *int8, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vsprintf_p(_Options uint64, _Buffer *int8, _BufferCount uint64, _Format *int8, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vsprintf_s(_Options uint64, _Buffer *int8, _BufferCount uint64, _Format *int8, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vsscanf(_Options uint64, _Buffer *int8, _BufferCount uint64, _Format *int8, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vswprintf(_Options uint64, _Buffer *uint16, _BufferCount uint64, _Format *uint16, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vswprintf_p(_Options uint64, _Buffer *uint16, _BufferCount uint64, _Format *uint16, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vswprintf_s(_Options uint64, _Buffer *uint16, _BufferCount uint64, _Format *uint16, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __stdio_common_vswscanf(_Options uint64, _Buffer *uint16, _BufferCount uint64, _Format *uint16, _Locale *struct___crt_locale_pointers, _ArgList []interface {
}) int32 {
	panic("notimpl")
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
