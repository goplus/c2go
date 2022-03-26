package parser

import (
	"bytes"
	"os/exec"

	"github.com/goplus/c2go/clang/ast"
	jsoniter "github.com/json-iterator/go"
)

type Mode uint

// -----------------------------------------------------------------------------

type ParseFileError struct {
	Err    error
	Stderr []byte
	File   string
}

func (p *ParseFileError) Error() string {
	return p.Err.Error() // TODO:
}

// -----------------------------------------------------------------------------

func DumpAST(filename string) ([]byte, error) {
	stdout := NewPagedWriter()
	stderr := new(bytes.Buffer)
	cmd := exec.Command(
		"clang", "-Xclang", "-ast-dump=json", "-fsyntax-only", "-fno-color-diagnostics", filename)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		return nil, &ParseFileError{Err: err, Stderr: stderr.Bytes(), File: filename}
	}
	return stdout.Bytes(), nil
}

// -----------------------------------------------------------------------------

var json = jsoniter.ConfigCompatibleWithStandardLibrary

func ParseFile(filename string, mode Mode) (*ast.Node, error) {
	out, err := DumpAST(filename)
	if err != nil {
		return nil, err
	}
	var doc ast.Node
	err = json.Unmarshal(out, &doc)
	if err != nil {
		return nil, &ParseFileError{Err: err, File: filename}
	}
	return &doc, nil
}

// -----------------------------------------------------------------------------
