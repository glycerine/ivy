package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
	"time"
)

// CEC correspondence must preserve every reachable output trace, including
// when one register is eliminated. Both delegation modes are exercised.
func TestCecDelegationPreservesReachableTraces(t *testing.T) {
	for _, latch := range []bool{false, true} {
		g := aig.NewGraph()
		pi, a, b := g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(a, b))
		g.AddOutput(pi)
		g.AddOutput(pi)
		aig.Aig_ManSetRegNum(g, 2)
		var pars Ssw_Pars_t
		Ssw_ManSetDefaultParams(&pars)
		pars.fScorrGia = 1
		if latch {
			pars.fLatchCorrOpt = 1
		}
		reduced := Ssw_SignalCorrespondence(g, &pars)
		if reduced == nil {
			t.Fatalf("latch=%t returned nil", latch)
		}
		for sequence := 0; sequence < 16; sequence++ {
			oldState := []bool{false, false}
			newState := make([]bool, int(reduced.RegisterCount()))
			for frame := 0; frame < 4; frame++ {
				input := sequence&(1<<frame) != 0
				old, err := g.Evaluate(append([]bool{input}, oldState...))
				if err != nil {
					t.Fatal(err)
				}
				got, err := reduced.Evaluate(append([]bool{input}, newState...))
				if err != nil {
					t.Fatal(err)
				}
				if got[0] != old[0] {
					t.Fatalf("latch=%t sequence=%d frame=%d", latch, sequence, frame)
				}
				copy(oldState, old[1:])
				copy(newState, got[1:])
			}
		}
	}
}

func TestCecDelegationRestoresConfiguredTimeout(t *testing.T) {
	for _, latch := range []bool{false, true} {
		g := sswDeadlineGraph()
		var pars Ssw_Pars_t
		Ssw_ManSetDefaultParams(&pars)
		pars.fScorrGia = 1
		pars.TimeLimit = 5
		if latch {
			pars.fLatchCorrOpt = 1
		}
		deadline := time.Now().Add(time.Hour)
		Ssw_ParsSetDeadline(&pars, deadline)
		if Ssw_SignalCorrespondence(g, &pars) == nil {
			t.Fatalf("latch=%t did not complete before deadline", latch)
		}
		if !pars.deadline.Equal(deadline) {
			t.Fatal("caller deadline was not restored")
		}
	}
}
