package framgr

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestClassAPIRefinement(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	classes := Fra_ClassesStart(graph)
	sim := newFraSimulation(graph, 0, 1, 1)
	classes.simulation = sim
	sim.NodeWords(a.NodeID())[0] = 3
	sim.NodeWords(b.NodeID())[0] = 3
	Fra_ClassesPrepare(classes, 0, 0)
	if got := Fra_ClassesCountLits(classes); got != 1 {
		t.Fatalf("initial literals = %d, want 1", got)
	}
	sim.NodeWords(b.NodeID())[0] = 5
	if got := Fra_ClassesRefine(classes); got != 1 {
		t.Fatalf("refinements = %d, want 1", got)
	}
	if got := Fra_ClassesCountLits(classes); got != 0 {
		t.Fatalf("remaining literals = %d, want 0", got)
	}
}

func TestClassAPIConstantRefinementSkipped(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	classes := Fra_ClassesStart(graph)
	sim := newFraSimulation(graph, 0, 1, 1)
	classes.simulation = sim
	Fra_ClassesPrepare(classes, 0, 0)
	sim.NodeWords(a.NodeID())[0] = 1
	sim.NodeWords(b.NodeID())[0] = 2
	var skipped int32
	if got := Fra_ClassesRefine1(classes, 0, &skipped); got != 1 || skipped != 1 {
		t.Fatalf("refinements, skipped = %d, %d; want 1, 1", got, skipped)
	}
	if got := Fra_ClassesRefine(classes); got != 1 {
		t.Fatalf("later refinements = %d, want 1", got)
	}
}

func TestClassRefinementUsesAlternatePredicates(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	classes := Fra_ClassesStart(graph)
	classes.classes = [][]int{{a.NodeID(), b.NodeID()}}
	classes.reprs[b.NodeID()] = a.NodeID()
	classes.nodeIsConst = func(int) bool { return false }
	classes.nodesEqual = func(int, int) bool { return false }
	if got := Fra_ClassesRefine(classes); got != 1 {
		t.Fatalf("alternate refinement count = %d", got)
	}
	if classes.reprs[b.NodeID()] != -1 {
		t.Fatal("alternate refinement did not remove representative")
	}
}
