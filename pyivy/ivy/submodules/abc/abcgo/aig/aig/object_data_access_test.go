package aig

import "testing"

func TestObjectDataAccessUsesRegularNode(t *testing.T) {
	g := NewGraph()
	x := g.NewInput()
	g.SetObjectData(x, "payload")
	if g.ObjectData(x.Not()) != "payload" {
		t.Fatal("complemented edge did not refer to the same node")
	}
	g.SetObjectData(x.Not(), nil)
	if g.ObjectData(x) != nil {
		t.Fatal("payload was not cleared")
	}
}
