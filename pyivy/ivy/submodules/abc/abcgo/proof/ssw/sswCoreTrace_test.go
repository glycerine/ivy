package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSignalCorrespondencePreservesSequentialTraces(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(pi, lo))
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	var pars Ssw_Pars_t
	Ssw_ManSetDefaultParams(&pars)
	pars.nStepsMax = 1
	reduced := Ssw_SignalCorrespondence(g, &pars)
	for sequence := 0; sequence < 8; sequence++ {
		oldState, newState := false, false
		for frame := 0; frame < 3; frame++ {
			input := sequence&(1<<uint(frame)) != 0
			old, err := g.Evaluate([]bool{input, oldState})
			if err != nil {
				t.Fatal(err)
			}
			next, err := reduced.Evaluate([]bool{input, newState})
			if err != nil {
				t.Fatal(err)
			}
			if old[0] != next[0] {
				t.Fatalf("sequence %03b frame %d differs", sequence, frame)
			}
			oldState, newState = old[1], next[1]
		}
	}
}
