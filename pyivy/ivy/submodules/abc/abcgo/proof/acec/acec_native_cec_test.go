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
func TestSolveNativeMuxWithoutXorRootsViolatesSourcePrecondition(t *testing.T) {
	a, b := nativeMuxFixture(false), nativeMuxFixture(true)
	var pars Acec_ParCec_t
	Acec_ManCecSetDefaultParams(&pars)
	defer func() {
		if recover() == nil {
			t.Fatal("missing source root-ordering assertion")
		}
	}()
	Acec_Solve(a, b, &pars)
}
