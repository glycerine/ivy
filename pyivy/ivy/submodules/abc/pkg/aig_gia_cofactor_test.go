package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaCofactorRebuildEquivalenceAndMappings(t *testing.T) {
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
	g.Name = "cofactor"
	g.Spec = "source.aig"
	raw, err := os.ReadFile("test_vectors/oracle/gia_cofactor.json")
	if err != nil {
		t.Fatal(err)
	}
	var native [][]uint32
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 5 {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	for _, pivot := range []GiaID{a.ID(), b.ID(), c.ID(), n.ID(), root.ID()} {
		result, mapping, err := g.CofactorRebuild(pivot, GiaCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if result.Name != g.Name || result.Spec != g.Spec || result.InputCount() != 3 || result.OutputCount() != 2 || mapping[pivot] != GiaConst1 {
			t.Fatal("metadata or pivot map")
		}
		for _, id := range g.cos {
			obj, err := result.Object(mapping[id].ID())
			if err != nil || !obj.IsCO() {
				t.Fatal("CO map", err)
			}
		}
		masks := make([]uint32, 2)
		for assignment := 0; assignment < 8; assignment++ {
			bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			values, err := result.Evaluate(bits)
			if err != nil {
				t.Fatal(err)
			}
			for i, value := range values {
				if value {
					masks[i] |= 1 << assignment
				}
			}
			if values[0] != (bits[0] && bits[1]) || values[1] != (bits[0] && bits[1] && bits[2]) {
				t.Fatal("Shannon reconstruction", pivot, assignment, values)
			}
		}
		if !slices.Equal(masks, native[int(pivot)-1]) {
			t.Fatal("native cofactor reconstruction", pivot, masks, native[int(pivot)-1])
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("source mutated")
	}
	for _, pivot := range []GiaID{0, g.cos[0], GiaID(g.ObjectCount())} {
		if _, _, err := g.CofactorRebuild(pivot, GiaCofactorOptions{}); err == nil {
			t.Fatal("invalid pivot")
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		if _, _, err := g.CofactorRebuild(n.ID(), options); err == nil {
			t.Fatal("limits")
		}
	}
	for _, threshold := range []int{1, 4, 10} {
		calls := 0
		graph, mapping, err := g.CofactorRebuild(n.ID(), GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) || graph != nil || mapping != nil {
			t.Fatal("cancellation", threshold, err)
		}
	}
}
func TestGiaCofactorRebuildExtendedAndConstantPivot(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	state, _ := g.AppendCI()
	mux, err := g.AppendMuxReal(state, a, b)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(mux); err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(a); err != nil {
		t.Fatal(err)
	}
	if err = g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	result, _, err := g.CofactorRebuild(mux.ID(), GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if result.RegisterCount() != 1 {
		t.Fatal("lost registers")
	}
	for assignment := 0; assignment < 8; assignment++ {
		bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, err := g.Evaluate(bits)
		if err != nil {
			t.Fatal(err)
		}
		got, err := result.Evaluate(bits)
		if err != nil || !slices.Equal(got, want) {
			t.Fatal("native mux reconstruction", got, want, err)
		}
	}
	constant, _ := g.AppendAnd(a, GiaConst0)
	result, _, err = g.CofactorRebuild(constant.ID(), GiaCofactorOptions{})
	if err != nil || result == nil {
		t.Fatal("constant-zero pivot", err)
	}
}
