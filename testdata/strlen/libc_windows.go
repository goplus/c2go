package main

import (
	"fmt"
	"unsafe"
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

func strnlen(_Str *int8, _MaxCount uint64) uint64 {
	panic("notimpl")
}
func wcsnlen(_Src *uint16, _MaxCount uint64) uint64 {
	panic("notimpl")
}
