package parser

import (
	"bytes"
	"encoding/json"
	"os/exec"

	"github.com/goplus/c2go/clang/ast"
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
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(
		"clang", "-Xclang", "-ast-dump=json", "-fsyntax-only", "-fno-color-diagnostics", filename)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	stdout.Grow(0x200000)
	err := cmd.Run()
	if err != nil {
		return nil, &ParseFileError{Err: err, Stderr: stderr.Bytes(), File: filename}
	}
	return stdout.Bytes(), nil
}

// -----------------------------------------------------------------------------

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
