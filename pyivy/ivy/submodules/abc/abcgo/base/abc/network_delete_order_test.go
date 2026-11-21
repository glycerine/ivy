package abc

import "testing"

func TestDeleteInterleavedNodesPreservesEnumeration(t *testing.T) {
	n := NewNetwork("delete")
	var nodes, live []*Object
	for i := range 1537 {
		if i%3 == 0 {
			n.NewObject(ObjectPI)
		}
		nodes = append(nodes, n.NewObject(ObjectNode))
	}
	for pass := range 3 {
		for i := pass; i < len(nodes); i += 4 {
			if err := n.DeleteObject(nodes[i]); err != nil {
				t.Fatal(err)
			}
		}
		live = live[:0]
		for i, o := range nodes {
			if i%4 > pass {
				live = append(live, o)
			}
		}
		got := n.ObjectsOfKind(ObjectNode)
		if len(got) != len(live) || int(Abc_NtkNodeNum(n)) != len(live) || int(Abc_NtkObjNum(n)) != 1+513+len(live) {
			t.Fatalf("pass %d: nodes %d/%d, total %d", pass, len(got), len(live), Abc_NtkObjNum(n))
		}
		for i := range got {
			if got[i] != live[i] {
				t.Fatalf("pass %d: position %d has ID %d, want %d", pass, i, got[i].ID, live[i].ID)
			}
		}
		if !checkObjectInventories(n) || !checkKindOrder(n, ObjectNode) {
			t.Fatalf("pass %d: inconsistent kind index", pass)
		}
	}
	extra := n.NewObject(ObjectNode)
	if got := n.ObjectsOfKind(ObjectNode); got[len(got)-1] != extra {
		t.Fatal("new node lost creation order")
	}
}

func BenchmarkDeleteAlternateNodes(b *testing.B) {
	const count = 4096
	b.ReportAllocs()
	for b.Loop() {
		n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
		for range count {
			Abc_NtkCreateNode(n)
		}
		for id := 1; id < count; id += 2 {
			Abc_NtkDeleteObj(n.Object(id))
		}
		nodes := n.ObjectsOfKind(ObjectNode)
		if Abc_NtkNodeNum(n) != count/2 || len(nodes) != count/2 {
			b.Fatal("unexpected number of surviving nodes")
		}
		for i, o := range nodes {
			if o.ID != 2*(i+1) {
				b.Fatal("unexpected node order")
			}
		}
		Abc_NtkDelete(n)
	}
}

func TestNodeDirectoryAfterIDReassignment(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	m := n.FunctionManager.(*Abc_Aig_t)
	a, b := n.NewObject(ObjectPI).Signal(), n.NewObject(ObjectPI).Signal()
	ab := Abc_AigAnd(m, a, b)
	removed := Abc_AigAnd(m, a.Not(), b)
	later := Abc_AigAnd(m, a, b.Not())
	last := Abc_AigAnd(m, a.Not(), b.Not())
	if err := n.AddFanin(n.NewObject(ObjectPO), last); err != nil {
		t.Fatal(err)
	}
	Abc_AigDeleteNode(m, removed.Object)
	snapshot := n.ObjectsOfKind(ObjectNode)
	Abc_NtkReassignIds(n)
	if !checkObjectInventories(n) || Abc_AigCheck(m) == 0 {
		t.Fatal("reassigned network has inconsistent node inventory")
	}
	Abc_AigDeleteNode(m, later.Object)
	got := n.ObjectsOfKind(ObjectNode)
	if len(got) != 2 || got[0].ID >= got[1].ID || (got[0] != ab.Object && got[1] != ab.Object) || (got[0] != last.Object && got[1] != last.Object) {
		t.Fatal("deleting a reassigned ID lost the wrong gate or changed enumeration order")
	}
	if len(snapshot) != 3 || snapshot[0] != ab.Object || snapshot[1] != later.Object || snapshot[2] != last.Object {
		t.Fatal("mutations overwrote the earlier node snapshot")
	}
}
