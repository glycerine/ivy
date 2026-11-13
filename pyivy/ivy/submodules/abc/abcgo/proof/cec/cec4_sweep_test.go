package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestCec4SweepPreservesOutputs(t *testing.T) {
	for mode := int32(0); mode < 3; mode++ {
		g := gia.Gia_ManStart(16)
		a, b, c := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		x := gia.Gia_ManAppendAnd(g, a, b)
		y := gia.Gia_ManAppendAnd(g, x, a)
		z := gia.Gia_ManAppendAnd(g, y, c^1)
		gia.Gia_ManAppendCo(g, x)
		gia.Gia_ManAppendCo(g, z)
		p := Cec_ParFra_t{jType: mode, nWords: 1, nRounds: 1, nItersMax: 10, nBTLimit: 1000}
		out, cex, stats := cec4Sweep(g, &p, false, true)
		if stats.unsat == 0 {
			t.Fatalf("mode %d: no SAT proof exercised", mode)
		}
		if cex != nil || out == nil {
			t.Fatalf("mode %d: unexpected result", mode)
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
					t.Fatalf("mode %d input %v: %v want %v", mode, in, got, want)
				}
			}
		}
	}
}
