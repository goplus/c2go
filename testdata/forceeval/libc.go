package main

import (
	"fmt"
	"strings"
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
	s := strings.ToLower(fmt.Sprintf(goformat, args...))
	fmt.Print(s)
	return 0
}

func __swbuf(_c int32, _p *FILE) int32 {
	return _c
}

type struct___sFILEX struct{}

type struct__IO_marker struct{}
type struct__IO_codecvt struct{}
type struct__IO_wide_data struct{}
