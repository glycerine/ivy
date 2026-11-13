package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestImplicationRefinementIgnoresPrefix(t *testing.T) {
	graph := aig.NewGraph()
	a, b, c := graph.NewInput(), graph.NewInput(), graph.NewInput()
	sim := newFraSimulation(graph, 1, 1, 1)
	sim.NodeWords(a.NodeID())[0], sim.NodeWords(b.NodeID())[0] = 1, 0
	sim.NodeWords(a.NodeID())[1], sim.NodeWords(b.NodeID())[1] = 3, 1
	sim.NodeWords(c.NodeID())[1] = 7
	imps := vec.Vec_IntAlloc(2)
	vec.Vec_IntPush(imps, Fra_ImpCreate(int32(a.NodeID()), int32(b.NodeID())))
	keep := Fra_ImpCreate(int32(a.NodeID()), int32(c.NodeID()))
	vec.Vec_IntPush(imps, keep)
	p := &Fra_Man_t{PManAig: graph, PSml: sim}
	if got := Fra_ImpRefineUsingCex(p, imps); got != 1 {
		t.Fatalf("changed = %d", got)
	}
	Fra_ImpCompactArray(imps)
	if vec.Vec_IntSize(imps) != 1 || vec.Vec_IntEntry(imps, 0) != keep {
		t.Fatalf("remaining implications: %v", vec.Vec_IntArray(imps))
	}
	if Fra_ImpRefineUsingCex(p, imps) != 0 {
		t.Fatal("valid implication was removed")
	}
}
