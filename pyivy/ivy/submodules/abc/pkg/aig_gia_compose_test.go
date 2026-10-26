package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaCompositionSequentialConstraints(t *testing.T) {
	build := func(reverse bool) *GiaManager {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		r, _ := m.AppendCI()
		x, _ := m.AppendXorReal(a, b)
		mux, _ := m.AppendMuxReal(r, x, a)
		m.AppendCO(mux.NotIf(reverse))
		m.AppendCO(b.Not())
		m.AppendCO(x)
		m.SetRegisterCount(1)
		m.SetConstraintCount(1)
		return m
	}
	first, second := build(false), build(true)
	first.Name = "compose"
	first.Spec = "first.aig"
	before := slices.Clone(first.objects)
	r, e := ComposeGiaSharedInputs(first, second, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	g := r.Graph
	if g.InputCount() != 4 || g.registers != 2 || g.constraints != 2 || g.Name != first.Name || g.Spec != first.Spec || !slices.Equal(before, first.objects) {
		t.Fatal("metadata/source")
	}
	for assignment := 0; assignment < 16; assignment++ {
		a, b, r1, r2 := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0
		x := a != b
		v1, v2 := a, a
		if r1 {
			v1 = x
		}
		if r2 {
			v2 = x
		}
		want := []bool{v1, !v2, !b, !b, x, x}
		got, e := g.Evaluate([]bool{a, b, r1, r2})
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, e)
		}
	}
	if r.First[first.cis[0]] != r.Second[second.cis[0]] || r.First[first.cis[2]] == r.Second[second.cis[2]] {
		t.Fatal("input mappings")
	}
	for source, m := range []*GiaManager{first, second} {
		mapping := r.First
		if source == 1 {
			mapping = r.Second
		}
		for _, id := range m.cos {
			if !g.objects[mapping[id].ID()].IsCO() {
				t.Fatal("CO map")
			}
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := ComposeGiaSharedInputs(first, second, options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 5, 15, 25} {
		calls := 0
		r, e := ComposeGiaSharedInputs(first, second, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if _, e := ComposeGiaSharedInputs(first, NewGiaManager(), GiaCofactorOptions{}); e == nil {
		t.Fatal("PI mismatch")
	}
}

func TestGiaCompositionNative(t *testing.T) {
	build := func(phase bool) *GiaManager {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		r, _ := m.AppendCI()
		x, _ := m.AppendAnd(a, b)
		y, _ := m.AppendAnd(x, r)
		m.AppendCO(y.NotIf(phase))
		m.AppendCO(x)
		m.SetRegisterCount(1)
		return m
	}
	result, e := ComposeGiaSharedInputs(build(false), build(true), GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_compose.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct {
		Objects int
		Drivers []GiaLit
	}
	if e = json.Unmarshal(raw, &native); e != nil {
		t.Fatal(e)
	}
	drivers := make([]GiaLit, 0)
	for _, id := range result.Graph.cos {
		l, _ := result.Graph.objects[id].Fanin0(id)
		drivers = append(drivers, l)
	}
	if result.Graph.ObjectCount() != native.Objects || !slices.Equal(drivers, native.Drivers) {
		t.Fatal(drivers, native)
	}
}
func FuzzGiaComposition(f *testing.F) {
	f.Add(uint8(7), uint8(42))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		build := func(bit uint8) *GiaManager {
			m := NewGiaManager()
			a, _ := m.AppendCI()
			b, _ := m.AppendCI()
			r, _ := m.AppendCI()
			x, _ := m.AppendXorReal(a.NotIf(bit&1 != 0), b)
			y, _ := m.AppendMuxReal(r, x, a)
			m.AppendCO(y)
			m.AppendCO(b.NotIf(bit&2 != 0))
			m.SetRegisterCount(1)
			return m
		}
		r, e := ComposeGiaSharedInputs(build(phase), build(phase>>2), GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		a, b, r1, r2 := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0
		expected := func(bit uint8, state bool) bool {
			if state {
				return (a != (bit&1 != 0)) != b
			}
			return a
		}
		got, e := r.Graph.Evaluate([]bool{a, b, r1, r2})
		want := []bool{expected(phase, r1), expected(phase>>2, r2), b != (phase&2 != 0), b != (phase&8 != 0)}
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	})
}
