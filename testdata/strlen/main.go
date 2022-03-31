package main

func main() {
	cstr := C("Hello, world")
	printf(C("%s, len: %d\n"), cstr, strlen(cstr))
}
