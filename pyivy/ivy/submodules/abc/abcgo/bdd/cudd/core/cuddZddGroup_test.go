package core

import (
	"reflect"
	"testing"

	"github.com/glycerine/abcgo/bdd/mtr"
)

func TestZddNestedGroupSiftPreservesFamilyAndOrder(t *testing.T) {
	dd := NewBDDManager(4)
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to initialize ZDD variables")
	}
	f := Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 0), Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 2), Cudd_zddChange(dd, dd.One, 3)))
	Cudd_Ref(f)
	want, err := dd.CountZDD(f)
	if err != nil {
		t.Fatal(err)
	}
	parent := Cudd_MakeZddTreeNode(dd, 0, 4, mtr.MTR_DEFAULT)
	if parent == nil {
		t.Fatal("failed to create parent")
	}
	left := Cudd_MakeZddTreeNode(dd, 0, 2, mtr.MTR_DEFAULT)
	right := Cudd_MakeZddTreeNode(dd, 2, 2, mtr.MTR_DEFAULT)
	if left == nil || right == nil || parent.Child != left || left.Younger != right {
		t.Fatal("failed to create nested groups")
	}
	if Cudd_zddTreeSifting(dd, CUDD_REORDER_GROUP_SIFT) != 1 {
		t.Fatal("group sifting failed")
	}
	got, err := dd.CountZDD(f)
	if err != nil || got.Cmp(want) != 0 || !validZddOrder(dd) {
		t.Fatal("group reorder changed represented family")
	}
	if !reflect.DeepEqual([]int{int(left.Size), int(right.Size)}, []int{2, 2}) {
		t.Fatal("group sizes changed")
	}
	for _, node := range []*mtr.MtrNode{left, right} {
		lo, hi := zddFindNodeHiLo(dd, node)
		if hi-lo+1 != int(node.Size) {
			t.Fatalf("group %d not contiguous: [%d,%d]", node.Index, lo, hi)
		}
	}
}

func TestZddGroupMovementBacktracksToBestRecordedHeap(t *testing.T) {
	dd := NewBDDManager(4)
	if Cudd_zddVarsFromBddVars(dd, 1) != 1 {
		t.Fatal("failed to initialize ZDD variables")
	}
	f := Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 0), Cudd_zddUnion(dd, Cudd_zddChange(dd, dd.One, 2), Cudd_zddChange(dd, dd.One, 3)))
	Cudd_Ref(f)
	want, err := dd.CountZDD(f)
	if err != nil {
		t.Fatal(err)
	}
	state := &zddGroupState{next: []int{1, 0, 3, 2}}
	moves := []zddGroupMoveRecord{}
	initialOrder := append([]int(nil), dd.InvPermZ...)
	if zddGroupMove(dd, 1, 2, &moves, state) == 0 {
		t.Fatal("group move failed")
	}
	if len(moves) == 0 || moves[0].size == 0 {
		t.Fatal("group move was not recorded")
	}
	if !zddGroupMoveBackward(dd, moves[0].x, moves[0].y, state) {
		t.Fatal("group move backward failed")
	}
	got, err := dd.CountZDD(f)
	if err != nil || got.Cmp(want) != 0 || !reflect.DeepEqual(dd.InvPermZ, initialOrder) || !validZddOrder(dd) {
		t.Fatalf("group rollback corrupted order or family: got=%v want=%v order=%v ring=%v", got, want, dd.InvPermZ, state.next)
	}
}
