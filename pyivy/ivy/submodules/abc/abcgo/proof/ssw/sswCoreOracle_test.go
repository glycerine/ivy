package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"testing"
)

func TestSignalCorrespondencePreservesSmallSequentialCircuits(t *testing.T) {
	r := rand.New(rand.NewSource(373))
	for trial := 0; trial < 20; trial++ {
		g := aig.NewGraph()
		pi, lo := g.NewInput(), g.NewInput()
		nodes := []aig.Edge{pi, lo, g.Const1(), g.Const0()}
		for k := 0; k < 8; k++ {
			a, b := nodes[r.Intn(len(nodes))], nodes[r.Intn(len(nodes))]
			if r.Intn(2) == 0 {
				a = a.Not()
			}
			if r.Intn(2) == 0 {
				b = b.Not()
			}
			nodes = append(nodes, g.AND(a, b))
		}
		output, next := nodes[4+r.Intn(len(nodes)-4)], nodes[4+r.Intn(len(nodes)-4)]
		if r.Intn(2) == 0 {
			output = output.Not()
		}
		if r.Intn(2) == 0 {
			next = next.Not()
		}
		g.AddOutput(output)
		g.AddOutput(next)
		aig.Aig_ManSetRegNum(g, 1)
		var pars Ssw_Pars_t
		Ssw_ManSetDefaultParams(&pars)
		pars.nStepsMax = 3
		reduced := Ssw_SignalCorrespondence(g, &pars)
		if reduced == nil {
			t.Fatalf("trial %d: no reduced graph", trial)
		}
		for sequence := 0; sequence < 16; sequence++ {
			oldState, newState := false, false
			for frame := 0; frame < 4; frame++ {
				input := sequence&(1<<uint(frame)) != 0
				old, err := g.Evaluate([]bool{input, oldState})
				if err != nil {
					t.Fatal(err)
				}
				ins := []bool{input}
				if aig.Aig_ManRegNum(reduced) > 0 {
					ins = append(ins, newState)
				}
				got, err := reduced.Evaluate(ins)
				if err != nil {
					t.Fatalf("trial %d: %v", trial, err)
				}
				if old[0] != got[0] {
					t.Fatalf("trial %d sequence %04b frame %d: output differs", trial, sequence, frame)
				}
				oldState = old[1]
				if aig.Aig_ManRegNum(reduced) > 0 {
					newState = got[1]
				}
			}
		}
	}
}
