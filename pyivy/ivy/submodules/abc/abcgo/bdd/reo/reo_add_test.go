package reo

import (
	"testing"

	cudd "github.com/glycerine/abcgo/bdd/cudd/core"
)

func addMux(dd *DdManager) *DdNode {
	left := cudd.Cudd_addIte(dd, cudd.Cudd_addIthVar(dd, 0), cudd.Cudd_addConst(dd, -7), cudd.Cudd_addConst(dd, 2.5))
	right := cudd.Cudd_addIte(dd, cudd.Cudd_addIthVar(dd, 1), cudd.Cudd_addConst(dd, 11), cudd.Cudd_addConst(dd, 2.5))
	return cudd.Cudd_addIte(dd, cudd.Cudd_addIthVar(dd, 2), right, left)
}
func checkADDMapping(t *testing.T, dd *DdManager, old, new *DdNode, inverse []int) {
	t.Helper()
	if cudd.Cudd_addPermute(dd, new, inverse) != old {
		t.Fatal("ADD inverse lost canonical identity")
	}
	for bits := 0; bits < 1<<len(dd.Vars); bits++ {
		input := make([]bool, len(dd.Vars))
		mapped := make([]bool, len(input))
		for i := range input {
			input[i] = bits&(1<<i) != 0
		}
		for i := range mapped {
			mapped[i] = input[inverse[i]]
		}
		before, err := dd.EvaluateADD(old, input)
		if err != nil {
			t.Fatal(err)
		}
		after, err := dd.EvaluateADD(new, mapped)
		if err != nil {
			t.Fatal(err)
		}
		if before != after {
			t.Fatalf("ADD terminal changed at %b: %g != %g", bits, before, after)
		}
	}
}
func TestADDReorderingNumericTerminalsAllCriteria(t *testing.T) {
	dd := cudd.NewBDDManager(4)
	root := addMux(dd)
	for _, criterion := range []Minimize{REO_MINIMIZE_NODES, REO_MINIMIZE_WIDTH, REO_MINIMIZE_APL} {
		t.Run(string(rune('0'+criterion)), func(t *testing.T) {
			p := New(4, 0)
			p.SetVerification(true)
			p.SetIterations(3)
			p.SetMinimization(criterion)
			roots, inverse, err := p.Reorder(dd, []*DdNode{root, root, cudd.Cudd_addConst(dd, -3.25)})
			if err != nil {
				t.Fatal(err)
			}
			for i, old := range []*DdNode{root, root, cudd.Cudd_addConst(dd, -3.25)} {
				checkADDMapping(t, dd, old, roots[i], inverse)
			}
			if p.criterion != criterion {
				t.Fatal("ADD objective unexpectedly changed")
			}
			if p.cost(p.stats.After) > p.cost(p.stats.Before)+REO_COST_EPSILON {
				t.Fatal("ADD cost increased", p.stats)
			}
			if criterion == REO_MINIMIZE_NODES && p.stats.After.Nodes >= p.stats.Before.Nodes {
				t.Fatal("ADD mux did not improve", p.stats)
			}
			if p.hasADD != true {
				t.Fatal("ADD classified as BDD")
			}
			if p.stats.Before.LevelNodes[len(p.slots)] != 4 {
				t.Fatal("numeric terminals merged", p.stats.Before)
			}
		})
	}
}
func TestADDTransferRemapAndShuffle(t *testing.T) {
	dd := cudd.NewBDDManager(5)
	root := cudd.Cudd_addIte(dd, cudd.Cudd_addIthVar(dd, 4), cudd.Cudd_addConst(dd, -7), cudd.Cudd_addIte(dd, cudd.Cudd_addIthVar(dd, 2), cudd.Cudd_addConst(dd, 11), cudd.Cudd_addConst(dd, 2.5)))
	p := New(5, 0)
	p.SetRemapping(true)
	p.SetIterations(0)
	roots, inverse, err := p.Reorder(dd, []*DdNode{root})
	if err != nil {
		t.Fatal(err)
	}
	checkADDMapping(t, dd, root, roots[0], inverse)
	unit := reoTransferNodesToUnits_rec(p, roots[0])
	if back := reoTransferUnitsToNodes_rec(p, unit); back != roots[0] {
		t.Fatal("ADD transfer changed terminal values")
	}
	shuffled, mapping := p.Shuffle(dd, root, []int{1, 0})
	if shuffled == nil {
		t.Fatal("ADD shuffle failed")
	}
	checkADDMapping(t, dd, root, shuffled, mapping)
}
func TestMixedADDBDDSharedRoots(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	add := addMux(dd)
	bdd := cudd.Cudd_Not(mux(dd))
	p := New(0, 0)
	p.SetVerification(true)
	roots, inverse, err := p.Reorder(dd, []*DdNode{add, bdd, dd.One, dd.LogicZero})
	if err != nil {
		t.Fatal(err)
	}
	checkADDMapping(t, dd, add, roots[0], inverse)
	checkMapping(t, dd, bdd, roots[1], inverse)
	if roots[2] != dd.One || roots[3] != dd.LogicZero {
		t.Fatal("Boolean constants changed family")
	}
	// Reconstruct through units too, where One is shared between both families.
	a := reoTransferNodesToUnits_rec(p, roots[0])
	b := reoTransferNodesToUnits_rec(p, roots[1])
	if reoTransferUnitsToNodes_rec(p, a) != roots[0] || reoTransferUnitsToNodes_rec(p, b) != roots[1] {
		t.Fatal("mixed-family transfer lost edge semantics")
	}
}

func TestADDBenchmarkUsesNonzeroPattern(t *testing.T) {
	dd := cudd.NewBDDManager(3)
	root := addMux(dd) // Every leaf is nonzero; the pattern is constant one.
	if nodes := Extra_addReorderTest(dd, root); nodes != 1 {
		t.Fatalf("nonzero pattern has %d nodes, want 1", nodes)
	}
	root = cudd.Cudd_addIte(dd, cudd.Cudd_addIthVar(dd, 0), cudd.Cudd_addConst(dd, -4), cudd.Cudd_addConst(dd, 0))
	if nodes := Extra_addReorderTest(dd, root); nodes != 3 {
		t.Fatalf("binary pattern has %d nodes, want 3", nodes)
	}
}
