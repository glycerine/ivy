package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaFanoutsNativeIndexAndPostorder(t *testing.T) {
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
	r, err := g.AppendAnd(n, s)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range []GiaLit{r, n} {
		if _, err = g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_fanouts.json")
	if err != nil {
		t.Fatal(err)
	}
	var native struct{ Fanouts, Orders [][]GiaID }
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	if len(native.Fanouts) != g.ObjectCount() || len(native.Orders) != 3 {
		t.Fatal("fixture dimensions")
	}
	before := slices.Clone(g.objects)
	index, err := g.StaticFanouts(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	for id, want := range native.Fanouts {
		got, err := index.Fanouts(GiaID(id))
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(id, got, want, err)
		}
	}
	for i, roots := range [][]GiaID{{n.ID()}, {s.ID()}, {s.ID(), n.ID()}} {
		got, err := g.CollectTransitiveFanouts(roots, 0, 0)
		if err != nil || !slices.Equal(got, native.Orders[i]) {
			t.Fatal(roots, got, native.Orders[i], err)
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("fanouts mutated graph")
	}
	owned, err := index.Fanouts(n.ID())
	if err != nil {
		t.Fatal(err)
	}
	owned[0] = 99
	again, err := index.Fanouts(n.ID())
	if err != nil || again[0] == 99 {
		t.Fatal("snapshot exposes storage")
	}
	if _, err = g.AppendCO(n); err != nil {
		t.Fatal(err)
	}
	snapshot, err := index.Fanouts(n.ID())
	if err != nil || !slices.Equal(snapshot, native.Fanouts[n.ID()]) {
		t.Fatal("snapshot changed after append")
	}
	for _, limits := range [][2]int{{-1, 0}, {0, -1}, {1, 0}, {0, 1}} {
		if _, err := g.StaticFanouts(limits[0], limits[1]); err == nil {
			t.Fatal("limit ignored", limits)
		}
	}
	if _, err := index.Fanouts(GiaID(g.ObjectCount())); err == nil {
		t.Fatal("invalid ID")
	}
	if _, err := g.CollectTransitiveFanouts([]GiaID{GiaID(g.ObjectCount())}, 0, 0); err == nil {
		t.Fatal("invalid root")
	}
}
func TestGiaFanoutsExtendedEdgesAndBoundaryRoots(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	control, _ := g.AppendCI()
	mux, err := g.AppendMuxReal(control, a, b)
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := g.AppendBuffer(mux)
	if err != nil {
		t.Fatal(err)
	}
	co, err := g.AppendCO(buffer)
	if err != nil {
		t.Fatal(err)
	}
	index, err := g.StaticFanouts(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	got, err := index.Fanouts(a.ID())
	if err != nil || !slices.Equal(got, []GiaID{mux.ID()}) {
		t.Fatal("data edge", got, err)
	}
	got, err = g.CollectTransitiveFanouts([]GiaID{control.ID(), control.ID()}, 0, 0)
	if err != nil || !slices.Equal(got, []GiaID{buffer.ID(), mux.ID()}) {
		t.Fatal("extended traversal", got, err)
	}
	got, err = g.CollectTransitiveFanouts([]GiaID{co.ID()}, 0, 0)
	if err != nil || len(got) != 0 {
		t.Fatal("CO boundary", got, err)
	}
}
