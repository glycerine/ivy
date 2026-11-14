package abci

import (
	"testing"

	"github.com/glycerine/abcgo/base/abc"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestMffcRecursiveFanoutCollectionAndReferenceRestore(t *testing.T) {
	n := abc.NewNetwork("mffc-recursive")
	a, b, c, d := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	shared := n.CreateAndEdges(a.Signal(), b.Signal())
	left := n.CreateAndEdges(shared, c.Signal())
	right := n.CreateAndEdges(shared, d.Signal())
	root := n.CreateAndEdges(left, right)
	po := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(po, root.Object)

	nodes, leaves := Vec_PtrAlloc(8), Vec_PtrAlloc(8)
	Abc_MffcCollectNodes([]*Abc_Obj_t{root.Object}, nodes)
	if got := Vec_PtrSize(nodes); got != 4 {
		t.Fatalf("MFFC size %d, want root, both branches, and shared node", got)
	}
	Abc_MffcCollectLeaves(nodes, leaves)
	if got := Vec_PtrSize(leaves); got != 4 {
		t.Fatalf("MFFC leaf count %d, want four primary inputs", got)
	}
	for _, input := range []*abc.Object{a, b, c, d} {
		found := false
		for _, item := range vec.Vec_PtrArray(leaves) {
			found = found || item == input
		}
		if !found {
			t.Fatalf("missing leaf %d", input.ID)
		}
	}
	for _, input := range []*abc.Object{a, b, c, d} {
		if got := abc.Abc_ObjFanoutNum(input); got != 1 {
			t.Fatalf("PI %d fanout count restored as %d, want 1", input.ID, got)
		}
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestMffcMarkRootsRespectsPrimaryInputOption(t *testing.T) {
	n := abc.NewNetwork("mffc-roots")
	a, b := n.NewObject(abc.ObjectPI), n.NewObject(abc.ObjectPI)
	root := n.CreateAndEdges(a.Signal(), b.Signal())
	po := n.NewObject(abc.ObjectPO)
	abc.Abc_ObjAddFanin(po, root.Object)

	internal := Abc_NktMffcMarkRoots(n, 1)
	if Vec_PtrSize(internal) != 1 || vec.Vec_PtrEntry(internal, 0) != root.Object {
		t.Fatalf("skip-PIs roots = %v, want only output driver", vec.Vec_PtrArray(internal))
	}
	all := Abc_NktMffcMarkRoots(n, 0)
	if Vec_PtrSize(all) != 3 {
		t.Fatalf("expanded roots count = %d, want root plus two PI leaves", Vec_PtrSize(all))
	}
	if a.MarkA || b.MarkA || root.Object.MarkA {
		t.Fatal("root traversal left temporary marks set")
	}
}

func TestMffcCostAndSupportMerge(t *testing.T) {
	a, b := vec.Vec_IntAlloc(4), vec.Vec_IntAlloc(4)
	for _, id := range []int32{1, 3, 8} {
		Vec_IntPush(a, id)
	}
	for _, id := range []int32{3, 5} {
		Vec_IntPush(b, id)
	}
	if got := Abc_NktMffcCostTwo(a, b, 4, 4); got == -float64(^uint32(0)>>1) {
		t.Fatal("support union of size four should fit")
	}
	if got := Abc_NktMffcCostTwo(a, b, 4, 3); got != -float64(^uint32(0)>>1) {
		t.Fatalf("over-limit cost = %v", got)
	}
}
