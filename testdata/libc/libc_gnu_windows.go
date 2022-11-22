package libc

import (
	"fmt"
	unsafe "unsafe"
)

func a_cas(p *int32, t, s int32) int32 {
	panic("notimpl")
}

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

type struct___lc_time_data struct {
}
type struct_lconv struct {
}
type struct_threadmbcinfostruct struct {
}

func __acrt_iob_func(index uint32) *struct__iobuf {
	return nil
}
func X__builtin_llabs(int64) int64 {
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
func __mingw_vfprintf(f *struct__iobuf, format *int8, args []interface {
}) int32 {
	goformat := gostring(format)
	for i, arg := range args {
		if v, ok := arg.(*int8); ok {
			args[i] = gostring(v)
		}
	}
	fmt.Printf(goformat, args...)
	return 0
}
func __mingw_vfscanf(fp *struct__iobuf, Format *int8, argp []interface {
}) int32 {
	panic("notimpl")
}
func __mingw_vfwprintf(_File *struct__iobuf, _Format *uint16, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __mingw_vfwscanf(fp *struct__iobuf, Format *uint16, argp []interface {
}) int32 {
	panic("notimpl")
}
func __mingw_vsnprintf(_DstBuf *int8, _MaxCount uint64, _Format *int8, _ArgList []interface {
}) int32 {
	panic("notimpl")
}
func __mingw_vsnwprintf(*uint16, uint64, *uint16, []interface {
}) int32 {
	panic("notimpl")
}
func __mingw_vsprintf(*int8, *int8, []interface {
}) int32 {
	panic("notimpl")
}
func __mingw_vsscanf(_Str *int8, Format *int8, argp []interface {
}) int32 {
	panic("notimpl")
}
func __mingw_vswscanf(_Str *uint16, Format *uint16, argp []interface {
}) int32 {
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
func memcpy(_Dst unsafe.Pointer, _Src unsafe.Pointer, _Size uint64) unsafe.Pointer {
	panic("notimpl")
}
func strnlen(_Str *int8, _MaxCount uint64) uint64 {
	panic("notimpl")
}
func wcsnlen(_Src *uint16, _MaxCount uint64) uint64 {
	panic("notimpl")
}
