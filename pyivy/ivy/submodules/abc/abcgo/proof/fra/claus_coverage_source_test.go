package fra

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/cnf"
)

// fraClaus.c resets ABC's random generator before every coverage estimate.
func TestClauseCoverageResetsSimulationRandomState(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x)
	p := Fra_ClausAlloc(g, 1, 0, 4, 4, 1, 2, 1, 0, 0, 0, 0)
	p.pCnf = cnf.Cnf_Derive(g, 0)
	Fra_ClausEstimateCoverage(p)
	a := aig.Aig_ManRandom(0)
	for i := 0; i < 19; i++ {
		aig.Aig_ManRandom(0)
	}
	Fra_ClausEstimateCoverage(p)
	b := aig.Aig_ManRandom(0)
	if a != b {
		t.Fatalf("coverage left different random states: %d and %d", a, b)
	}
}

// Coverage marks assignments that falsify every literal of a clause. Unlike
// class comparisons, ABC examines all simulation words, including the prefix.
func TestClauseCoverageUsesEveryWordAndLiteralPolarity(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	sim := Fra_SmlStart(g, 1, 1, 1)
	copy(sim.NodeWords(x.NodeID()), []uint32{0xaaaaaaaa, 0x0f0f0f0f})
	copy(sim.NodeWords(y.NodeID()), []uint32{0xcccccccc, 0x33333333})
	ids := []int{x.NodeID(), y.NodeID()}
	got := make([]uint32, 2)
	Fra_ClausEstimateCoverageOne(sim, []int32{0, 3}, ids, got) // x OR !y
	want := []uint32{0x44444444, 0x30303030}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("word %d = %#x, want %#x", i, got[i], want[i])
		}
	}
	Fra_ClausEstimateCoverageOne(sim, nil, ids, got)
	if got[0] != ^uint32(0) || got[1] != ^uint32(0) {
		t.Fatalf("empty clause coverage = %x", got)
	}
}
