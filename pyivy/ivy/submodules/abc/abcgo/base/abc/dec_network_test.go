package abc

import "testing"

func TestDecNetworkCountsAndSharesMarkedAnd(t *testing.T) {
	n := NewNetwork("reuse")
	a, b := n.NewObject(ObjectPI).Signal(), n.NewObject(ObjectPI).Signal()
	gate := n.AndEdges(a, b)
	d := DecNetwork{n}
	Abc_NtkIncrementTravId(n)
	Abc_NodeSetTravIdCurrent(gate.Object)
	if !d.InDeletionCone(gate.Not()) {
		t.Error("marked AND omitted from candidate cost")
	}
	if got := d.And(a, b); got != gate {
		t.Error("decomposition duplicated a structurally identical AND")
	}
	Abc_NtkIncrementTravId(n)
	gate.Object.MarkA = true
	if d.InDeletionCone(gate) {
		t.Error("old traversal mark counted in a new candidate")
	}
}

func TestDecNetworkReplacementUsesAigUpdatePath(t *testing.T) {
	n := NewNetwork("dec replacement")
	aig := Abc_AigAlloc(n)
	n.FunctionManager = aig
	a, b, c := n.NewObject(ObjectPI), n.NewObject(ObjectPI), n.NewObject(ObjectPI)
	a.SetName("a")
	b.SetName("b")
	c.SetName("c")
	ab := Abc_AigAnd(aig, a.Signal(), b.Signal())
	root := Abc_AigAnd(aig, ab, c.Signal())
	po := n.NewObject(ObjectPO)
	po.SetName("f")
	if err := n.AddFanin(po, root); err != nil {
		t.Fatal(err)
	}

	DecNetwork{Network: n}.Replace(root, ab, false)
	if n.Object(root.Object.ID) != nil {
		t.Fatal("replacement left the old AIG node alive")
	}
	if po.Fanins()[0].Object != ab.Object {
		t.Fatal("replacement did not redirect the output")
	}
	if Abc_AigCheck(aig) == 0 || Abc_NtkCheck(n) == 0 {
		t.Fatal("replacement did not preserve AIG/network integrity")
	}
}
