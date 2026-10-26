package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaSelectedOutputsNativeOrderAndTruth(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	n, _ := g.AppendAnd(a, b)
	r, _ := g.AppendAnd(n, c)
	for _, l := range []GiaLit{n, r, a.Not(), GiaConst1} {
		if _, err := g.AppendCO(l); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := os.ReadFile("test_vectors/oracle/gia_select_outputs.json")
	if err != nil {
		t.Fatal(err)
	}
	var native []struct {
		Objects int
		Truth   []uint32
	}
	if err = json.Unmarshal(raw, &native); err != nil || len(native) != 4 {
		t.Fatal(err)
	}
	before := slices.Clone(g.objects)
	for i, selection := range [][]int{{1, 0}, {2}, {3, 1, 3}, nil} {
		graph, err := g.SelectOutputs(selection, GiaCofactorOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if graph.InputCount() != 3 || graph.OutputCount() != len(selection) || graph.ObjectCount() != native[i].Objects {
			t.Fatal("selection dimensions", selection)
		}
		masks := make([]uint32, len(selection))
		for assignment := 0; assignment < 8; assignment++ {
			bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			values, err := graph.Evaluate(bits)
			if err != nil {
				t.Fatal(err)
			}
			expected := []bool{bits[0] && bits[1], bits[0] && bits[1] && bits[2], !bits[0], true}
			for output, source := range selection {
				if values[output] != expected[source] {
					t.Fatal("selected truth", selection, assignment, values)
				}
				if values[output] {
					masks[output] |= 1 << assignment
				}
			}
		}
		if !slices.Equal(masks, native[i].Truth) {
			t.Fatal("native selected masks", masks, native[i])
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("selection mutated source")
	}
	mask, err := g.SelectOutputMask([]bool{true, false, true, false}, GiaCofactorOptions{})
	if err != nil || mask.OutputCount() != 2 {
		t.Fatal(err)
	}
	if _, err := g.SelectOutputMask([]bool{true}, GiaCofactorOptions{}); err == nil {
		t.Fatal("invalid mask")
	}
	for _, selection := range [][]int{{-1}, {4}, {0, 0, 0, 0, 0}} {
		if _, err := g.SelectOutputs(selection, GiaCofactorOptions{}); err == nil {
			t.Fatal("invalid selection")
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}} {
		if _, err := g.SelectOutputs([]int{1}, options); err == nil {
			t.Fatal("invalid cap")
		}
	}
	calls := 0
	graph, err := g.SelectOutputs([]int{1}, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= 4 }})
	if !errors.Is(err, ErrCancelled) || graph != nil {
		t.Fatal("selection cancellation", err)
	}
	if err := g.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	graph, err = g.SelectOutputs([]int{1, 3, 3}, GiaCofactorOptions{})
	if err != nil || graph.ConstraintCount() != 2 {
		t.Fatal("constraint selection", err)
	}
	if _, err := g.SelectOutputs([]int{3, 1}, GiaCofactorOptions{}); err == nil {
		t.Fatal("constraint ordering")
	}
}
func TestGiaSelectedOutputsPreservesExtendedKinds(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	c, _ := g.AppendCI()
	xor, err := g.AppendXorReal(a, b)
	if err != nil {
		t.Fatal(err)
	}
	mux, err := g.AppendMuxReal(c, xor, a)
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
	graph, err := g.SelectOutputs([]int{0}, GiaCofactorOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if graph.XorCount() != 1 || graph.MuxCount() != 1 || graph.BufferCount() != 1 {
		t.Fatal("extended kinds lost")
	}
	for assignment := 0; assignment < 8; assignment++ {
		bits := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		want, _ := g.Evaluate(bits)
		got, err := graph.Evaluate(bits)
		if err != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, err)
		}
	}
}
