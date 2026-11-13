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
