package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestConstantCandidateSATQueries(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	gate := g.AND(x, y)
	zero := g.AND(gate, x.Not())
	p := satTestManager(g)
	if got := Fra_NodeIsConst(p, zero); got != 1 {
		t.Fatalf("unsatisfiable candidate = %d, want proof", got)
	}
	if p.NSatCalls != 1 || p.NSatProof != 1 || p.NSatCallsUnsat != 1 {
		t.Fatalf("proof counters: calls=%d proofs=%d unsat=%d", p.NSatCalls, p.NSatProof, p.NSatCallsUnsat)
	}
	p = satTestManager(g)
	if got := Fra_NodeIsConst(p, gate); got != 0 {
		t.Fatalf("nonconstant candidate = %d, want witness", got)
	}
	if p.NSatCalls != 1 || p.NSatCallsSat != 1 {
		t.Fatalf("SAT counters: calls=%d sat=%d", p.NSatCalls, p.NSatCallsSat)
	}
	if p.PPatWords[0]&3 != 3 {
		t.Fatalf("pattern does not satisfy AND: %#x", p.PPatWords[0])
	}
}
