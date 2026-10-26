package abc

import (
	"slices"
	"testing"
)

func TestGiaCombinedNodesAndAllGates(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	dangling, err := g.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := g.AppendBuffer(a)
	if err != nil {
		t.Fatal(err)
	}
	xor, err := g.AppendXorReal(a, b)
	if err != nil {
		t.Fatal(err)
	}
	co, err := g.AppendCO(buffer)
	if err != nil {
		t.Fatal(err)
	}
	all, err := g.AndObjects(0)
	if err != nil || !slices.Equal(all, []GiaID{dangling.ID(), buffer.ID(), xor.ID()}) {
		t.Fatal(all, err)
	}
	combined, err := g.CollectNodesAndInputs([]GiaID{co.ID(), a.ID(), 0}, 0)
	if err != nil || !slices.Equal(combined, []GiaID{a.ID(), buffer.ID()}) {
		t.Fatal(combined, err)
	}
	empty, err := g.CollectNodesAndInputs([]GiaID{0}, 0)
	if err != nil || len(empty) != 0 {
		t.Fatal(empty, err)
	}
	empty, err = g.CollectNodesAndInputs(nil, 0)
	if err != nil || len(empty) != 0 {
		t.Fatal(empty, err)
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, err := g.AndObjects(cap); err == nil {
			t.Fatal("all gates cap")
		}
		if _, err := g.CollectNodesAndInputs([]GiaID{co.ID()}, cap); err == nil {
			t.Fatal("combined cap")
		}
	}
	if _, err := g.CollectNodesAndInputs([]GiaID{GiaID(g.ObjectCount())}, 0); err == nil {
		t.Fatal("invalid combined root")
	}
}
