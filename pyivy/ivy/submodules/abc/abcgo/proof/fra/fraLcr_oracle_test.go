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
				edges = append(edges, g.NewXor(a, b))
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
		reduced := Fra_FraigLatchCorrespondence(g, 0, 10000, 0, 0, nil, 0)
		if reduced == nil {
			t.Fatalf("trial %d returned no reduced graph", trial)
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
					t.Fatalf("trial %d sequence %d frame %d original %v reduced %v", trial, sequence, frame, original, got)
				}
				copy(originalState, original[len(original)-len(originalState):])
				copy(reducedState, got[len(got)-len(reducedState):])
			}
		}
	}
}
