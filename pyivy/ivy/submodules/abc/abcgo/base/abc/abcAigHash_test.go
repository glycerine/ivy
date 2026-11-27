package abc

import "testing"

func TestAigHashPreservesHeadAcrossUnrelatedEdits(t *testing.T) {
	for _, what := range []string{"add fanin", "delete fanin", "patch fanin", "set phase", "toggle phase", "set data", "complement node"} {
		t.Run(what, func(t *testing.T) {
			n := NewNetwork("head")
			m := Abc_AigAlloc(n)
			a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
			po := n.NewObject(ObjectPO)
			if what != "add fanin" {
				Abc_ObjAddFanin(po, c)
			}
			other := Abc_AigAndCreate(m, a, c)
			first := Abc_AigAndCreate(m, a, b)
			head := Abc_AigAndCreate(m, b, a)
			if Abc_AigAndLookup(m, a, b) != head {
				t.Fatal("fixture")
			}
			switch what {
			case "add fanin":
				Abc_ObjAddFanin(po, c)
			case "delete fanin":
				if err := n.DeleteFanin(po, 0); err != nil {
					t.Fatal(err)
				}
			case "patch fanin":
				if err := n.PatchFanin(po, 0, b.Signal()); err != nil {
					t.Fatal(err)
				}
			case "set phase":
				Abc_ObjSetFaninC(po, 0)
			case "toggle phase":
				Abc_ObjXorFaninC(po, 0)
			case "set data":
				Abc_ObjSetData(other.Object, nil)
			case "complement node":
				Abc_NodeComplement(other.Object)
			}
			if got := Abc_AigAndLookup(m, b, a); got != head {
				t.Fatalf("unrelated edit returned ID %d, want head %d (oldest %d)", got.Object.ID, head.Object.ID, first.Object.ID)
			}
		})
	}
}

func TestUnindexedAigRewireInvalidatesBothAffectedKeys(t *testing.T) {
	n := NewNetwork("keys")
	a, b, c := n.NewObject(ObjectPI).Signal(), n.NewObject(ObjectPI).Signal(), n.NewObject(ObjectPI).Signal()
	first := n.CreateAndEdges(a, c)
	second := n.CreateAndEdges(a, b)
	if n.AndEdges(a, c) != first || n.AndEdges(a, b) != second {
		t.Fatal("fixture")
	}
	if err := n.PatchFanin(first.Object, 1, b); err != nil {
		t.Fatal(err)
	}
	if got, ok := n.lookupAnd(a, c); ok {
		t.Fatal("stale old pair", got.Object.ID)
	}
	if got := n.AndEdges(b, a); got != first {
		t.Fatal("target pair still points to later raw duplicate")
	}
}

func BenchmarkAigLookupAfterOutputEdit(b *testing.B) {
	n := NewNetwork("shared inputs")
	m := Abc_AigAlloc(n)
	left, right := n.NewObject(ObjectPI).Signal(), n.NewObject(ObjectPI).Signal()
	for range 512 {
		input := n.NewObject(ObjectPI).Signal()
		Abc_AigAnd(m, left, input)
		Abc_AigAnd(m, right, input)
	}
	root := Abc_AigAnd(m, left, right)
	po := n.NewObject(ObjectPO)
	Abc_ObjAddFanin(po, root)
	b.ReportAllocs()
	for b.Loop() {
		Abc_ObjXorFaninC(po, 0)
		if m.LookupAndEdges(right, left) != root {
			b.Fatal("lookup")
		}
	}
}

func TestAigHashDuplicateOrderAfterDeleteRehashAndResize(t *testing.T) {
	n := NewNetwork("duplicates")
	m := Abc_AigAlloc(n)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	first := Abc_AigAndCreate(m, a, b)
	second := Abc_AigAndCreate(m, b, a)
	third := Abc_AigAndCreate(m, a, b)
	Abc_AigAndDelete(m, third.Object)
	if got := Abc_AigAndLookup(m, a, b); got != second {
		t.Errorf("popping the head returned ID %d, want %d", got.Object.ID, second.Object.ID)
	}
	Abc_AigRehash(m)
	if got := Abc_AigAndLookup(m, b, a); got != first {
		t.Errorf("rehash produced ID %d, want %d", got.Object.ID, first.Object.ID)
	}
	Abc_AigResize(m)
	if got := Abc_AigAndLookup(m, b, a); got != second {
		t.Errorf("resize produced ID %d, want %d", got.Object.ID, second.Object.ID)
	}
	Abc_AigAndDelete(m, second.Object)
	if got := Abc_AigAndLookup(m, a, b); got != first {
		t.Error("remaining entry")
	}
}

func TestAigHashDeleteUnlinksBeforeCutCallback(t *testing.T) {
	n := NewNetwork("cut callback")
	m := Abc_AigAlloc(n)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	f := Abc_AigAndCreate(m, a, b)
	called := false
	m.FreeNodeCuts = func(id int32) {
		called = true
		if int(id) != f.Object.ID || Abc_AigAndLookup(m, a, b).Object != nil {
			t.Error("cut callback saw deleted hash entry")
		}
	}
	Abc_AigAndDelete(m, f.Object)
	if !called {
		t.Fatal("missing cut callback")
	}
}

func TestAigHashGrowthCountsRegisteredDuplicates(t *testing.T) {
	n := NewNetwork("growth")
	m := Abc_AigAlloc(n)
	m.bins = 3
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	var nodes []Edge
	for range 7 {
		nodes = append(nodes, Abc_AigAndCreate(m, a, b))
	}
	if m.bins != 3 {
		t.Fatal("grew before exceeding the threshold")
	}
	head := Abc_AigAndCreate(m, a, b)
	if m.bins != 11 {
		t.Fatalf("after eighth insertion got %d bins, want 11", m.bins)
	}
	Abc_AigAndDelete(m, head.Object)
	if Abc_AigAndLookup(m, b, a) != nodes[0] {
		t.Fatal("resize did not reverse previous bucket entries")
	}
}

