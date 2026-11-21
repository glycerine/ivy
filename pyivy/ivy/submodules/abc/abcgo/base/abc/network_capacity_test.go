package abc

import "testing"

func TestDeleteObjectRejectsInvalidKindBeforeMutation(t *testing.T) {
	n := NewNetwork("invalid-kind")
	o := n.NewObject(ObjectPI)
	o.Kind = ObjectKind(255)
	if err := n.DeleteObject(o); err == nil || n.Object(o.ID) != o || o.Network() != n {
		t.Fatalf("invalid deletion returned %v, live object %p, owner %p", err, n.Object(o.ID), o.Network())
	}
	o.Kind = ObjectPI
	if err := n.DeleteObject(o); err != nil {
		t.Fatal(err)
	}
}

func TestReservedAIGInventoryAndInvalidQueries(t *testing.T) {
	n := NewAIGNetworkWithCapacity("capacity", 3, 1, 2, 1<<28)
	if Abc_NtkObjNum(n) != 1 || len(n.ObjectsOfKind(ObjectNode)) != 0 || cap(n.kindOrder[ObjectNode]) > 8192 {
		t.Fatal("capacity hints created objects or reserved a full speculative gate table")
	}
	if got := n.ObjectsOfKind(ObjectKind(255)); got != nil {
		t.Fatalf("unknown kind returned %v", got)
	}
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	gate := n.AndEdges(a.Signal(), b.Signal().Not())
	if n.AndEdges(b.Signal().Not(), a.Signal()) != gate || Abc_NtkObjNum(n) != 4 || Abc_NtkNodeNum(n) != 1 {
		t.Fatal("reserved AIG changed structural hashing or object counts")
	}
	box := n.NewObject(ObjectLatch)
	box.Kind = ObjectKind(254)
	if err := n.ChangeObjectKind(box, ObjectBlackbox); err == nil {
		t.Fatal("accepted an unindexed old kind")
	}
	box.Kind = ObjectLatch
	if err := n.ChangeObjectKind(box, ObjectBlackbox); err != nil {
		t.Fatal(err)
	}
	if len(n.ObjectsOfKind(ObjectBlackbox)) != 1 || len(n.ObjectsOfKind(ObjectLatch)) != 0 {
		t.Fatal("box kind ordering")
	}
}
