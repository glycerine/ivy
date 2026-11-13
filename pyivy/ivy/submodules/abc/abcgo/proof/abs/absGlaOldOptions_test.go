package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestOldGlaProofOptionPreservesResults(t *testing.T) {
	for _, full := range []int32{0, 1} {
		for _, satisfiable := range []bool{false, true} {
			g := gia.Gia_ManStart(16)
			x := gia.Gia_ManAppendCi(g)
			y := gia.Gia_ManAppendCi(g)
			a := gia.Gia_ManAppendAnd(g, x, y)
			b := gia.Gia_ManAppendAnd(g, x^1, y)
			out := a
			if !satisfiable {
				out = gia.Gia_ManAppendAnd(g, a, b)
			}
			gia.Gia_ManAppendCo(g, out)
			var pars Abs_Par_t
			Abs_ParSetDefaults(&pars)
			pars.fUseFullProof = full
			pars.nFramesMax = 1
			pars.nRatioMin = 0
			pars.nRatioMin2 = 0
			got := Gia_ManPerformGlaOld(g, &pars, 0)
			if satisfiable {
				if got != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
					t.Fatalf("full=%d invalid SAT result %d", full, got)
				}
			} else if got != -1 || g.SequentialCounterexample() != nil {
				t.Fatalf("full=%d invalid bounded UNSAT result %d", full, got)
			}
		}
	}
}

// Both source proof modes must handle a spurious abstract witness and retain
// a sound abstraction after core-driven rollback. The Go port currently uses
// the same full graph in both cases; this tests outcomes, not memory parity.
func TestOldGlaProofModesRefineAndRollback(t *testing.T) {
	for _, full := range []int32{0, 1} {
		g := gia.Gia_ManStart(20)
		var x [5]int32
		for i := range x {
			x[i] = gia.Gia_ManAppendCi(g)
		}
		left := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, x[0], x[1]), x[2])
		right := gia.Gia_ManAppendAnd(g, gia.Gia_ManAppendAnd(g, x[0]^1, x[3]), x[4])
		gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, left, right))
		p := NewParameters()
		p.fUseFullProof, p.nFramesMax, p.nRatioMin = full, 1, 0
		p.nFramesNoChange = 10
		got := Gia_ManPerformGlaOld(g, p, 0)
		if got != -1 || g.SequentialCounterexample() != nil || p.nFramesNoChange != 0 {
			t.Fatalf("full=%d result=%d witness=%+v unchanged=%d", full, got, g.SequentialCounterexample(), p.nFramesNoChange)
		}
		if gia.Gia_ManGateClasses(g) == nil {
			t.Fatalf("full=%d missing resulting abstraction", full)
		}
	}
}
