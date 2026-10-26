package abc

import (
	"slices"
	"testing"
)

func TestGiaReverseLevelOrderZeroDelayAndEmpty(t *testing.T) {
	empty := NewGiaManager()
	order, err := empty.ReverseLevelOrder(0)
	if err != nil || !slices.Equal(order, []GiaID{0}) {
		t.Fatal(order, err)
	}
	g := NewGiaManager()
	a, _ := g.AppendCI()
	buffer, err := g.AppendBuffer(a)
	if err != nil {
		t.Fatal(err)
	}
	co, err := g.AppendCO(buffer)
	if err != nil {
		t.Fatal(err)
	}
	order, err = g.ReverseLevelOrder(g.ObjectCount())
	if err != nil || !slices.Equal(order, []GiaID{co.ID(), 0, a.ID(), buffer.ID()}) {
		t.Fatal("same-level tie order", order, err)
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, err := g.ReverseLevelOrder(cap); err == nil {
			t.Fatal("invalid cap", cap)
		}
	}
}
