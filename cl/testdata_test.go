package cl_test

import (
	"testing"

	"github.com/goplus/c2go/cl"
	"github.com/goplus/c2go/internal/c2go"
)

// -----------------------------------------------------------------------------

func TestFromTestdata(t *testing.T) {
	cl.SetDebug(0)
	defer cl.SetDebug(cl.DbgFlagAll)

	c2go.Run("", "../testdata/...", c2go.FlagRunTest|c2go.FlagFailFast, nil)
}

// -----------------------------------------------------------------------------
