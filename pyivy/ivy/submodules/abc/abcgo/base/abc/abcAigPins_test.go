package abc

import "testing"

func TestRecreatingAndResetsObsoleteLevel(t *testing.T) {
	n := NewNetwork("levels")
	m := Abc_AigAlloc(n)
	a, b := n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	o := n.NewObject(ObjectNode)
	o.Level, o.Copy = 57, a
	b.Level = 4
	b.Phase = true
	got := Abc_AigAndCreateFrom(m, b, a.Signal().Not(), o)
	if got.Object != o || o.Level != 5 || o.Copy != nil || !o.Phase || o.fanins[0] != a.Signal().Not() || o.fanins[1] != b.Signal() {
		t.Fatalf("recreated node level %d, phase %t, copy %p, fanins %v", o.Level, o.Phase, o.Copy, o.fanins)
	}
}

func TestRawAigCreationKeepsOrderOfEqualIDPins(t *testing.T) {
	n := NewNetwork("pins")
	m := Abc_AigAlloc(n)
	a := n.NewObject(ObjectPI).Signal()
	forms := []Edge{Abc_AigAndCreate(m, a.Not(), a), Abc_AigAndCreateFrom(m, a.Not(), a, n.NewObject(ObjectNode)), n.CreateAndEdges(a.Not(), a)}
	for _, node := range forms {
		if node.Object.fanins[0] != a.Not() || node.Object.fanins[1] != a {
			t.Errorf("ID %d lost the order of equal-ID pins", node.Object.ID)
		}
		uses := 0
		for _, fanout := range a.Object.fanouts {
			if fanout == node.Object {
				uses++
			}
		}
		if uses != 2 {
			t.Errorf("stored %d occurrences of duplicate driver", uses)
		}
	}
}

func TestAndCreationRejectsIllegalDriverBeforeMutation(t *testing.T) {
	for _, mode := range []string{"raw", "manager", "recreate"} {
		t.Run(mode, func(t *testing.T) {
			n := NewNetwork("reject")
			m := Abc_AigAlloc(n)
			a, po, o := n.NewObject(ObjectPI), n.NewObject(ObjectPO), n.NewObject(ObjectNode)
			o.Level, o.Copy = 17, po
			m.excluded[o] = true
			before := len(n.objects)
			defer func() {
				if recover() == nil {
					t.Fatal("accepted PO driver")
				}
				if len(n.objects) != before || len(a.fanouts) != 0 || len(o.fanins) != 0 || !m.excluded[o] || o.Level != 17 || o.Copy != po || o.Data != nil {
					t.Fatal("rejected call changed network or destination")
				}
			}()
			switch mode {
			case "raw":
				n.CreateAndEdges(a.Signal(), po.Signal())
			case "manager":
				Abc_AigAndCreate(m, a, po)
			case "recreate":
				Abc_AigAndCreateFrom(m, a, po, o)
			}
		})
	}
}

func BenchmarkAigAndChain(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		n := NewNetwork("chain")
		m := Abc_AigAlloc(n)
		a := n.NewObject(ObjectPI).Signal()
		root := n.NewObject(ObjectPI).Signal()
		for range 1024 {
			root = Abc_AigAnd(m, root, a)
		}
		if root.Object.Level != 1024 || len(a.Object.fanouts) != 1024 {
			b.Fatal("chain")
		}
		Abc_AigFree(m)
	}
}
