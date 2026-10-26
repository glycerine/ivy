package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaDominatorsNativeReconvergence(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	s, _ := g.AppendAnd(n, c)
	r, _ := g.AppendAnd(n, s)
	co, err := g.AppendCO(r)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_dominators.json")
	if err != nil {
		t.Fatal(err)
	}
	var native [][]int64
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 2 {
		t.Fatal(err)
	}
	for k := 0; k < 2; k++ {
		if k == 1 {
			if _, err := g.AppendCO(n); err != nil {
				t.Fatal(err)
			}
		}
		before := slices.Clone(g.objects)
		dom, err := g.ClosestDominators(0)
		if err != nil {
			t.Fatal(err)
		}
		if len(dom) != len(native[k]) {
			t.Fatal(dom, native[k])
		}
		for i, value := range native[k] {
			expected := InvalidGiaID
			if value >= 0 {
				expected = GiaID(value)
			}
			if dom[i] != expected {
				t.Fatal("native dominator", k, i, dom, native[k])
			}
		}
		if dom[a.ID()] != n.ID() || dom[b.ID()] != n.ID() || dom[c.ID()] != s.ID() || dom[r.ID()] != co.ID() {
			t.Fatal("closest chain", dom)
		}
		if k == 0 && dom[n.ID()] != r.ID() {
			t.Fatal("reconvergence", dom)
		}
		if k == 1 && dom[n.ID()] != n.ID() {
			t.Fatal("multiple-output split", dom)
		}
		if !slices.Equal(before, g.objects) {
			t.Fatal("dominators mutated scratch")
		}
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, err := g.ClosestDominators(cap); err == nil {
			t.Fatal("invalid cap")
		}
	}
}
func TestGiaDominatorsDanglingAndNativeEdges(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	s, _ := g.AppendAnd(a, c)
	dom, err := g.ClosestDominators(0)
	if err != nil {
		t.Fatal(err)
	}
	if dom[a.ID()] != a.ID() || dom[n.ID()] != InvalidGiaID || dom[s.ID()] != InvalidGiaID {
		t.Fatal("dangling forest", dom)
	}
	h := NewGiaManager()
	control, _ := h.AppendCI()
	yes, _ := h.AppendCI()
	no, _ := h.AppendCI()
	mux, err := h.AppendMuxReal(control, yes, no)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = h.AppendCO(mux); err != nil {
		t.Fatal(err)
	}
	dom, err = h.ClosestDominators(0)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []GiaID{control.ID(), yes.ID(), no.ID()} {
		if dom[id] != mux.ID() {
			t.Fatal("native fanin dominator", dom)
		}
	}
	empty := NewGiaManager()
	dom, err = empty.ClosestDominators(0)
	if err != nil || len(dom) != 1 || dom[0] != InvalidGiaID {
		t.Fatal(dom, err)
	}
}
