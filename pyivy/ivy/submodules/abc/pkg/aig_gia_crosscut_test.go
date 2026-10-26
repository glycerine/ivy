package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaCrossCutNativeOrderAndSize(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, err := g.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	s, err := g.AppendAnd(n, c)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range []GiaLit{n, s, n, GiaConst0} {
		if _, err := g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_crosscut.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []struct {
		Order   []GiaID
		Maximum int
	}
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 2 {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	for reverse := 0; reverse < 2; reverse++ {
		order, err := g.DepthFirstCrossCutOrder(reverse != 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		size, err := g.CrossCut(reverse != 0, 0)
		if err != nil || size != native[reverse].Maximum || !slices.Equal(order, native[reverse].Order) {
			t.Fatal(order, size, native[reverse], err)
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("cross-cut mutated source")
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, err := g.CrossCut(false, cap); err == nil {
			t.Fatal("invalid limit")
		}
	}
}
func TestGiaCrossCutExtendedNodesAndDeepCones(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	xor, err := g.AppendXorReal(a, b)
	if err != nil {
		t.Fatal(err)
	}
	mux, err := g.AppendMuxReal(c, xor, GiaConst0)
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := g.AppendBuffer(mux)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(buffer); err != nil {
		t.Fatal(err)
	}
	order, err := g.DepthFirstCrossCutOrder(false, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(order, c.ID()) || !slices.Contains(order, GiaID(0)) {
		t.Fatal("extended fanins omitted", order)
	}
	size, err := g.CrossCut(false, 0)
	if err != nil || size < 3 {
		t.Fatal(size, err)
	}
	empty := NewGiaManager()
	if _, err := empty.AppendCO(GiaConst1); err != nil {
		t.Fatal(err)
	}
	size, err = empty.CrossCut(false, 0)
	if err != nil || size != 0 {
		t.Fatal("constant CO", size, err)
	}
	h := NewGiaManager()
	x, _ := h.AppendCI()
	y, _ := h.AppendCI()
	chain := x
	for i := 0; i < 10000; i++ {
		chain, err = h.AppendAnd(chain, y)
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err = h.AppendCO(chain); err != nil {
		t.Fatal(err)
	}
	size, err = h.CrossCut(false, 0)
	if err != nil || size != 3 {
		t.Fatal("deep iterative cone", size, err)
	}
}
