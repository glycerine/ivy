package pdr

import (
	"testing"

	"github.com/glycerine/abcgo/misc/vec"
)

// pdrSat.c keeps a frame solver while its activation count is below the
// recycle limit, and asserts if the last solver reaches that limit.
func TestRecycleLimitBoundaryAndLastFrameAssertion(t *testing.T) {
	p, _ := pdrTestSATManager(t, false)
	Pdr_ManCreateSolver(p, 1)
	p.pPars.nRecycle = 2
	original := p.vSolvers[0]
	vec.Vec_IntWriteEntry(p.vActVars, 0, 1)
	if got := Pdr_ManFetchSolver(p, 0); got != original || p.nStarts != 0 {
		t.Fatal("solver was recycled below the limit")
	}
	vec.Vec_IntWriteEntry(p.vActVars, 0, 2)
	if got := Pdr_ManFetchSolver(p, 0); got != original || p.nStarts != 1 || vec.Vec_IntEntry(p.vActVars, 0) != 0 {
		t.Fatal("solver was not recycled at the limit")
	}
	vec.Vec_IntWriteEntry(p.vActVars, 1, 2)
	defer func() {
		if recover() == nil {
			t.Fatal("recycling the last frame did not assert")
		}
	}()
	Pdr_ManFetchSolver(p, 1)
}
