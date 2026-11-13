package cnf

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestCollectLeavesEntryPointsAgree(t *testing.T) {
	g := aig.NewGraph()
	x, y, z := g.NewInput(), g.NewInput(), g.NewInput()
	shared := g.AND(x, y)
	aig.Aig_ObjSetMarkA(g, shared)
	root := g.AND(shared.Not(), z)
	for _, stop := range []bool{false, true} {
		leaves := vec.Vec_PtrAlloc(0)
		var flag int32
		if stop {
			flag = 1
		}
		Cnf_CollectLeaves(g, &root, leaves, flag)
		var want []aig.Edge
		for _, item := range vec.Vec_PtrArray(leaves) {
			want = append(want, item.(aig.Edge))
		}
		if got := CollectAIGLeaves(g, root, stop); !reflect.DeepEqual(got, want) {
			t.Fatalf("stop=%v got=%v want=%v", stop, got, want)
		}
	}
}

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

func TestLargeAndCutClauses(t *testing.T) {
	g := aig.NewGraph()
	leaves, nodes := vec.Vec_PtrAlloc(7), vec.Vec_PtrAlloc(1)
	var inputs []aig.Edge
	for i := 0; i < 7; i++ {
		e := g.NewInput()
		inputs = append(inputs, e)
		vec.Vec_PtrPush(leaves, e)
	}
	root := inputs[0]
	for i, e := range inputs[1:] {
		if i == 2 {
			e = e.Not()
		}
		root = g.AND(root, e)
	}
	vec.Vec_PtrPush(nodes, root)
	clauses := vec.Vec_IntAlloc(0)
	Cnf_ComputeClauses(g, &root, leaves, nodes, nil, nil, clauses)
	if got := Cnf_CutCountClauses(g, leaves, nodes, nil); got != 8 {
		t.Fatalf("clause count %d", got)
	}
	for assignment := 0; assignment < 256; assignment++ {
		valid, clause, started := true, false, false
		for _, lit := range vec.Vec_IntArray(clauses) {
			if lit == 0 {
				if started && !clause {
					valid = false
				}
				clause, started = false, true
				continue
			}
			var value bool
			if int(lit>>1) == root.NodeID() {
				value = assignment&128 != 0
			} else {
				value = assignment&(1<<(int(lit>>1)-1)) != 0
			}
			if value != (lit&1 != 0) {
				clause = true
			}
		}
		valid = valid && started && clause
		wantOutput := assignment&0x7f == 0x77
		if valid != ((assignment&128 != 0) == wantOutput) {
			t.Fatalf("assignment %08b accepted=%v", assignment, valid)
		}
	}
}

func TestCutClausesComplementedLeafAndReuse(t *testing.T) {
	g := aig.NewGraph()
	x, y := g.NewInput(), g.NewInput()
	root := g.AND(x, y)
	leaves, nodes := vec.Vec_PtrAlloc(2), vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(leaves, x.Not())
	vec.Vec_PtrPush(leaves, y)
	vec.Vec_PtrPush(nodes, root)
	clauses := vec.Vec_IntAlloc(0)
	Cnf_ComputeClauses(g, &root, leaves, nodes, nil, nil, clauses)
	Cnf_ComputeClauses(g, &root, leaves, nodes, nil, nil, clauses)
	count := 0
	for _, lit := range vec.Vec_IntArray(clauses) {
		if lit == 0 {
			count++
		}
	}
	if count != 4 {
		t.Fatalf("reused output contains %d clauses, want 4", count)
	}
	for assignment := 0; assignment < 8; assignment++ {
		valid, clause, started := true, false, false
		for _, lit := range vec.Vec_IntArray(clauses) {
			if lit == 0 {
				if started && !clause {
					valid = false
				}
				started, clause = true, false
				continue
			}
			variable := int(lit >> 1)
			bit := variable - 1
			if variable == root.NodeID() {
				bit = 2
			}
			if (assignment&(1<<bit) != 0) != (lit&1 != 0) {
				clause = true
			}
		}
		valid = valid && clause
		want := (assignment&4 != 0) == (assignment&1 != 0 && assignment&2 != 0)
		if valid != want {
			t.Fatalf("assignment %03b accepted=%v want=%v", assignment, valid, want)
		}
	}
}
