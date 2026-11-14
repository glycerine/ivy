package aig

import "testing"

func TestReorderInterfacesPreservesIndicesUntilExplicitReset(t *testing.T) {
	g := NewGraph()
	a, b := g.NewInput(), g.NewInput()
	x, y := g.AddOutput(a), g.AddOutput(b.Not())
	Aig_ManSetRegNum(g, 1)
	if err := Aig_ManReorderCisCos(g, []Edge{b, a}, []Edge{y, x}, 1); err != nil {
		t.Fatal(err)
	}
	if g.Node(a).CioID != 0 || g.Node(b).CioID != 1 || g.Node(x).CioID != 0 || g.Node(y).CioID != 1 {
		t.Fatal("reordering changed stored source indices")
	}
	Aig_ManSetCioIds(g)
	for i, e := range g.Inputs() {
		if got := g.Node(e).CioID; got != int32(i) {
			t.Errorf("CI %d index=%d want=%d", e.NodeID(), got, i)
		}
	}
	for i, e := range g.Outputs() {
		if got := g.Node(e).CioID; got != int32(i) {
			t.Errorf("CO %d index=%d want=%d", e.NodeID(), got, i)
		}
	}
	values, err := g.Evaluate([]bool{true, false})
	if err != nil || len(values) != 2 || values[0] || values[1] {
		t.Fatalf("reordered outputs=%v error=%v", values, err)
	}
}
