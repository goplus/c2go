//go:build windows_gnu
// +build windows_gnu

package main

type struct___lc_time_data struct {
}
type struct_lconv struct {
}
type struct_threadmbcinfostruct struct {
}

func __acrt_iob_func(index uint32) *struct__iobuf {
	return nil
}
func __mingw_vfprintf(f *struct__iobuf, format *int8, args []interface {
}) int32 {
	return goprintf(format, args...)
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
func strnlen(_Str *int8, _MaxCount uint64) uint64 {
	panic("notimpl")
}
func wcsnlen(_Src *uint16, _MaxCount uint64) uint64 {
	panic("notimpl")
}
