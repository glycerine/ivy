package cnf

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/opt/dar"
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
	if err := cnfApplyDarMapping(m); err != nil {
		t.Fatal(err)
	}
	cut := m.selectedCuts[root.NodeID()]
	if cut == nil || len(cut.pFanins) != 3 {
		t.Fatalf("wrong selected cut: %+v", cut)
	}
}
