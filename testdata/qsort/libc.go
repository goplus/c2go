package main

import (
	"fmt"
	"strings"
	"unsafe"

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

func goprintf(format *int8, args ...interface{}) int32 {
	goformat := strings.ReplaceAll(gostring(format), "%lld", "%d")
	for i, arg := range args {
		if v, ok := arg.(*int8); ok {
			args[i] = gostring(v)
		}
	}
	fmt.Printf(goformat, args...)
	return 0
}

func sliceOf(v unsafe.Pointer, bytes c.SizeT) []byte {
	return (*[1 << 20]byte)(v)[:bytes]
}

func memcpy(dst unsafe.Pointer, src unsafe.Pointer, n c.SizeT) unsafe.Pointer {
	copy(sliceOf(dst, n), sliceOf(src, n))
	return dst
}

func a_cas(p *int32, t int32, s int32) int32 {
	panic("notimpl")
}
