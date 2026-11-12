package abc

import "testing"

func TestOriginalFaninAPI(t *testing.T) {
	n := NewNetwork("api")
	a := Abc_NtkCreatePi(n)
	b := Abc_NtkCreatePi(n)
	node := Abc_NtkCreateNode(n)
	Abc_ObjAddFanin(node, Abc_ObjNot(a))
	Abc_ObjAddFanin(node, a)
	if Abc_ObjFaninNum(node) != 2 || Abc_ObjFaninC(node, 0) != 1 || Abc_ObjFanoutNum(a) != 2 {
		t.Fatal("add fanin")
	}
	Abc_ObjPatchFanin(node, a, Abc_ObjNot(b))
	if Abc_ObjChild0(node) != b.Signal() || Abc_ObjFanoutNum(a) != 1 || Abc_ObjFanoutNum(b) != 1 {
		t.Fatal("patch polarity/multiplicity")
	}
	middle := Abc_ObjInsertBetween(a, node, ObjectNode)
	if middle == nil || Abc_ObjFanin0(middle) != a || Abc_ObjFanoutFaninNum(node, middle) != 1 {
		t.Fatal("insert")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
	Abc_ObjDeleteFanin(node, b)
	Abc_ObjRemoveFanins(node)
	if Abc_ObjFaninNum(node) != 0 || len(b.fanouts) != 0 {
		t.Fatal("remove")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}
func TestOriginalReplacementAPI(t *testing.T) {
	n := NewNetwork("api-replace")
	a := Abc_NtkCreatePi(n)
	b := Abc_NtkCreatePi(n)
	old := n.AndEdges(a.Signal(), b.Signal()).Object
	po := Abc_NtkCreatePo(n)
	Abc_ObjAddFanin(po, Abc_ObjNot(old))
	Abc_ObjReplace(old, a)
	if Abc_ObjChild0(po) != a.Signal().Not() || n.Object(old.ID) != nil {
		t.Fatal("replacement")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestTransferRejectsCycleBeforeMutation(t *testing.T) {
	n := NewNetwork("transfer")
	a := n.NewObject(ObjectPI)
	b := n.NewObject(ObjectPI)
	downstream := n.AndEdges(a.Signal(), b.Signal()).Object
	defer func() {
		if recover() == nil {
			t.Error("accepted cyclic transfer")
		}
		if err := n.Check(); err != nil {
			t.Error("mutated on failed transfer", err)
		}
	}()
	Abc_ObjTransferFanout(a, downstream)
}
