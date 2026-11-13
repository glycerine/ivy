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

func TestCec4SweepPreservesSmallAIGs(t *testing.T) {
	for mode := int32(0); mode < 3; mode++ {
		for seed := uint32(1); seed <= 24; seed++ {
			g := gia.Gia_ManStart(32)
			lits := []int32{gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)}
			state := seed
			next := func() uint32 { state = state*1664525 + 1013904223; return state }
			for gate := 0; gate < 6; gate++ {
				a := lits[int(next())%len(lits)] ^ int32(next()&1)
				b := lits[int(next())%len(lits)] ^ int32(next()&1)
				lits = append(lits, gia.Gia_ManAppendAnd2(g, a, b))
			}
			gia.Gia_ManAppendCo(g, lits[len(lits)-1])
			gia.Gia_ManAppendCo(g, lits[len(lits)-2]^1)
			p := Cec_ParFra_t{jType: mode, nWords: 1, nRounds: 1, nItersMax: 10, nBTLimit: 100}
			out, cex, _ := cec4Sweep(g, &p, false, true)
			if cex != nil || out == nil {
				t.Fatalf("mode %d seed %d: missing output", mode, seed)
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
					if want[i] != got[i] {
						t.Fatalf("mode %d seed %d input %v got %v want %v", mode, seed, in, got, want)
					}
				}
			}
		}
	}
}
