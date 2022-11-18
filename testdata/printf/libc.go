//go:build !windows
// +build !windows

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

func vfprintf(fp *FILE, format *int8, args []interface{}) int32 {
	goformat := gostring(format)
	for i, arg := range args {
		if v, ok := arg.(*int8); ok {
			args[i] = gostring(v)
		}
	}
	fmt.Printf(goformat, args...)
	return 0
}

func __swbuf(_c int32, _p *FILE) int32 {
	return _c
}

type struct___sFILEX struct{}

type struct__IO_marker struct{}
type struct__IO_codecvt struct{}
type struct__IO_wide_data struct{}

var (
	stdout    *FILE
	__stdoutp *FILE
)
