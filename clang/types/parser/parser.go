package parser

import (
	"go/token"
	"go/types"
	"io"
	"log"
	"strconv"

	"github.com/goplus/c2go/clang/types/scanner"
)

// -----------------------------------------------------------------------------

type TypeSystem interface {
	LookupType(typ string, unsigned bool) (t types.Type, err error)
}

// qualType can be:
//   unsigned int
//   struct ConstantString
//   volatile uint32_t
//   int (*)(void *, int, char **, char **)
//   int (*)(const char *, ...)
//   int (*)(void)
//   const char *restrict
//   const char [7]
//   char *
//   void
//   ...
func ParseType(ts TypeSystem, fset *token.FileSet, qualType string, isParam bool) (t types.Type, err error) {
	var s scanner.Scanner
	file := fset.AddFile("", fset.Base(), len(qualType))
	s.Init(file, qualType, nil)
	unsigned := false
	for {
		pos, tok, lit := s.Scan()
		switch tok {
		case token.IDENT:
			switch lit {
			case "unsigned":
				unsigned = true
			case "const", "signed", "volatile", "restrict":
			case "struct", "union":
				pos, tok, lit = s.Scan()
				if tok != token.IDENT {
					log.Fatalln("c.types.ParseType: struct/union - TODO:", lit, "@", pos)
				}
				fallthrough
			default:
				if t != nil {
					return nil, newError(pos, qualType, "illegal syntax: multiple types?")
				}
				if t, err = ts.LookupType(lit, unsigned); err != nil {
					return
				}
			}
		case token.MUL: // *
			if t == nil {
				return nil, newError(pos, qualType, "pointer to nil")
			}
			t = types.NewPointer(t)
		case token.LBRACK: // [
			if t == nil {
				return nil, newError(pos, qualType, "pointer to nil")
			}
			pos, tok, lit = s.Scan()
			if tok != token.INT {
				return nil, newError(pos, qualType, "array length not an integer")
			}
			n, e := strconv.Atoi(lit)
			if e != nil {
				return nil, newError(pos, qualType, e.Error())
			}
			pos, tok, lit = s.Scan()
			if tok != token.RBRACK {
				return nil, newError(pos, qualType, "expect ]")
			}
			if isParam {
				t = types.NewPointer(t)
			} else {
				t = types.NewArray(t, int64(n))
			}
		case token.EOF:
			if t == nil {
				err = io.ErrUnexpectedEOF
			}
			return
		default:
			log.Fatalln("c.types.ParseType: unknown -", tok, lit)
		}
	}
}

// -----------------------------------------------------------------------------

type ParseTypeError struct {
	QualType string
	Pos      token.Pos
	ErrMsg   string
}

func newError(pos token.Pos, qualType, errMsg string) *ParseTypeError {
	return &ParseTypeError{QualType: qualType, Pos: pos, ErrMsg: errMsg}
}

func (p *ParseTypeError) Error() string {
	return p.ErrMsg // TODO
}

// -----------------------------------------------------------------------------
