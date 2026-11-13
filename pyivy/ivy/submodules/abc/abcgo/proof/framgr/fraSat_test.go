package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func satTestManager(graph *aig.Graph) *Fra_Man_t {
	var pars Fra_Par_t
	Fra_ParamsDefault(&pars)
	p := Fra_ManStart(graph, &pars)
	p.PManFraig = graph
	Fra_ManClean(p, int32(len(graph.Nodes())))
	return p
}

func TestSATImplicationProofAndWitness(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	both := g.AND(a, b)
	p := satTestManager(g)
	if got := Fra_NodesAreImp(p, both, a, 0, 0); got != 1 {
		t.Fatalf("(a and b) => a: %d", got)
	}
	if got := Fra_NodesAreImp(p, a, both, 0, 0); got != 0 {
		t.Fatalf("a => (a and b): %d", got)
	}
	if p.PPatWords[0]&1 == 0 || p.PPatWords[0]&2 != 0 {
		t.Fatalf("unexpected counterexample bits: %b", p.PPatWords[0])
	}
}

func TestSATEquivalenceAndClause(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	native := g.NewXor(a, b)
	expanded := aig.Aig_Regular(g.XOR(a, b))
	p := satTestManager(g)
	// The expanded XOR's regular node is complemented relative to XOR, so
	// phase normalization should recognize the pair as equivalent up to phase.
	if got := Fra_NodesAreEquiv(p, native, expanded); got != 1 {
		t.Fatalf("XOR equivalence: %d", got)
	}
	p = satTestManager(g)
	if got := Fra_NodesAreClause(p, a, b, 0, 0); got != 0 {
		t.Fatalf("a or b is not tautological: %d", got)
	}
	p = satTestManager(g)
	if got := Fra_NodesAreClause(p, a, native, 0, 0); got != 0 {
		t.Fatalf("a or xor(a,b) is not tautological: %d", got)
	}
}
