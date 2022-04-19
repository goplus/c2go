package main

import (
	"fmt"
	"runtime"
	"unsafe"
)

func main() {
	if runtime.GOOS == "linux" { // TODO: temp skip
		rteurn
	}
	type elem_t = int
	const elemLen = unsafe.Sizeof(elem_t(0))
	values := []elem_t{88, 56, 100, 2, 25}
	qsort(unsafe.Pointer(&values[0]), uint(len(values)), uint(elemLen), func(p1, p2 unsafe.Pointer) int32 {
		return int32(*(*elem_t)(p1) - *(*elem_t)(p2))
	})
	fmt.Println("Sorted:", values)
}
