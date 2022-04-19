package main

func __swbuf_r(_ptr *struct__reent, _c int32, _p *FILE) int32 {
	return _c
}

func __srget_r(_ptr *struct__reent, _p *FILE) int32 {
	return 0
}

func __getreent() *struct__reent {
	return nil
}

func ungetc(_c int32, _p *FILE) {
}

type struct___locale_t struct{} // Windows
