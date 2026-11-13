package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"testing"
)

func TestLatchCorrespondenceAgainstExhaustiveTraces(t *testing.T) {
	rng := rand.New(rand.NewSource(4831))
	for trial := 0; trial < 48; trial++ {
		g := aig.NewGraph()
		edges := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 8; i++ {
			a, b := edges[rng.Intn(len(edges))], edges[rng.Intn(len(edges))]
			if rng.Intn(2) == 0 {
				a = a.Not()
			}
			if rng.Intn(2) == 0 {
				b = b.Not()
			}
			if rng.Intn(2) == 0 {
				edges = append(edges, g.AND(a, b))
			} else {
				edges = append(edges, g.XOR(a, b))
			}
		}
		for i := 0; i < 4; i++ {
			e := edges[rng.Intn(len(edges))]
			if rng.Intn(2) == 0 {
				e = e.Not()
			}
			g.AddOutput(e)
		}
		aig.Aig_ManSetRegNum(g, 2)
		for _, prefix := range []int32{0, 1} {
			inputGraph := aig.Aig_ManDupOrdered(g)
			reduced := Fra_FraigLatchCorrespondence(inputGraph, prefix, 10000, 0, 0, nil, 0)
			if reduced == nil {
				t.Fatalf("trial %d prefix %d returned no reduced graph", trial, prefix)
			}
			for sequence := 0; sequence < 256; sequence++ {
				originalState, reducedState := make([]bool, 2), make([]bool, reduced.RegisterCount())
				for frame := 0; frame < 4; frame++ {
					input := []bool{sequence&(1<<(2*frame)) != 0, sequence&(1<<(2*frame+1)) != 0}
					original, err := g.Evaluate(append(append([]bool(nil), input...), originalState...))
					if err != nil {
						t.Fatal(err)
					}
					got, err := reduced.Evaluate(append(append([]bool(nil), input...), reducedState...))
					if err != nil {
						t.Fatal(err)
					}
					if original[0] != got[0] || original[1] != got[1] {
						t.Fatalf("trial %d prefix %d sequence %d frame %d original %v reduced %v", trial, prefix, sequence, frame, original, got)
					}
					copy(originalState, original[len(original)-len(originalState):])
					copy(reducedState, got[len(got)-len(reducedState):])
				}
			}
		}
	}
}
