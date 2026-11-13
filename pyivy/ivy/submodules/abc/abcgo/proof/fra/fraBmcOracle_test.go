package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"math/rand"
	"testing"
)

func TestBoundedMiterAgainstEnumeratedTraces(t *testing.T) {
	rng := rand.New(rand.NewSource(97412))
	for trial := 0; trial < 80; trial++ {
		g := aig.NewGraph()
		edges := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
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
			edge := edges[rng.Intn(len(edges))]
			if rng.Intn(2) == 0 {
				edge = edge.Not()
			}
			g.AddOutput(edge)
		}
		aig.Aig_ManSetRegNum(g, 2)
		// All 3-bit input sequences from initial register state zero.
		fails := false
		for sequence := 0; sequence < 8; sequence++ {
			r0, r1 := false, false
			for frame := 0; frame < 3; frame++ {
				outputs, err := g.Evaluate([]bool{sequence&(1<<frame) != 0, r0, r1})
				if err != nil {
					t.Fatal(err)
				}
				fails = fails || outputs[0] || outputs[1]
				r0, r1 = outputs[2], outputs[3]
			}
		}
		for _, rewrite := range []int32{0, 1} {
			g.ClearSequentialCounterexample()
			Fra_BmcPerformSimple(g, 3, 10000, rewrite, 0)
			cex := g.SequentialCounterexample()
			if cex != nil && (cex.IFrame >= 3 || saig.Saig_ManVerifyCex(g, cex) == 0) {
				t.Fatalf("trial %d rewrite=%d invalid cex %+v", trial, rewrite, cex)
			}
			// The source fast path can miss a real direct-input witness. It must
			// never fabricate a witness for a trace that has no failing output.
			if cex != nil && !fails {
				t.Fatalf("trial %d rewrite=%d BMC witness %v but oracle does not fail", trial, rewrite, cex)
			}
		}
	}
}