func TestAigHashRecreateEarlierObjectKeepsPriorEntry(t *testing.T) {
	n := NewNetwork("recreation")
	m := Abc_AigAlloc(n)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	empty := n.NewObject(ObjectNode)
	previous := Abc_AigAndCreate(m, a, b)
	head := Abc_AigAndCreateFrom(m, b, a, empty)
	if Abc_AigAndLookup(m, a, b) != head {
		t.Fatal("recreated entry is not at the head")
	}
	Abc_AigResize(m)
	if Abc_AigAndLookup(m, a, b) != previous {
		t.Fatal("earlier object displaced its predecessor during recreation")
	}
}

func TestAigHashDataChangeRevealsRegisteredPredecessor(t *testing.T) {
	n := NewNetwork("opaque data")
	m := Abc_AigAlloc(n)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	Abc_AigAndCreate(m, a, b)
	middle := Abc_AigAndCreate(m, b, a)
	head := Abc_AigAndCreate(m, a, b)
	head.Object.Data = nil
	if got := Abc_AigAndLookup(m, a, b); got != middle {
		t.Fatal("lookup lost prior registered entry after raw function change")
	}
}

func TestAigHashUnchangedFaninAndDataKeepOrder(t *testing.T) {
	for _, patch := range []bool{false, true} {
		n := NewNetwork("unchanged")
		m := Abc_AigAlloc(n)
		a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
		Abc_AigAndCreate(m, a, b)
		head := Abc_AigAndCreate(m, b, a)
		if patch {
			if err := n.PatchFanin(head.Object, 0, head.Object.fanins[0]); err != nil {
				t.Fatal(err)
			}
		} else {
			Abc_ObjSetData(head.Object, head.Object.Data)
		}
		if Abc_AigAndLookup(m, a, b) != head {
			t.Errorf("unchanged key or function displaced head (patch %t)", patch)
		}
	}
}

func TestLookupRechecksPublicObjectKind(t *testing.T) {
	n := NewNetwork("changed kind")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	head := n.AndEdges(a.Signal(), b.Signal())
	head.Object.Kind = ObjectNet
	if _, ok := n.lookupAnd(a.Signal(), b.Signal()); ok {
		t.Fatal("cached edge no longer identifies an AND object")
	}
}

func TestAigHashMissDoesNotRegisterExcludedNodes(t *testing.T) {
	n := NewNetwork("excluded")
	m := Abc_AigAlloc(n)
	m.bins = 3
	a := n.NewObject(ObjectPI)
	var other []*Object
	var gates []Edge
	for range 7 {
		b := n.NewObject(ObjectPI)
		other = append(other, b)
		gates = append(gates, Abc_AigAndCreate(m, a, b))
	}
	for i, gate := range gates {
		Abc_AigAndDelete(m, gate.Object)
		if Abc_AigAndLookup(m, a, other[i]).Object != nil {
			t.Fatal("excluded entry")
		}
	}
	Abc_AigAnd(m, a, other[0])
	if m.bins != 3 {
		t.Fatalf("empty table grew to %d after probing excluded nodes", m.bins)
	}
}

func TestAndHashKeysKeepAllIDAndPhaseBits(t *testing.T) {
	high := int(^uint(0) >> 1)
	ids := []int{0, 1, high / 2, high - 1, high}
	seen := make(map[[2]int]bool)
	for i, x := range ids {
		for _, y := range ids[i+1:] {
			for phase := range 4 {
				a := Edge{Object: &Object{ID: x}, Complement: phase&1 != 0}
				b := Edge{Object: &Object{ID: y}, Complement: phase&2 != 0}
				key := andHashKey(edgePair(a, b))
				if seen[key] || key != andHashKey(edgePair(b, a)) {
					t.Fatalf("ambiguous key for IDs %d and %d, phases %d", x, y, phase)
				}
				seen[key] = true
			}
		}
	}
}

func TestReassignIDsRekeysDuplicateAndsAndAllPhases(t *testing.T) {
	n := Abc_NtkAlloc(ABC_NTK_STRASH, ABC_FUNC_AIG, 1)
	m := n.FunctionManager.(*Abc_Aig_t)
	hole := n.NewObject(ObjectNode)
	if err := n.DeleteObject(hole); err != nil {
		t.Fatal(err)
	}
	a, b := n.NewObject(ObjectPI).Signal(), n.NewObject(ObjectPI).Signal()
	original := a.Object.ID
	var first, last [4]Edge
	for phase := range first {
		x, y := a, b
		x.Complement, y.Complement = phase&1 != 0, phase&2 != 0
		first[phase] = Abc_AigAndCreate(m, x, y)
		last[phase] = Abc_AigAndCreate(m, x, y)
		Abc_ObjAddFanin(n.NewObject(ObjectPO), last[phase])
	}
	Abc_NtkReassignIds(n)
	if original == a.Object.ID || n.hashed.len() != 4 || n.extraAnds != 4 {
		t.Fatal("reassignment did not move inputs or rebuild all entries")
	}
	for phase := range first {
		x, y := a, b
		x.Complement, y.Complement = phase&1 != 0, phase&2 != 0
		if got := m.LookupAndEdges(y, x); got != first[phase] {
			t.Fatalf("rehash head for phases %d", phase)
		}
		Abc_AigAndDelete(m, first[phase].Object)
		if got := m.LookupAndEdges(x, y); got != last[phase] {
			t.Fatalf("rehash predecessor for phases %d", phase)
		}
	}
}
