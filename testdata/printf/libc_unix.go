//go:build !windows
// +build !windows

package main

func vfprintf(fp *FILE, format *int8, args []interface{}) int32 {
	return goprintf(format, args...)
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
