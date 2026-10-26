package abc

import "testing"

func TestGiaReverseLevelMetricAndLimits(t *testing.T) {
	g := NewGiaManager()
	control, _ := g.AppendCI()
	yes, _ := g.AppendCI()
	no, _ := g.AppendCI()
	mux, err := g.AppendMuxReal(control, yes, no)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(mux); err != nil {
		t.Fatal(err)
	}
	levels, maximum, err := g.ReverseLevels(0)
	if err != nil || maximum != 2 || levels[control.ID()] != 0 || levels[yes.ID()] != 2 || levels[no.ID()] != 2 {
		t.Fatal("C reverse metric", levels, maximum, err)
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, _, err := g.ReverseLevels(cap); err == nil {
			t.Fatal("invalid cap", cap)
		}
	}
	empty := NewGiaManager()
	groups, err := empty.ReverseLevelize(0)
	if err != nil || len(groups) != 1 || len(groups[0]) != 0 {
		t.Fatal(groups, err)
	}
	// A dangling cone still propagates reverse distance in the C routine.
	h := NewGiaManager()
	a, _ := h.AppendCI()
	b, _ := h.AppendCI()
	chain := a
	for i := 0; i < 50; i++ {
		chain, err = h.AppendAnd(chain, b)
		if err != nil {
			t.Fatal(err)
		}
	}
	levels, maximum, err = h.ReverseLevels(0)
	if err != nil || maximum != 50 || levels[chain.ID()] != 0 {
		t.Fatal("dangling reverse levels", maximum, err)
	}
}
