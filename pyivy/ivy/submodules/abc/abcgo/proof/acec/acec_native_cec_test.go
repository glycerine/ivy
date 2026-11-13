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
func TestSolveNativeMuxFallback(t *testing.T) {
	for _, invert := range []bool{false, true} {
		a, b := nativeMuxFixture(false), nativeMuxFixture(invert)
		pars := &Acec_ParCec_t{}
		Acec_ManCecSetDefaultParams(pars)
		got := Acec_Solve(a, b, pars)
		want := int32(1)
		if invert {
			want = 0
		}
		if got != want {
			t.Fatalf("invert %v: got %d want %d", invert, got, want)
		}
	}
}
