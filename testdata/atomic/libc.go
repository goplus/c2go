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
	goformat := strings.ReplaceAll(gostring(format), "%lld", "%d")
	for i, arg := range args {
		if v, ok := arg.(*int8); ok {
			args[i] = gostring(v)
		}
	}
	fmt.Printf(goformat, args...)
	return 0
}

func __atomic_store_n_i32(p *int32, memorder int32, v int32) {
	*p = v
}

func __atomic_store_n_i64(p *int64, memorder int32, v int64) {
	*p = v
}

func __atomic_load_n_i32(p *int32, memorder int32) int32 {
	return *p
}

func __atomic_load_n_i64(p *int64, memorder int32) int64 {
	return *p
}

func __swbuf(_c int32, _p *FILE) int32 {
	return _c
}

type struct___sFILEX struct{}

type struct__IO_marker struct{} // Linux
type struct__IO_codecvt struct{}
type struct__IO_wide_data struct{}
