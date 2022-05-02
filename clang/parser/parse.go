package parser

import (
	"bytes"
	"os/exec"

	"github.com/goplus/c2go/clang/ast"
	jsoniter "github.com/json-iterator/go"
)

type Mode uint

// -----------------------------------------------------------------------------

type ParseError struct {
	Err    error
	Stderr []byte
}

func (p *ParseError) Error() string {
	if len(p.Stderr) > 0 {
		return string(p.Stderr)
	}
	return p.Err.Error()
}

// -----------------------------------------------------------------------------

type Config struct {
	Json *[]byte
}

func DumpAST(filename string, conf *Config) (result []byte, warning []byte, err error) {
	stdout := NewPagedWriter()
	stderr := new(bytes.Buffer)
	cmd := exec.Command(
		"clang", "-Xclang", "-ast-dump=json", "-fsyntax-only", filename)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err = cmd.Run()
	errmsg := stderr.Bytes()
	if err != nil {
		return nil, nil, &ParseError{Err: err, Stderr: errmsg}
	}
	return stdout.Bytes(), errmsg, nil
}

// -----------------------------------------------------------------------------

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func ParseFileEx(filename string, mode Mode, conf *Config) (file *ast.Node, warning []byte, err error) {
	out, warning, err := DumpAST(filename, conf)
	if err != nil {
		return
	}
	if conf != nil && conf.Json != nil {
		*conf.Json = out
	}
	file = new(ast.Node)
	err = json.Unmarshal(out, file)
	if err != nil {
		err = &ParseError{Err: err}
	}
	return
}

func ParseFile(filename string, mode Mode) (file *ast.Node, warning []byte, err error) {
	return ParseFileEx(filename, mode, nil)
}

// -----------------------------------------------------------------------------
