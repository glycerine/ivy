package abc

import (
	"slices"
	"sync"
	"testing"
)

func TestTopologicalOrderSkipsHolesAndStopsAtBoxOutputs(t *testing.T) {
	n := NewNetwork("postorder")
	a := n.NewObject(ObjectPI)
	hole := n.NewObject(ObjectNode)
	if err := n.DeleteObject(hole); err != nil {
		t.Fatal(err)
	}
	b := n.NewObject(ObjectBO)
	gate := n.NewObject(ObjectNode)
	// A sequential feedback edge closes a loop through the box output.
	for _, pair := range [][2]*Object{{gate, b}, {gate, a}, {b, gate}} {
		if err := n.AddFanin(pair[0], pair[1].Signal()); err != nil {
			t.Fatal(err)
		}
	}
	roots := []*Object{gate, a, gate}
	want := []*Object{b, a, gate}
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			for range 25 {
				order, err := n.Topological(roots)
				if err != nil || !slices.Equal(order, want) {
					t.Errorf("unexpected postorder %v, %v", order, err)
					return
				}
			}
		})
	}
	wg.Wait()
	for _, bad := range []*Object{hole, NewNetwork("foreign").One().Object, nil} {
		if order, err := n.Topological([]*Object{gate, bad}); err == nil || order != nil {
			t.Fatal("invalid root returned a partial order")
		}
	}
	if err := n.AddFanin(gate, gate.Signal()); err != nil {
		t.Fatal(err)
	}
	if order, err := n.Topological(roots); err == nil || order != nil {
		t.Fatal("combinational cycle returned an order")
	}
}

func TestDeleteObjectReleasesBackingReferences(t *testing.T) {
	for _, kind := range []ObjectKind{ObjectNode, ObjectPI, ObjectPO, ObjectBO, ObjectBI, ObjectLatch, ObjectBlackbox} {
		n := NewNetwork("release")
		o := n.NewObject(kind)
		byKind := n.kindOrder[kind]
		cis, cos, boxes := n.cis, n.cos, n.boxes
		if err := n.DeleteObject(o); err != nil {
			t.Fatal(err)
		}
		for _, list := range [][]*Object{byKind, cis, cos, boxes} {
			if slices.Contains(list, o) {
				t.Errorf("kind %d still occupies an obsolete slice slot", kind)
			}
		}
	}
}

func TestDeleteAndPatchFaninReleaseBackingReferences(t *testing.T) {
	n := NewNetwork("release")
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	o := n.NewObject(ObjectNode)
	if err := n.AddFanin(o, a.Signal().Not()); err != nil {
		t.Fatal(err)
	}
	oldFanouts := a.fanouts
	if err := n.PatchFanin(o, 0, b.Signal()); err != nil {
		t.Fatal(err)
	}
	if oldFanouts[0] != nil {
		t.Error("patch retained former fanout")
	}
	oldFanins, oldFanouts := o.fanins, b.fanouts
	if err := n.DeleteFanin(o, 0); err != nil {
		t.Fatal(err)
	}
	if oldFanins[0] != (Edge{}) || oldFanouts[0] != nil {
		t.Error("edge deletion retained old endpoints")
	}
}

func BenchmarkTopologicalAndChain(b *testing.B) {
	n := NewNetwork("chain")
	m := Abc_AigAlloc(n)
	defer Abc_AigFree(m)
	a, root := n.NewObject(ObjectPI).Signal(), n.NewObject(ObjectPI).Signal()
	for range 1024 {
		root = Abc_AigAnd(m, root, a)
	}
	roots := []*Object{root.Object}
	b.ReportAllocs()
	for b.Loop() {
		order, err := n.Topological(roots)
		if err != nil || len(order) != 1026 || order[len(order)-1] != root.Object {
			b.Fatal("incomplete fanin cone", err)
		}
	}
}

func TestTopologicalSparseIDsAndNilNetwork(t *testing.T) {
	var empty *Network
	if order, err := empty.Topological(nil); order != nil || err != nil {
		t.Fatal("empty root set")
	}
	n := NewNetwork("sparse")
	pi := n.NewObject(ObjectPI)
	o := n.NewObject(ObjectNode)
	if err := n.AddFanin(o, pi.Signal().Not()); err != nil {
		t.Fatal(err)
	}
	if order, err := empty.Topological([]*Object{o}); err == nil || order != nil {
		t.Fatal("nil manager accepted a foreign root")
	}
	n.objects[o.ID] = nil
	n.objects = append(n.objects, make([]*Object, 1<<17)...)
	o.ID = len(n.objects) - 1
	n.objects[o.ID] = o
	order, err := n.Topological([]*Object{o, pi})
	if err != nil || !slices.Equal(order, []*Object{pi, o}) {
		t.Fatal("sparse fanin cone", order, err)
	}
}

func TestAigReplacementRechecksCyclesAcrossGenerations(t *testing.T) {
	n := NewNetwork("replace")
	m := Abc_AigAlloc(n)
	defer Abc_AigFree(m)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	root := Abc_AigAnd(m, a, b).Object
	reject := func() {
		defer func() {
			if r := recover(); r != "cyclic AIG replacement" {
				t.Fatalf("cycle check returned %v", r)
			}
		}()
		Abc_AigReplace(m, a, root, 0)
	}
	allocs := testing.AllocsPerRun(300, reject)
	if allocs != 0 {
		t.Fatalf("warm replacement cycle check allocated %.0f times", allocs)
	}
	if len(m.pending) != 0 || !n.contains(root) || len(a.fanouts) != 1 {
		t.Fatal("cycle rejection mutated the network")
	}
}
