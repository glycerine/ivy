package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaGuardedCofactorNativeAndSemantics(t *testing.T) {
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
	if err := g.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_cofactor_object.json")
	if err != nil {
		t.Fatal(err)
	}
	var native [][]uint32
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 10 {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	for pivot := GiaID(1); pivot <= 5; pivot++ {
		for phase := 0; phase < 2; phase++ {
			graph, err := g.CofactorObjectGuarded(pivot, phase != 0, GiaCofactorOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if graph.InputCount() != 3 || graph.OutputCount() != 2 || graph.ConstraintCount() != 1 {
				t.Fatal("guarded metadata")
			}
			masks := make([]uint32, 2)
			for assignment := 0; assignment < 8; assignment++ {
				bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
				nodes := []bool{false, bits[0], bits[1], bits[2], bits[0] && bits[1], bits[0] && bits[1] && bits[2]}
				values, err := graph.Evaluate(bits)
				if err != nil {
					t.Fatal(err)
				}
				allowed := nodes[pivot] == (phase != 0)
				if values[0] != (nodes[n.ID()] && allowed) || values[1] != (nodes[root.ID()] && allowed) {
					t.Fatal("guarded truth", pivot, phase, assignment, values)
				}
				for output, value := range values {
					if value {
						masks[output] |= 1 << assignment
					}
				}
			}
			if !slices.Equal(masks, native[2*(int(pivot)-1)+phase]) {
				t.Fatal("native guarded masks", pivot, phase, masks, native)
			}
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("source mutated")
	}
	for _, pivot := range []GiaID{0, g.cos[0], GiaID(g.ObjectCount())} {
		if _, err := g.CofactorObjectGuarded(pivot, false, GiaCofactorOptions{}); err == nil {
			t.Fatal("invalid pivot")
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}} {
		if _, err := g.CofactorObjectGuarded(n.ID(), true, options); err == nil {
			t.Fatal("invalid bounds")
		}
	}
	for _, threshold := range []int{1, 5, 9} {
		calls := 0
		graph, err := g.CofactorObjectGuarded(n.ID(), true, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if !errors.Is(err, ErrCancelled) || graph != nil {
			t.Fatal("guarded cancellation", threshold, err)
		}
	}
}
func TestGiaGuardedCofactorPivotAfterOutput(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	if _, err := g.AppendCO(a); err != nil {
		t.Fatal(err)
	}
	pivot, err := g.AppendAnd(a, b)
	if err != nil {
		t.Fatal(err)
	}
	for _, fixed := range []bool{false, true} {
		graph, err := g.CofactorObjectGuarded(pivot.ID(), fixed, GiaCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for assignment := 0; assignment < 4; assignment++ {
			bits := []bool{assignment&1 != 0, assignment&2 != 0}
			values, err := graph.Evaluate(bits)
			expected := bits[0] && ((bits[0] && bits[1]) == fixed)
			if err != nil || values[0] != expected {
				t.Fatal("late pivot", fixed, assignment, values, err)
			}
		}
	}
	if err := g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if _, err := g.CofactorObjectGuarded(pivot.ID(), true, GiaCofactorOptions{}); err == nil {
		t.Fatal("sequential graph accepted")
	}
}
