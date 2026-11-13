package acec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// Acec_Solve swaps the old witness with the miter's witness only when miter
// construction succeeds; ABC leaves it untouched when construction fails.
func TestSourceSolveWitnessOnMiterFailureAndProof(t *testing.T) {
	for _, mismatch := range []bool{true, false} {
		a, b := reviewAdderChain(-1), reviewAdderChain(-1)
		if mismatch {
			gia.Gia_ManAppendCo(b, 0)
		}
		if err := a.SetCombinationalCounterexample([]bool{true, false, false, false, false}, 0); err != nil {
			t.Fatal(err)
		}
		old := a.CombinationalCounterexample()
		var p Acec_ParCec_t
		Acec_ManCecSetDefaultParams(&p)
		got := Acec_Solve(a, b, &p)
		if mismatch {
			if got != -1 || a.CombinationalCounterexample() != old {
				t.Fatalf("miter failure status=%d witness=%v", got, a.CombinationalCounterexample())
			}
		} else if got != 1 || a.CombinationalCounterexample() != nil {
			t.Fatalf("proof status=%d witness=%v", got, a.CombinationalCounterexample())
		}
	}
}
