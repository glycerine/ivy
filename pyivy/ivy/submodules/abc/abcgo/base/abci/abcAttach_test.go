package abci

import (
	"github.com/glycerine/abcgo/base/abc"
	"testing"
)

func TestAttachTransferFaninsPreservesEdgesAndCounts(t *testing.T) {
	n := abc.Abc_NtkAlloc(abc.ABC_NTK_LOGIC, abc.ABC_FUNC_SOP, 0)
	a, b, c := abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n), abc.Abc_NtkCreatePi(n)
	node := abc.Abc_NtkCreateNode(n)
	for _, input := range []*abc.Object{a, b, c} {
		if err := n.AddFanin(node, input.Signal()); err != nil {
			t.Fatal(err)
		}
	}
	old := []abc.Edge{a.Signal().Not(), b.Signal(), c.Signal()}
	perm := []int32{2, 0, 1}
	Abc_AttachTransferFanins(n, node, old, perm, 3)
	got := node.Fanins()
	if len(got) != 3 {
		t.Fatalf("fanin count=%d, want 3", len(got))
	}
	for i, want := range []abc.Edge{old[1], old[2], old[0]} {
		if got[i] != want {
			t.Errorf("fanin[%d]=%v, want %v", i, got[i], want)
		}
	}
	if abc.Abc_ObjFanoutNum(a) != 1 || abc.Abc_ObjFanoutNum(b) != 1 || abc.Abc_ObjFanoutNum(c) != 1 {
		t.Fatal("fanin transfer changed input fanout counts")
	}
}

func TestAttachTruthTableSetupAndCount(t *testing.T) {
	var truths [6][2]uint32
	Abc_AttachSetupTruthTables(truths[:])
	var result [2]uint32
	Abc_AttachComputeTruth("11 1\n", truths[:], result[:])
	if result[0] != 0x88888888 || result[1] != 0x88888888 {
		t.Fatalf("AND truth=%08x:%08x", result[1], result[0])
	}
	Abc_AttachComputeTruth("0 1\n", truths[:], result[:])
	if result[0] != 0x55555555 || result[1] != 0x55555555 {
		t.Fatalf("inverter truth=%08x:%08x", result[1], result[0])
	}
}
