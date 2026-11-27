package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func nativeMuxFixture(invert bool) *Gia_Man_t {
	p := gia.Gia_ManStart(32)
	p.EnableMuxes()
	a, b, c := gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p), gia.Gia_ManAppendCi(p)
	m := gia.Gia_ManAppendMuxReal(p, a, b, c)
	gia.Gia_ManAppendCo(p, m)
	gia.Gia_ManAppendCo(p, b)
	if invert {
		m ^= 1
	}
	gia.Gia_ManAppendCo(p, m)
	return p
}

// Without XOR roots there is no adder-tree chain, so no box is found and
// Acec_Solve falls back to regular CEC (C bug 32: C read past vXorRoots).
func TestSolveNativeMuxWithoutXorRootsFallsBackToCec(t *testing.T) {
	a, b := nativeMuxFixture(false), nativeMuxFixture(true)
	var pars Acec_ParCec_t
	Acec_ManCecSetDefaultParams(&pars)
	if status := Acec_Solve(a, b, &pars); status != 0 && status != 1 {
		t.Fatalf("status %d", status)
	}
}
