package dch

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestClassesLitNumExcludesConstantCandidates(t *testing.T) {
	// C Dch_ClassesLitNum returns nLits, which excludes nCands1.
	classes := &Classes{groups: [][]aig.Edge{{0, 2, 4}, {6, 8, 10}, {12}}}
	if got := Dch_ClassesLitNum(classes); got != 2 {
		t.Fatalf("ordinary class literals = %d, want 2", got)
	}
}

func TestSATCallCountIsPerCandidate(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	x, y := g.AND(g.AND(a, b), c), g.AND(a, g.AND(b, c))
	m, err := newManager(g, nil)
	if err != nil {
		t.Fatal(err)
	}
	result, _, err := m.provePair(x, y)
	if err != nil || result != proofEquivalent {
		t.Fatalf("result %v, error %v", result, err)
	}
	if got := m.stats.SATCalls; got != 1 {
		t.Fatalf("SAT calls %d, want one candidate query", got)
	}
}
