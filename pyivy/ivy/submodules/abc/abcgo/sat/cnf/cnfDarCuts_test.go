package cnf

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
	dar "github.com/glycerine/abcgo/opt/dar/core"
	"testing"
)

func TestDarCutSlotsFollowSourceMergeOrder(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, c)
	g.AddOutput(root)
	slots, err := dar.Dar_ManComputeCutSlots(g, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	set := slots[root]
	want := [][]aig.Edge{{root}, {c, inner}, {a, b, c}}
	for i, leaves := range want {
		if !set[i].Used || len(set[i].Cut.Leaves) != len(leaves) {
			t.Fatalf("slot %d got %+v", i, set[i])
		}
		for j, e := range leaves {
			if set[i].Cut.Leaves[j] != e {
				t.Fatalf("slot %d leaves %v want %v", i, set[i].Cut.Leaves, leaves)
			}
		}
	}
}
func TestDarCutFindsConstantContradiction(t *testing.T) {
	g := aig.NewGraph()
	a, b := g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, a.Not())
	g.AddOutput(root)
	slots, err := dar.Dar_ManComputeCutSlots(g, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, slot := range slots[root] {
		if slot.Used && len(slot.Cut.Leaves) == 0 && slot.Cut.Truth == 0 {
			found = true
		}
	}
	if !found {
		t.Fatalf("constant cut not found: %+v", slots[root])
	}
}
func TestDarAreaMappingChoosesSourceCost(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, c)
	g.AddOutput(root)
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	if err := cnfApplyDarMapping(m, false); err != nil {
		t.Fatal(err)
	}
	cut := m.selectedCuts[root.NodeID()]
	if cut == nil || len(cut.pFanins) != 3 {
		t.Fatalf("wrong selected cut: %+v", cut)
	}
}

func TestAreaFlowAfterInPlaceAIGReplacement(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	root := g.AND(a.Not(), b)
	g.AddOutput(root)
	later := g.AND(a, b)
	replacement := g.AND(later, c)
	aig.Aig_ObjReplace(g, root, replacement, 0)
	if err := g.Check(); err != nil {
		t.Fatal(err)
	}
	if root.NodeID() >= later.NodeID() || g.Node(root).Kind != aig.And {
		t.Fatal("fixture must have an AND whose fanin was allocated later")
	}
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	Cnf_DeriveMapping(m)
	chosen := m.selectedCuts[root.NodeID()]
	// The three-input cut costs four clauses. Keeping the intermediate
	// gate costs three clauses at the root plus three at its sole fanin gate.
	if len(chosen.pFanins) != 3 || chosen.Cost != 4 {
		t.Fatalf("selected leaves %v at cost %d, want all three inputs at cost 4", chosen.pFanins, chosen.Cost)
	}
}

func TestAreaFlowTieWithHighFanoutInput(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, c)
	g.AddOutput(root)
	g.AddOutput(inner)
	g.AddOutput(inner)
	for i := 0; i < 2047; i++ {
		g.AddOutput(a)
	}
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	Cnf_DeriveMapping(m)
	chosen := m.selectedCuts[root.NodeID()]
	// Both cuts score 40. DAR's 11-bit Value field gives the two-input
	// cut a tie value of 4 and the three-input cut (2050 mod 2048) 2.
	if len(chosen.pFanins) != 2 || chosen.pFanins[0] != int32(c.NodeID()) || chosen.pFanins[1] != int32(inner.NodeID()) {
		t.Fatalf("selected leaves %v, want the two-input cut", chosen.pFanins)
	}
}

func TestCNFAreaFlowUsesSOPCostAndFanoutDivision(t *testing.T) {
	g := aig.NewGraph()
	a, b, c := g.NewInput(), g.NewInput(), g.NewInput()
	inner := g.AND(a, b)
	root := g.AND(inner, c)
	g.AddOutput(root)
	g.AddOutput(inner)
	m := Cnf_ManStart()
	defer Cnf_ManStop(m)
	m.pManAig = g
	area := make([]int32, aig.Aig_ManObjNumMax(g))
	area[inner.NodeID()] = 83
	cuts := aig.Aig_ComputeCuts(g, 10, 4, 1, 0)
	defer aig.Aig_ManCutStop(cuts)
	found := false
	for _, cut := range cuts.Cuts(root) {
		leaves := aig.Aig_CutLeaves(cut)
		if len(leaves) == 2 && (leaves[0] == int32(inner.NodeID()) || leaves[1] == int32(inner.NodeID())) {
			found = true
			if got := Cnf_CutAssignAreaFlow(m, cut, area); got != 71 {
				t.Errorf("area for two-input AND plus shared internal leaf = %d, want 30+83/2", got)
			}
		}
	}
	if !found {
		t.Fatal("missing two-leaf AIG cut")
	}
	super := vec.Vec_PtrAlloc(2)
	if got := Cnf_CutSuperAreaFlow(super, area, g); got != 100 {
		t.Errorf("empty supergate area %d, want 100", got)
	}
	vec.Vec_PtrPush(super, c)
	vec.Vec_PtrPush(super, inner.Not())
	if got := Cnf_CutSuperAreaFlow(super, area, g); got != 341 {
		t.Errorf("area for two-input supergate with complemented leaf = %d, want 300+83/2", got)
	}
}
