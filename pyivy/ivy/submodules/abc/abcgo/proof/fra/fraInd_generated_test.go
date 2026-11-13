package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"math/rand"
	"testing"
)

func TestInductionGeneratedTwoRegisterCircuits(t *testing.T) {
	r := rand.New(rand.NewSource(7919))
	completed := 0
	for trial := 0; trial < 24; trial++ {
		for _, config := range []struct{ frames, partition int32 }{{1, 0}, {2, 0}, {1, 1}} {
			frames := config.frames
			g := aig.NewGraph()
			edges := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput(), g.NewInput()}
			for n := 0; n < 9; n++ {
				a, b := edges[r.Intn(len(edges))], edges[r.Intn(len(edges))]
				if r.Intn(2) == 0 {
					a = a.Not()
				}
				if r.Intn(2) == 0 {
					b = b.Not()
				}
				if r.Intn(2) == 0 {
					edges = append(edges, g.AND(a, b))
				} else {
					edges = append(edges, g.XOR(a, b))
				}
			}
			for i := 0; i < 4; i++ {
				e := edges[r.Intn(len(edges))]
				if r.Intn(2) == 0 {
					e = e.Not()
				}
				g.AddOutput(e)
			}
			aig.Aig_ManSetRegNum(g, 2)
			reduced := Fra_FraigInduction(g, &Fra_Ssw_t{nFramesK: frames, nPartSize: config.partition})
			if reduced == nil {
				// A stalled refinement returns unknown rather than an unsafe result.
				continue
			}
			completed++
			for seq := 0; seq < 256; seq++ {
				state := make([]bool, 2)
				newstate := make([]bool, reduced.RegisterCount())
				for f := 0; f < 4; f++ {
					in := []bool{seq&(1<<(2*f)) != 0, seq&(1<<(2*f+1)) != 0}
					a, err := g.Evaluate(append(append([]bool{}, in...), state...))
					if err != nil {
						t.Fatal(err)
					}
					b, err := reduced.Evaluate(append(append([]bool{}, in...), newstate...))
					if err != nil {
						t.Fatal(err)
					}
					if a[0] != b[0] || a[1] != b[1] {
						t.Fatalf("trial %d frames %d seq %d frame %d got %v expected %v", trial, frames, seq, f, b, a)
					}
					copy(state, a[len(a)-2:])
					copy(newstate, b[len(b)-len(newstate):])
				}
			}
		}
	}
	if completed != 72 {
		t.Fatalf("only %d of 72 induction runs completed", completed)
	}
}
