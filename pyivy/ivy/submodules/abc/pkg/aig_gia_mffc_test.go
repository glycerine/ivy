package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaMffcNativeSharedBoundary(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	s, _ := g.AppendAnd(n, c)
	r, _ := g.AppendAnd(n, s)
	if _, err := g.AppendCO(r); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_mffc.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []struct {
		Size     int
		Boundary []GiaID
		Deltas   []int64
	}
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
		got, err := g.Mffc(r.ID(), 0)
		if err != nil || got.Size != native[k].Size || !slices.Equal(got.Boundary, native[k].Boundary) || !slices.Equal(got.BoundaryReferenceDeltas, native[k].Deltas) {
			t.Fatal(got, native[k], err)
		}
		expected := 3
		if k == 1 {
			expected = 2
		}
		if got.Size != expected {
			t.Fatal("shared boundary size", got)
		}
		if !slices.Equal(before, g.objects) {
			t.Fatal("MFFC mutated graph")
		}
	}
	for _, root := range []GiaID{0, a.ID(), GiaID(g.ObjectCount())} {
		if _, err := g.Mffc(root, 0); err == nil {
			t.Fatal("invalid MFFC root")
		}
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, err := g.Mffc(r.ID(), cap); err == nil {
			t.Fatal("invalid cap")
		}
	}
}
func TestGiaMffcExtendedEdges(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	control, _ := g.AppendAnd(a, b)
	mux, err := g.AppendMuxReal(control, a, GiaConst0)
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := g.AppendBuffer(mux)
	if err != nil {
		t.Fatal(err)
	}
	got, err := g.Mffc(buffer.ID(), 0)
	if err != nil || got.Size != 3 || len(got.Boundary) != 3 {
		t.Fatal(got, err)
	}
	for _, leaf := range []GiaID{a.ID(), b.ID(), 0} {
		if !slices.Contains(got.Boundary, leaf) {
			t.Fatal("missing extended boundary", got)
		}
	}
	for _, delta := range got.BoundaryReferenceDeltas {
		if delta >= 0 {
			t.Fatal("wrong reference delta", got)
		}
	}
}
