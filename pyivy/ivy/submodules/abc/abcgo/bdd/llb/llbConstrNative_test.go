package llb

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestBDDTableMatchesComplementedAIGEdges(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	f := g.NewXor(x.Not(), y)
	out := g.AddOutput(f.Not())
	dd := core.NewBDDManager(2)
	defer core.Cudd_Quit(dd)
	vars := []*core.DdNode{core.Cudd_bddIthVar(dd, 0), core.Cudd_bddIthVar(dd, 1)}
	funcs := llbBuildBDDTable(g, dd, vars)
	root := llbBDDForEdge(g, funcs, out)
	for mask := 0; mask < 4; mask++ {
		assignment := []bool{mask&1 != 0, mask&2 != 0}
		want, err := g.Evaluate(assignment)
		if err != nil {
			t.Fatal(err)
		}
		got, err := dd.Evaluate(root, assignment)
		if err != nil {
			t.Fatal(err)
		}
		if got != want[0] {
			t.Fatalf("assignment %v: BDD=%t AIG=%t", assignment, got, want[0])
		}
	}
}

func TestSequentialConstraintRequiresBaseAndInduction(t *testing.T) {
	identity := aig.NewGraph()
	state := identity.NewInput()
	identity.AddOutput(state) // the only primary output is the bad property
	identity.AddOutput(state) // latch input preserves this state bit
	aig.Aig_ManSetRegNum(identity, 1)
	candidates := Llb_ManDeriveConstraints(identity)
	if candidates == nil {
		t.Fatal("expected the inductive state=0 constraint")
	}
	if got := vec.Vec_IntEntry(candidates, state.NodeID()); got != 0 {
		t.Fatalf("state constraint = %d, want 0", got)
	}

	flip := aig.NewGraph()
	state = flip.NewInput()
	flip.AddOutput(state)
	flip.AddOutput(state.Not()) // transition violates state=0
	aig.Aig_ManSetRegNum(flip, 1)
	if got := Llb_ManDeriveConstraints(flip); got != nil {
		t.Fatalf("non-inductive constraint survived: %v", vec.Vec_IntArray(got))
	}
}

func TestCombinationalCandidatesAreRejectedAsNonInductive(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(x, y))
	if got := Llb_ManDeriveConstraints(g); got != nil {
		t.Fatalf("unconstrained PI literals should not be inductive: %v", vec.Vec_IntArray(got))
	}
}
