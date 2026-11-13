package dch

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// Sequential interfaces are treated combinationally by source DCH: every
// register output is a free SAT input, including otherwise unreachable states.
func TestSequentialSweepModesProveOnlyCombinationalEquivalences(t *testing.T) {
	rng := rand.New(rand.NewSource(84671))
	for trial := 0; trial < 16; trial++ {
		original := aig.NewGraph()
		edges := []aig.Edge{original.Const1()}
		for i := 0; i < 4; i++ {
			edges = append(edges, original.NewInput())
		}
		for i := 0; i < 24; i++ {
			a, b := edges[rng.Intn(len(edges))], edges[rng.Intn(len(edges))]
			if rng.Intn(2) != 0 {
				a = a.Not()
			}
			if rng.Intn(2) != 0 {
				b = b.Not()
			}
			edges = append(edges, original.AND(a, b))
			if i%7 == 0 {
				c := edges[rng.Intn(5)]
				edges = append(edges, original.AND(original.AND(a, b), c), original.AND(a, original.AND(b, c)))
			}
		}
		for _, e := range original.LiveEdges() {
			if original.Node(e).Kind == aig.And {
				original.AddOutput(e)
			}
		}
		original.AddOutput(edges[len(edges)-1])
		original.AddOutput(edges[len(edges)-2])
		aig.Aig_ManSetRegNum(original, 2)
		aig.Aig_ManSetConstrNum(original, 1)
		for mode := 0; mode < 8; mode++ {
			g := aig.Aig_ManDupDfs(original)
			p := DefaultParams()
			p.PolarFlip = mode&1 != 0
			p.SimulateTFO = mode&2 != 0
			p.SkipRedundantSupport = mode&4 != 0
			p.SatVarMax = 8
			p.CallsRecycle = 1
			out, err := ComputeChoices(g, &p)
			if err != nil {
				t.Fatalf("trial %d mode %d: %v", trial, mode, err)
			}
			if out.RegisterCount() != 2 || aig.Aig_ManConstrNum(out) != 1 {
				t.Fatalf("trial %d mode %d lost sequential metadata", trial, mode)
			}
			order, err := graphOrder(out)
			if err != nil {
				t.Fatal(err)
			}
			for mask := 0; mask < 16; mask++ {
				in := []bool{mask&1 != 0, mask&2 != 0, mask&4 != 0, mask&8 != 0}
				want, e := original.Evaluate(in)
				if e != nil {
					t.Fatal(e)
				}
				got, e := out.Evaluate(in)
				if e != nil {
					t.Fatal(e)
				}
				for i := range want {
					if want[i] != got[i] {
						t.Fatalf("trial %d mode %d input %d output %d", trial, mode, mask, i)
					}
				}
				vals := simulateAssignment(out, order, in)
				for _, node := range out.LiveEdges() {
					alt := aig.Aig_ObjEquiv(out, node)
					if alt == aig.InvalidEdge {
						continue
					}
					phase := aig.Aig_ObjPhase(out, node) != aig.Aig_ObjPhase(out, alt)
					if (vals[node.NodeID()] != vals[alt.NodeID()]) != phase {
						t.Fatalf("trial %d mode %d input %d unsound choice", trial, mode, mask)
					}
				}
			}
		}
	}
}
