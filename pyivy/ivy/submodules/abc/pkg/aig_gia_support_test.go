package abc

import (
	"slices"
	"testing"
)

func TestGiaConeLeafBoundariesAndCallIndependence(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	s, _ := g.AppendAnd(n, c)
	r, _ := g.AppendAnd(n, s)
	for _, tc := range []struct{ leaves, want []GiaID }{
		{[]GiaID{n.ID()}, []GiaID{s.ID(), r.ID()}},
		{[]GiaID{s.ID()}, []GiaID{n.ID(), r.ID()}},
		{[]GiaID{r.ID()}, nil},
		{nil, []GiaID{n.ID(), s.ID(), r.ID()}},
	} {
		got, err := g.CollectConeAnds([]GiaID{r.ID()}, tc.leaves, 0)
		if err != nil || !slices.Equal(got, tc.want) {
			t.Fatal(tc, got, err)
		}
	}
	if _, err := g.CollectConeAnds([]GiaID{r.ID()}, []GiaID{GiaID(g.ObjectCount())}, 0); err == nil {
		t.Fatal("invalid leaf")
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, err := g.CollectInputs([]GiaID{r.ID()}, cap); err == nil {
			t.Fatal("invalid support cap")
		}
	}
	if _, err := g.CollectInputs([]GiaID{GiaID(g.ObjectCount())}, 0); err == nil {
		t.Fatal("invalid root")
	}
}
func TestGiaInputSupportExtendedNodes(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	control, _ := g.AppendCI()
	mux, err := g.AppendMuxReal(control, a, b)
	if err != nil {
		t.Fatal(err)
	}
	co, err := g.AppendCO(mux)
	if err != nil {
		t.Fatal(err)
	}
	support, err := g.CollectInputs([]GiaID{co.ID(), a.ID()}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(support) != 3 {
		t.Fatal(support)
	}
	for _, id := range []GiaID{a.ID(), b.ID(), control.ID()} {
		if !slices.Contains(support, id) {
			t.Fatal("missing support", support, id)
		}
	}
	empty, err := g.CollectInputs([]GiaID{0}, 0)
	if err != nil || len(empty) != 0 {
		t.Fatal(empty, err)
	}
}
