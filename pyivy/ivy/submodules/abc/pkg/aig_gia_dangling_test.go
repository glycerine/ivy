package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaDanglingNativeRootsAndMarks(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	s, _ := g.AppendAnd(n, c)
	buffer, _ := g.AppendBuffer(b)
	if _, err := g.AppendCO(n); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_dangling.json")
	if err != nil {
		t.Fatal(err)
	}
	var native struct {
		Roots []GiaID
		Count int
	}
	if err = json.Unmarshal(raw, &native); err != nil {
		t.Fatal(err)
	}
	// Arbitrary caller mark bits survive the analysis.
	g.objects[n.ID()].word0 |= uint32(1) << 30
	before := slices.Clone(g.objects)
	roots, err := g.DanglingRoots(0)
	if err != nil || !slices.Equal(roots, native.Roots) || !slices.Equal(roots, []GiaID{s.ID(), buffer.ID()}) {
		t.Fatal(roots, native, err)
	}
	count, err := g.DanglingRootCount(0)
	if err != nil || count != native.Count {
		t.Fatal(count, native, err)
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("mark bits changed")
	}
	for _, cap := range []int{-1, g.ObjectCount() - 1} {
		if _, err := g.DanglingRoots(cap); err == nil {
			t.Fatal("invalid cap")
		}
	}
}
func TestGiaDanglingNativeControlEdge(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	control, _ := g.AppendAnd(a, b)
	mux, err := g.AppendMuxReal(control, a, c)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(mux); err != nil {
		t.Fatal(err)
	}
	roots, err := g.DanglingRoots(0)
	if err != nil || len(roots) != 0 {
		t.Fatal("control driver considered dangling", roots, err)
	}
	empty := NewGiaManager()
	count, err := empty.DanglingRootCount(0)
	if err != nil || count != 0 {
		t.Fatal(count, err)
	}
}
