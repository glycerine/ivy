package framgr

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	"testing"
)

func TestOneHotRefinesCounterexample(t *testing.T) {
	graph := aig.NewGraph()
	a, b := graph.NewInput(), graph.NewInput()
	graph.AddOutput(a)
	graph.AddOutput(b)
	aig.Aig_ManSetRegNum(graph, 2)
	sim := newFraSimulation(graph, 1, 1, 1)
	sim.NodeWords(a.NodeID())[0] = ^uint32(0)
	sim.NodeWords(b.NodeID())[0] = ^uint32(0)
	sim.NodeWords(a.NodeID())[1] = 3
	sim.NodeWords(b.NodeID())[1] = 1
	clauses := vec.Vec_IntAlloc(4)
	vec.Vec_IntPush(clauses, Fra_RegToLit(0, 1))
	vec.Vec_IntPush(clauses, Fra_RegToLit(1, 1))
	vec.Vec_IntPush(clauses, Fra_RegToLit(1, 1))
	vec.Vec_IntPush(clauses, Fra_RegToLit(0, 0))
	p := &Fra_Man_t{pManAig: graph, pSml: sim}
	if Fra_OneHotRefineUsingCex(p, clauses) != 1 {
		t.Fatal("expected disproved clause")
	}
	if Fra_OneHotCount(p, clauses) != 1 {
		t.Fatalf("clauses = %v", vec.Vec_IntArray(clauses))
	}
	if Fra_OneHotRefineUsingCex(p, clauses) != 0 {
		t.Fatal("valid implication was removed")
	}
}
