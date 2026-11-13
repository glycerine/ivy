package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/sat/bsat"
	"testing"
)

// The three two-node source queries simplify queued root assignments after
// adding their CNF, whereas Fra_NodeIsConst omits that simplification.
func TestSATPreparationSourceRootPropagation(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	p := satTestManager(g)
	fraPrepareSAT(p, a, b)
	v := p.PSat.NewVariable()
	if !p.PSat.AddClause(bsat.Positive(v)) || !p.PSat.NeedsPropagation() {
		t.Fatal("fixture lacks queued root unit")
	}
	fraPrepareSAT(p, a, b)
	if p.PSat.NeedsPropagation() {
		t.Fatal("two-node preparation did not simplify queued root unit")
	}
	v = p.PSat.NewVariable()
	if !p.PSat.AddClause(bsat.Positive(v)) || !p.PSat.NeedsPropagation() {
		t.Fatal("fixture lacks second queued root unit")
	}
	fraPrepareSAT(p, aig.InvalidEdge, a)
	if !p.PSat.NeedsPropagation() {
		t.Fatal("constant-node preparation should preserve the queued unit")
	}
}

func TestSATPreparationRejectsRootConflict(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	p := satTestManager(g)
	fraPrepareSAT(p, a, b)
	x, y := p.PSat.NewVariable(), p.PSat.NewVariable()
	if !p.PSat.AddClause(bsat.Positive(x), bsat.Positive(y)) || !p.PSat.AddClause(bsat.Negative(x)) || !p.PSat.AddClause(bsat.Negative(y)) {
		t.Fatal("fixture must leave a pending root conflict")
	}
	defer func() {
		if got := recover(); got != "fra: SAT root simplification failed" {
			t.Fatalf("panic = %v, want root simplification assertion", got)
		}
	}()
	fraPrepareSAT(p, a, b)
}
