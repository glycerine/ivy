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

func TestSATImplicationAcceptsSourceSignedLimits(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	both := g.AND(a, b)
	p := satTestManager(g)
	p.PPars.NBTLimitNode = -1
	if got := Fra_NodesAreImp(p, both, a, 0, 0); got != 1 {
		t.Fatalf("source's immediately contradictory assumptions result = %d, want proof", got)
	}
}

func TestSATEquivalenceAndClause(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	native := g.NewXor(a, b)
	expanded := g.AND(a, b)
	p := satTestManager(g)
	// Source FRA CNF treats a native XOR node as AND.
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
