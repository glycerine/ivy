package cnf

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCutDeriveTruthRepeatsAcrossWord(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	root := g.AND(x, y)
	leaves, nodes := vec.Vec_PtrAlloc(2), vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(leaves, x)
	vec.Vec_PtrPush(leaves, y)
	vec.Vec_PtrPush(nodes, root)
	if got, want := Cnf_CutDeriveTruth(g, leaves, nodes), uint64(0x8888888888888888); got != want {
		t.Fatalf("truth = %#016x, want %#016x", got, want)
	}
}

func TestCutClauseCountMatchesEmittedClauses(t *testing.T) {
	g := aig.NewGraph()
	x, y, z := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(g.AND(x, y), z)
	leaves, nodes := vec.Vec_PtrAlloc(3), vec.Vec_PtrAlloc(1)
	for _, e := range []aig.Edge{x, y, z} {
		vec.Vec_PtrPush(leaves, e)
	}
	vec.Vec_PtrPush(nodes, root)
	mapping := vec.Vec_IntAlloc(len(g.Nodes()))
	for i := range g.Nodes() {
		vec.Vec_IntPush(mapping, int32(i+1))
	}
	clauses := vec.Vec_IntAlloc(0)
	Cnf_ComputeClauses(g, &root, leaves, nodes, mapping, nil, clauses)
	count := int32(0)
	for _, lit := range vec.Vec_IntArray(clauses) {
		if lit == 0 {
			count++
		}
	}
	if got := Cnf_CutCountClauses(g, leaves, nodes, nil); got != count {
		t.Fatalf("count = %d, emitted = %d", got, count)
	}
}
