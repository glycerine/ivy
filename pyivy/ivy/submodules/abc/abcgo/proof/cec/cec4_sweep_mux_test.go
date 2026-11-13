package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4SweepPreservesNativeMuxOutputs(t *testing.T) {
	for mode := int32(0); mode < 3; mode++ {
		g := gia.Gia_ManStart(12)
		g.EnableMuxes()
		c, a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x, y := gia.Gia_ManAppendMuxReal(g, c, a, b), gia.Gia_ManAppendMuxReal(g, c, a, b)
		gia.Gia_ManAppendCo(g, x)
		gia.Gia_ManAppendCo(g, y^1)
		pars := Cec_ParFra_t{jType: mode, nWords: 1, nRounds: 1, nItersMax: 10, nBTLimit: 100}
		out, cex, _ := cec4Sweep(g, &pars, false, true)
		if out == nil || cex != nil {
			t.Fatalf("mode %d: no reduced graph", mode)
		}
		for mask := 0; mask < 8; mask++ {
			in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0}
			want, e := g.Evaluate(in)
			if e != nil {
				t.Fatal(e)
			}
			got, e := out.Evaluate(in)
			if e != nil {
				t.Fatal(e)
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("mode %d inputs %v got %v want %v", mode, in, got, want)
				}
			}
		}
	}
}
