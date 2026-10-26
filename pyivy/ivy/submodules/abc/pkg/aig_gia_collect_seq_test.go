package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaCollectSequentialNativeRootsAndCycles(t *testing.T) {
	g := NewGiaManager()
	input, _ := g.AppendCI()
	r0, _ := g.AppendCI()
	r1, _ := g.AppendCI()
	r2, _ := g.AppendCI()
	n, err := g.AppendAnd(input, r0)
	if err != nil {
		t.Fatal(err)
	}
	s, err := g.AppendAnd(r1, r2)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range []GiaLit{n, s, r1, r0, GiaConst0} {
		if _, err = g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	if err = g.SetRegisterCount(3); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_collect_seq.json")
	if err != nil {
		t.Fatal(err)
	}
	var native [][]GiaID
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 4 {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	for i, selection := range [][]int{{0}, {1}, {0, 1}, {1, 0, 1}} {
		order, err := g.CollectSequential(selection, 0)
		if err != nil || !slices.Equal(order, native[i]) {
			t.Fatal(selection, order, native[i], err)
		}
		seen := make(map[GiaID]bool)
		for _, id := range order {
			if seen[id] {
				t.Fatal("duplicate ID")
			}
			seen[id] = true
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("source mutated")
	}
	empty, err := g.CollectSequential(nil, 0)
	if err != nil || !slices.Equal(empty, []GiaID{0}) {
		t.Fatal(empty, err)
	}
	for _, selection := range [][]int{{-1}, {2}} {
		if _, err := g.CollectSequential(selection, 0); err == nil {
			t.Fatal("invalid output")
		}
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, err := g.CollectSequential([]int{0}, cap); err == nil {
			t.Fatal("invalid cap")
		}
	}
}
func TestGiaCollectSequentialNativeMuxControls(t *testing.T) {
	g := NewGiaManager()
	input, _ := g.AppendCI()
	control, _ := g.AppendCI()
	unused, _ := g.AppendCI()
	mux, err := g.AppendMuxReal(control, input, GiaConst0)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range []GiaLit{mux, control, unused} {
		if _, err = g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	if err = g.SetRegisterCount(2); err != nil {
		t.Fatal(err)
	}
	order, err := g.CollectSequential([]int{0}, 0)
	if err != nil {
		t.Fatal(err)
	}
	ri, err := g.RegisterInputForOutput(control.ID())
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Contains(order, control.ID()) || !slices.Contains(order, ri) || slices.Contains(order, unused.ID()) {
		t.Fatal("control closure", order)
	}
}
