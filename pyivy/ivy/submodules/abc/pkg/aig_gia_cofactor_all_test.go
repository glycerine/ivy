package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaCofactorAllNativeAndEquivalence(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	root, _ := g.AppendAnd(n, c)
	for _, l := range []GiaLit{n, root} {
		if _, err := g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_cofactor_all.json")
	if err != nil {
		t.Fatal(err)
	}
	var native [][]uint32
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 4 {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	for i, signals := range [][]GiaID{nil, {a.ID()}, {n.ID(), a.ID()}, {root.ID(), n.ID(), b.ID()}} {
		graph, err := g.CofactorRebuildAll(signals, GiaMultiCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if graph == g || graph.InputCount() != 3 || graph.OutputCount() != 2 {
			t.Fatal("clone/dimensions")
		}
		masks := make([]uint32, 2)
		for assignment := 0; assignment < 8; assignment++ {
			bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			values, err := graph.Evaluate(bits)
			if err != nil {
				t.Fatal(err)
			}
			if values[0] != (bits[0] && bits[1]) || values[1] != (bits[0] && bits[1] && bits[2]) {
				t.Fatal("multi-pivot equivalence", signals, assignment)
			}
			for output, value := range values {
				if value {
					masks[output] |= 1 << assignment
				}
			}
		}
		if !slices.Equal(masks, native[i]) {
			t.Fatal("native multi-pivot truth", signals, masks, native[i])
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("source mutated")
	}
	for _, options := range []GiaMultiCofactorOptions{{MaxPasses: 1}, {MaxPasses: -1}, {GiaCofactorOptions: GiaCofactorOptions{MaxObjects: 1}}, {GiaCofactorOptions: GiaCofactorOptions{MaxMappingEntries: 1}}} {
		if graph, err := g.CofactorRebuildAll([]GiaID{n.ID(), a.ID()}, options); err == nil || graph != nil {
			t.Fatal("resource cap", options, err)
		}
	}
	for _, threshold := range []int{1, 5, 20} {
		calls := 0
		graph, err := g.CofactorRebuildAll([]GiaID{n.ID(), a.ID()}, GiaMultiCofactorOptions{GiaCofactorOptions: GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }}})
		if !errors.Is(err, ErrCancelled) || graph != nil {
			t.Fatal("cancellation", threshold, err)
		}
	}
	if _, err := g.CofactorRebuildAll(make([]GiaID, 201), GiaMultiCofactorOptions{}); err == nil {
		t.Fatal("C signal cap")
	}
	if _, err := g.CofactorRebuildAll([]GiaID{0}, GiaMultiCofactorOptions{}); err == nil {
		t.Fatal("invalid pivot")
	}
}
