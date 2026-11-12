package abc

import "testing"

func TestNetworkScratchCopiesAndMarks(t *testing.T) {
	n := NewNetwork("scratch")
	input := n.NewObject(ObjectPI)
	node := n.NewObject(ObjectNode)
	setCopy(node, input.Signal().Not())
	input.Next = node
	node.MarkA, node.MarkB, node.MarkC = true, true, true
	saved := Abc_NtkSaveCopy(n)
	Abc_NtkCleanCopy(n)
	if Abc_NtkCountCopy(n) != 0 {
		t.Fatal("copy was not cleared")
	}
	Abc_NtkLoadCopy(n, saved)
	if Abc_NtkCountCopy(n) != 1 || getCopy(node) != input.Signal().Not() {
		t.Fatal("complemented copy was not restored")
	}
	Abc_NtkCleanMarkABC(n)
	Abc_NtkCleanNext(n)
	Abc_NtkFillTemp(n)
	if node.MarkA || node.MarkB || node.MarkC || input.Next != nil || input.Temp != -1 || node.Temp != -1 {
		t.Fatal("scratch fields not reset")
	}
}

func TestFanoutQueriesRespectPolarityAndNames(t *testing.T) {
	n := NewNetwork("fanouts")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	node := n.AndEdges(a.Signal(), b.Signal()).Object
	child := n.AndEdges(node.Signal(), a.Signal()).Object
	out1, out2 := n.NewObject(ObjectPO), n.NewObject(ObjectPO)
	out1.SetName("same")
	out2.SetName("same")
	Abc_ObjAddFanin(out1, node)
	Abc_ObjAddFanin(out2, node.Signal().Not())
	if Abc_NodeFindFanin(node, a) < 0 || Abc_NodeFindCoFanout(node) != out1 || Abc_NodeFindNonCoFanout(node) != child || Abc_NodeHasUniqueCoFanout(node) != out1 {
		t.Fatal("incorrect fanin or fanout query")
	}
	out3 := n.NewObject(ObjectPO)
	out3.SetName("different")
	Abc_ObjAddFanin(out3, node)
	if Abc_NodeHasUniqueCoFanout(node) != nil {
		t.Fatal("different output names should not be unique")
	}
}

func TestTransferCopyComposesEdgePolarity(t *testing.T) {
	src, middle, dst := NewNetwork("src"), NewNetwork("middle"), NewNetwork("dst")
	a, b, c := src.NewObject(ObjectPI), middle.NewObject(ObjectPI), dst.NewObject(ObjectPI)
	setCopy(a, b.Signal().Not())
	setCopy(b, c.Signal().Not())
	Abc_NtkTransferCopy(src)
	if getCopy(a) != c.Signal() {
		t.Fatal("copy composition did not cancel complemented edges")
	}
	setCopy(a, b.Signal().Not())
	Abc_NtkCleanCopy(middle)
	Abc_NtkTransferCopy(src)
	if getCopy(a).Object != nil {
		t.Fatal("copy through an unmapped object should be empty")
	}
}

func TestCrossCutSharedNode(t *testing.T) {
	n := NewNetwork("crosscut")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	shared := n.AndEdges(a.Signal(), b.Signal()).Object
	for range 2 {
		Abc_ObjAddFanin(n.NewObject(ObjectPO), shared)
	}
	setCopy(shared, a.Signal().Not())
	if got := Abc_NtkCrossCut(n); got != 2 {
		t.Fatalf("cross-cut = %d, want 2", got)
	}
	if getCopy(shared) != a.Signal().Not() {
		t.Fatal("cross-cut changed copy mapping")
	}
}
