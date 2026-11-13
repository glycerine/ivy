package dch

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// Exhaustively compare every mapped node and every published representative
// with the original graph, independently of the SAT solver and simulator.
func TestIncrementalSweepTruthTableOracle(t *testing.T) {
	rng := rand.New(rand.NewSource(0x4dc11))
	var proofs, calls int
	for trial := 0; trial < 80; trial++ {
		g := aig.NewGraph()
		edges := []aig.Edge{g.Const1()}
		for i := 0; i < 4; i++ {
			edges = append(edges, g.NewInput())
		}
		pick := func() aig.Edge {
			e := edges[rng.Intn(len(edges))]
			if rng.Intn(2) != 0 {
				e = e.Not()
			}
			return e
		}
		for i := 0; i < 40; i++ {
			a, b := pick(), pick()
			if i%3 == 0 {
				edges = append(edges, g.XOR(a, b))
			} else {
				edges = append(edges, g.AND(a, b))
			}
			// Introduce genuinely equivalent but structurally different AND cones.
			if i%10 == 0 {
				c := pick()
				edges = append(edges, g.AND(g.AND(a, b), c), g.AND(a, g.AND(b, c)))
			}
		}
		var observed []aig.Edge
		for _, edge := range g.LiveEdges() {
			if g.Node(edge).Kind != aig.Output {
				observed = append(observed, edge)
				g.AddOutput(edge)
			}
		}
		for mode := 0; mode < 4; mode++ {
			pars := DefaultParams()
			pars.PolarFlip = mode&1 != 0
			pars.SimulateTFO = mode&2 != 0
			pars.SatVarMax = 8
			pars.CallsRecycle = 2
			m, err := newManager(g, &pars)
			if err != nil {
				t.Fatal(err)
			}
			if err = m.computeEquivalences(); err != nil {
				t.Fatalf("trial %d mode %d: %v", trial, mode, err)
			}
			proofs += m.stats.Proofs
			calls += m.stats.SATCalls
			for _, edge := range observed {
				mapped := m.mapped[edge.NodeID()]
				if mapped == aig.InvalidEdge {
					t.Fatalf("trial %d mode %d: unmapped node %d", trial, mode, edge.NodeID())
				}
				m.satGraph.AddOutput(mapped)
			}
			index := make(map[int]int, len(observed))
			for i, e := range observed {
				index[e.NodeID()] = i
			}
			for bits := 0; bits < 16; bits++ {
				input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0}
				want, err := g.Evaluate(input)
				if err != nil {
					t.Fatal(err)
				}
				got, err := m.satGraph.Evaluate(input)
				if err != nil {
					t.Fatal(err)
				}
				for i, edge := range observed {
					if got[i] != want[i] {
						t.Fatalf("trial %d mode %d input %d mapped node %d: got %v want %v", trial, mode, bits, edge.NodeID(), got[i], want[i])
					}
					if rep, ok := m.proven[edge.NodeID()]; ok {
						r := index[rep.NodeID()]
						phase := g.Node(edge).Phase != g.Node(rep).Phase
						if want[i] != (want[r] != phase) {
							t.Fatalf("trial %d mode %d input %d invalid representative %d -> %d", trial, mode, bits, edge.NodeID(), rep.NodeID())
						}
					}
				}
			}
		}
	}
	if proofs == 0 || calls == 0 {
		t.Fatalf("oracle did not exercise proofs and SAT queries: proofs=%d calls=%d", proofs, calls)
	}
}
