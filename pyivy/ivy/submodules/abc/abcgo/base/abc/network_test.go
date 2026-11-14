package abc

import "testing"

func TestNetworkAIG(t *testing.T) {
	n := NewNetwork("logic")
	a := n.NewObject(ObjectPI)
	b := n.NewObject(ObjectPI)
	a.SetName("a")
	b.SetName("b")
	and := n.AndEdges(a.Signal(), b.Signal())
	if n.AndEdges(b.Signal(), a.Signal()) != and {
		t.Fatal("structural hashing")
	}
	if n.AndEdges(a.Signal(), a.Signal()) != a.Signal() || n.AndEdges(a.Signal(), a.Signal().Not()) != n.One().Not() {
		t.Fatal("simplification")
	}
	po := n.NewObject(ObjectPO)
	if err := n.AddFanin(po, and.Not()); err != nil {
		t.Fatal(err)
	}
	for _, av := range []bool{false, true} {
		for _, bv := range []bool{false, true} {
			got, err := n.Evaluate(po.Signal(), map[*Object]bool{a: av, b: bv})
			if err != nil || got == (av && bv) {
				t.Fatal(av, bv, got, err)
			}
		}
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
	if a.Name() != "a" || len(and.Object.Fanouts()) != 1 {
		t.Fatal("name/fanout")
	}
	if err := n.AddFanin(po, a.Signal()); err == nil {
		t.Fatal("accepted second PO input")
	}
}

func TestMappingLoaderAccessorsPreserveChoiceAndCopyEdge(t *testing.T) {
	n := NewNetwork("mapping loader")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	root := n.CreateAndEdges(a.Signal(), b.Signal()).Object
	choice := n.CreateAndEdges(a.Signal().Not(), b.Signal().Not()).Object
	root.Choice = choice
	po := n.NewObject(ObjectPO)
	if err := n.AddFanin(po, root.Signal()); err != nil {
		t.Fatal(err)
	}
	if Abc_ObjChoice(root) != choice || Abc_AigNodeIsChoice(root) == 0 {
		t.Fatal("choice accessor did not preserve the alternate node")
	}
	root.SetCopyEdge(choice.Signal().Not())
	got := Abc_ObjCopyEdge(root)
	if got.Object != choice || !got.Complement {
		t.Fatalf("copy edge = %+v, want complemented choice mapping", got)
	}
}
func TestNetworkMutation(t *testing.T) {
	n := NewNetwork("mutate")
	a := n.NewObject(ObjectPI)
	b := n.NewObject(ObjectPI)
	node := n.CreateAndEdges(a.Signal(), a.Signal())
	if len(a.Fanouts()) != 2 {
		t.Fatal("lost duplicate edge")
	}
	if err := n.PatchFanin(node.Object, 1, b.Signal().Not()); err != nil {
		t.Fatal(err)
	}
	if len(a.Fanouts()) != 1 || len(b.Fanouts()) != 1 {
		t.Fatal("reciprocity")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
	if err := n.DeleteObject(a); err == nil {
		t.Fatal("deleted live fanout source")
	}
	if err := n.DeleteObject(node.Object); err != nil {
		t.Fatal(err)
	}
	if len(a.Fanouts()) != 0 || len(b.Fanouts()) != 0 || n.Object(node.Object.ID) != nil {
		t.Fatal("deletion")
	}
	foreign := NewNetwork("other").NewObject(ObjectPI)
	if err := n.AddFanin(n.NewObject(ObjectNode), foreign.Signal()); err == nil {
		t.Fatal("accepted foreign input")
	}
}
func TestNetworkCycle(t *testing.T) {
	n := NewNetwork("cycle")
	a := n.NewObject(ObjectNode)
	b := n.NewObject(ObjectNode)
	if err := n.AddFanin(a, b.Signal()); err != nil {
		t.Fatal(err)
	}
	if err := n.AddFanin(b, a.Signal()); err != nil {
		t.Fatal(err)
	}
	if err := n.Check(); err == nil {
		t.Fatal("missed cycle")
	}
}
func TestDecNetworkAdapterReplacement(t *testing.T) {
	n := NewNetwork("dec")
	a := n.NewObject(ObjectPI)
	b := n.NewObject(ObjectPI)
	d := DecNetwork{n}
	old := d.And(a.Signal(), b.Signal()).(Edge)
	po := n.NewObject(ObjectPO)
	if err := n.AddFanin(po, old.Not()); err != nil {
		t.Fatal(err)
	}
	if d.LookupAnd(a.Signal(), b.Signal()) == nil || !d.SameRegular(old, old.Not()) {
		t.Fatal("lookup")
	}
	d.Replace(old, a.Signal(), true)
	if po.Fanins()[0] != a.Signal().Not() {
		t.Fatal("complement not propagated")
	}
	if err := n.Check(); err != nil {
		t.Fatal(err)
	}
}

func TestNetworkSequentialBoundary(t *testing.T) {
	n := NewNetwork("latch")
	box := n.NewObject(ObjectLatch)
	in := n.NewObject(ObjectBI)
	out := n.NewObject(ObjectBO)
	// The latch output is a combinational input, even though its physical fanin
	// points back to the latch box. It breaks the sequential feedback traversal.
	for _, pair := range [][2]*Object{{out, box}, {in, out}, {box, in}} {
		if err := n.AddFanin(pair[0], pair[1].Signal()); err != nil {
			t.Fatal(err)
		}
	}
	if err := n.Check(); err != nil {
		t.Fatal("latch feedback is not a combinational cycle", err)
	}
	value, err := n.Evaluate(in.Signal(), map[*Object]bool{out: true})
	if err != nil || !value {
		t.Fatal(value, err)
	}
}

func TestStructuralHashInvalidatedByMutation(t *testing.T) {
	n := NewNetwork("cache")
	a := n.NewObject(ObjectPI)
	b := n.NewObject(ObjectPI)
	c := n.NewObject(ObjectPI)
	old := n.AndEdges(a.Signal(), b.Signal())
	if err := n.AddFanin(old.Object, c.Signal()); err != nil {
		t.Fatal(err)
	}
	if next := n.AndEdges(a.Signal(), b.Signal()); next == old {
		t.Fatal("mutated node returned from structural hash")
	}
}
func TestPatchRejectsOutputSource(t *testing.T) {
	n := NewNetwork("patch")
	a := n.NewObject(ObjectPI)
	po := n.NewObject(ObjectPO)
	node := n.CreateAndEdges(a.Signal(), a.Signal())
	if err := n.PatchFanin(node.Object, 0, po.Signal()); err == nil {
		t.Fatal("accepted PO as source")
	}
	if node.Object.Fanins()[0] != a.Signal() {
		t.Fatal("changed object on error")
	}
}

func TestReplaceRejectsCycleBeforeMutation(t *testing.T) {
	n := NewNetwork("replace")
	a := n.NewObject(ObjectPI)
	b := n.NewObject(ObjectPI)
	first := n.AndEdges(a.Signal(), b.Signal())
	second := n.AndEdges(first, a.Signal())
	defer func() {
		if recover() == nil {
			t.Error("accepted cyclic replacement")
		}
		if err := n.Check(); err != nil {
			t.Error("failed replacement mutated network", err)
		}
	}()
	DecNetwork{n}.Replace(first, second, true)
}

func TestStructuralHashRechecksPublicFunctionData(t *testing.T) {
	n := NewNetwork("changed-function")
	a := n.NewObject(ObjectPI)
	b := n.NewObject(ObjectPI)
	old := n.AndEdges(a.Signal(), b.Signal())
	old.Object.Data = func(v []bool) bool { return v[0] || v[1] }
	if fresh := n.AndEdges(a.Signal(), b.Signal()); fresh == old {
		t.Fatal("structural hash ignored changed function")
	}
}
