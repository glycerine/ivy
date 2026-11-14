package abc

import "testing"

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
