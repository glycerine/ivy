package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestConstrainedSignalCorrespondencePreservesAllowedTraces(t *testing.T) {
	g := aig.NewGraph()
	a, b, q := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.OR(q, g.AND(a, b)))
	g.AddOutput(a)
	g.AddOutput(b)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	var pars Ssw_Pars_t
	Ssw_ManSetDefaultParams(&pars)
	pars.fConstrs = 1
	pars.nStepsMax = 3
	reduced := Ssw_SignalCorrespondence(g, &pars)
	if reduced == nil {
		t.Fatal("constraint-compatible graph was rejected")
	}
	for seq := 0; seq < 16; seq++ {
		oldState, newState := false, false
		for frame := 0; frame < 4; frame++ {
			bval := seq&(1<<frame) != 0
			old, err := g.Evaluate([]bool{false, bval, oldState})
			if err != nil {
				t.Fatal(err)
			}
			args := []bool{false, bval}
			if aig.Aig_ManRegNum(reduced) > 0 {
				args = append(args, newState)
			}
			got, err := reduced.Evaluate(args)
			if err != nil {
				t.Fatal(err)
			}
			if old[0] != got[0] {
				t.Fatalf("sequence=%d frame=%d original=%v reduced=%v", seq, frame, old, got)
			}
			oldState = old[2]
			if aig.Aig_ManRegNum(reduced) > 0 {
				newState = got[len(got)-1]
			}
		}
	}
}
