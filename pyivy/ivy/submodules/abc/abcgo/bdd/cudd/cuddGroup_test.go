package cudd

import (
	"testing"

	"github.com/glycerine/abcgo/bdd/cudd/core"
	"github.com/glycerine/abcgo/bdd/mtr"
)

func TestMakeVariableGroupsUsesCurrentOrders(t *testing.T) {
	dd := NewBDDManager(4)
	if core.Cudd_ShuffleHeap(dd, []int{2, 0, 3, 1}) != 1 {
		t.Fatal("BDD shuffle failed")
	}
	g := Cudd_MakeTreeNode(dd, 0, 2, mtr.MTR_FIXED)
	if g == nil || g.Low != 1 || g.Size != 2 || g.Index != 0 {
		t.Fatalf("unexpected BDD group: %#v", g)
	}
	if Cudd_MakeTreeNode(dd, 2, 2, mtr.MTR_FIXED) != nil {
		t.Fatal("partially overlapping group should be rejected")
	}
	z := Cudd_MakeZddTreeNode(dd, 0, 2, mtr.MTR_FIXED)
	if z == nil || z.Index != 0 || z.Size != 2 {
		t.Fatalf("unexpected ZDD group: %#v", z)
	}
}
