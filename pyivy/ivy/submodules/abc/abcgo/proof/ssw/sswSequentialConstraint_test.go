package ssw

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestConstrainedSweepCarriesRegisterConstraintAcrossFrames(t *testing.T) {
	g := aig.NewGraph()
	pi, q := g.NewInput(), g.NewInput()
	prop := g.AND(pi, q.Not())
	g.AddOutput(prop)
	g.AddOutput(q)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	var pars Ssw_Pars_t
	Ssw_ManSetDefaultParams(&pars)
	pars.FConstrs = 1
	pars.NFramesK = 2
	pars.nStepsMax = 3
	reduced := Ssw_SignalCorrespondence(g, &pars)
	if reduced == nil {
		t.Fatal("initially satisfiable constraint rejected")
	}
	for seq := 0; seq < 16; seq++ {
		oldState, newState := false, false
		for frame := 0; frame < 4; frame++ {
			if oldState {
				break
			} // q=1 violates the constraint in this frame.
			input := seq&(1<<frame) != 0
			old, err := g.Evaluate([]bool{input, oldState})
			if err != nil {
				t.Fatal(err)
			}
			args := []bool{input}
			if aig.Aig_ManRegNum(reduced) > 0 {
				args = append(args, newState)
			}
			got, err := reduced.Evaluate(args)
			if err != nil {
				t.Fatal(err)
			}
			if old[0] != got[0] {
				t.Fatalf("sequence=%d frame=%d old=%v reduced=%v", seq, frame, old, got)
			}
			oldState = old[2]
			if aig.Aig_ManRegNum(reduced) > 0 {
				newState = got[len(got)-1]
			}
		}
	}
}
