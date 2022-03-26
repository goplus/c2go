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
	Pkg() *types.Package
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
	p := &parser{ts: ts}
	file := fset.AddFile("", fset.Base(), len(qualType))
	p.s.Init(file, qualType, nil)

	if t, err = p.parse(isParam); err != nil {
		return
	}
	if p.tok != token.EOF {
		err = p.newError("unexpect token " + p.tok.String())
	}
	return
}

// -----------------------------------------------------------------------------

type parser struct {
	s  scanner.Scanner
	ts TypeSystem

	pos token.Pos
	tok token.Token
	lit string

	tyVoid types.Type
}

func (p *parser) notVoid(t types.Type) bool {
	return p.void() != t
}

func (p *parser) void() types.Type {
	if p.tyVoid == nil {
		p.tyVoid, _ = p.ts.LookupType("void", false)
	}
	return p.tyVoid
}

func (p *parser) next() {
	p.pos, p.tok, p.lit = p.s.Scan()
}

func (p *parser) newError(errMsg string) *ParseTypeError {
	return &ParseTypeError{QualType: p.s.Source(), Pos: p.pos, ErrMsg: errMsg}
}

func (p *parser) expect(tokExp token.Token) error {
	p.next()
	if p.tok != tokExp {
		return p.newError("expect " + tokExp.String())
	}
	return nil
}

func (p *parser) parse(isParam bool) (t types.Type, err error) {
	unsigned := false
	for {
		p.next()
		switch p.tok {
		case token.IDENT:
			switch p.lit {
			case "unsigned":
				unsigned = true
			case "const", "signed", "volatile", "restrict":
			case "struct", "union":
				p.next()
				if p.tok != token.IDENT {
					log.Fatalln("c.types.ParseType: struct/union - TODO:", p.lit, "@", p.pos)
				}
				fallthrough
			default:
				if t != nil {
					return nil, p.newError("illegal syntax: multiple types?")
				}
				if t, err = p.ts.LookupType(p.lit, unsigned); err != nil {
					return
				}
			}
		case token.MUL: // *
			if t == nil {
				return nil, p.newError("pointer to nil")
			}
			t = types.NewPointer(t)
		case token.LBRACK: // [
			if t == nil {
				return nil, p.newError("pointer to nil")
			}
			p.next()
			if p.tok != token.INT {
				return nil, p.newError("array length not an integer")
			}
			n, e := strconv.Atoi(p.lit)
			if e != nil {
				return nil, p.newError(e.Error())
			}
			if err = p.expect(token.RBRACK); err != nil { // ]
				return
			}
			if isParam {
				t = types.NewPointer(t)
			} else {
				t = types.NewArray(t, int64(n))
			}
		case token.LPAREN: // (
			if t == nil {
				return nil, p.newError("no function return type")
			}
			if err = p.expect(token.MUL); err != nil { // *
				return
			}
			if err = p.expect(token.RPAREN); err != nil { // )
				return
			}
			if err = p.expect(token.LPAREN); err != nil { // (
				return
			}
			var args []*types.Var
			var results *types.Tuple
			var pkg = p.ts.Pkg()
			for {
				arg, e := p.parse(true)
				if e != nil {
					return nil, e
				}
				if p.notVoid(arg) {
					args = append(args, types.NewParam(token.NoPos, pkg, "", arg))
				}
				if p.tok != token.COMMA {
					break
				}
			}
			if p.tok != token.RPAREN { // )
				return nil, p.newError("expect )")
			}
			if p.notVoid(t) {
				results = types.NewTuple(types.NewParam(token.NoPos, pkg, "", t))
			}
			t = types.NewSignature(nil, types.NewTuple(args...), results, false)
		case token.RPAREN:
			if t == nil {
				t = p.void()
			}
			return
		case token.COMMA, token.EOF:
			if t == nil {
				err = io.ErrUnexpectedEOF
			}
			return
		default:
			log.Fatalln("c.types.ParseType: unknown -", p.tok, p.lit)
		}
	}
}

// -----------------------------------------------------------------------------

type ParseTypeError struct {
	QualType string
	Pos      token.Pos
	ErrMsg   string
}

func (p *ParseTypeError) Error() string {
	return p.ErrMsg // TODO
}

// -----------------------------------------------------------------------------
