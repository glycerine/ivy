package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

// fraHot.c resets the AIG RNG, then consumes exactly 16384 words for each
// register, even if there are no one-hot clauses. The next value below was
// obtained by calling the original compiled C routine on this fixture.
func TestOneHotCoverageResetsAndAdvancesSourceRNG(t *testing.T) {
	g := aig.NewGraph()
	g.NewInput()
	g.NewInput()
	g.AddOutput(g.Const0())
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 2)
	p := &Fra_Man_t{PManAig: g}
	clauses := vec.Vec_IntAlloc(0)
	const want uint32 = 112181958
	for round := 0; round < 2; round++ {
		for i := 0; i < 17; i++ {
			aig.Aig_ManRandom(0)
		}
		Fra_OneHotEstimateCoverage(p, clauses)
		if got := aig.Aig_ManRandom(0); got != want {
			t.Fatalf("round %d: next random %#x, want %#x", round, got, want)
		}
	}
}
