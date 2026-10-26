package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaReferenceCountsKindsAndOwnership(t *testing.T) {
	g := NewGiaManager()
	a, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	b, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	and, err := g.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	buffer, err := g.AppendBuffer(and)
	if err != nil {
		t.Fatal(err)
	}
	xor, err := g.AppendXorReal(a, b)
	if err != nil {
		t.Fatal(err)
	}
	mux, err := g.AppendMuxReal(a, and, xor)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(buffer); err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(mux); err != nil {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	counts, err := g.ReferenceCounts(0)
	if err != nil {
		t.Fatal(err)
	}
	want := make([]uint64, g.ObjectCount())
	want[a.ID()] = 3
	want[b.ID()] = 2
	want[and.ID()] = 2
	want[buffer.ID()] = 1
	want[xor.ID()] = 1
	want[mux.ID()] = 1
	data, err := os.ReadFile("test_vectors/oracle/gia_refs.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []uint64
	if err = json.Unmarshal(data, &native); err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(counts, native) {
		t.Fatal("native reference mismatch", counts, native)
	}
	if !slices.Equal(counts, want) {
		t.Fatal(counts, want)
	}
	if !slices.Equal(g.objects, before) {
		t.Fatal("reference counting mutated scratch")
	}
	counts[a.ID()] = 99
	again, err := g.ReferenceCounts(g.ObjectCount())
	if err != nil || !slices.Equal(again, want) {
		t.Fatal("counts alias manager")
	}
	if _, err := g.ReferenceCounts(g.ObjectCount() - 1); err == nil {
		t.Fatal("object cap ignored")
	}
	if _, err := g.ReferenceCounts(-1); err == nil {
		t.Fatal("negative cap")
	}
}
