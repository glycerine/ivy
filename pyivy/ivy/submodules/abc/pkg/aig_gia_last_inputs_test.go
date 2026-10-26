package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaLastInputsCheckedCopy(t *testing.T) {
	m := NewGiaManager()
	m.Name = "last"
	m.Spec = "last.aig"
	m.AppendCI()
	m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	buf, _ := m.AppendBuffer(a.Not())
	m.AppendCO(x)
	m.AppendCO(buf)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for count := 2; count <= 4; count++ {
		g, e := m.RetainLastPrimaryInputs(count, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if g.InputCount() != count || g.constraints != 1 || g.Name != m.Name || g.Spec != m.Spec || g.XorCount() != 1 || g.BufferCount() != 1 || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source")
		}
		for k := 0; k < 1<<count; k++ {
			input := make([]bool, count)
			for i := range input {
				input[i] = k&(1<<i) != 0
			}
			got, e := g.Evaluate(input)
			want := []bool{input[count-2] != input[count-1], !input[count-2]}
			if e != nil || !slices.Equal(got, want) {
				t.Fatal(count, k, got, want, e)
			}
		}
	}
	if g, e := m.RetainLastPrimaryInputs(1, GiaCofactorOptions{}); g != nil || !errors.Is(e, ErrInvariant) {
		t.Fatal("missing dependency", e)
	}
	for _, count := range []int{-1, 5} {
		if _, e := m.RetainLastPrimaryInputs(count, GiaCofactorOptions{}); e == nil {
			t.Fatal(count)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.RetainLastPrimaryInputs(2, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 3, 5, 7} {
		calls := 0
		g, e := m.RetainLastPrimaryInputs(2, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	constant := NewGiaManager()
	constant.AppendCI()
	constant.AppendCO(GiaConst1)
	g, e := constant.RetainLastPrimaryInputs(0, GiaCofactorOptions{})
	if e != nil || g.InputCount() != 0 {
		t.Fatal("zero", e)
	}
	// Even dangling logic cannot reference a dropped input.
	dangling := NewGiaManager()
	a, _ = dangling.AppendCI()
	b, _ = dangling.AppendCI()
	dangling.AppendAnd(a, b)
	dangling.AppendCO(b)
	if g, e := dangling.RetainLastPrimaryInputs(1, GiaCofactorOptions{}); g != nil || e == nil {
		t.Fatal("dangling dependency", e)
	}
}
func FuzzGiaLastInputs(f *testing.F) {
	f.Add(uint8(2), uint8(5))
	f.Fuzz(func(t *testing.T, skips, assignment uint8) {
		omitted := int(skips % 6)
		m := NewGiaManager()
		for i := 0; i < omitted; i++ {
			m.AppendCI()
		}
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(a, b)
		m.AppendCO(x.Not())
		g, e := m.RetainLastPrimaryInputs(2, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		av, bv := assignment&1 != 0, assignment&2 != 0
		v, e := g.Evaluate([]bool{av, bv})
		if e != nil || v[0] != (av == bv) {
			t.Fatal(v, e)
		}
	})
}

func TestGiaLastInputsNative(t *testing.T) {
	m := NewGiaManager()
	m.AppendCI()
	m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	m.AppendCO(x.Not())
	raw, e := os.ReadFile("test_vectors/oracle/gia_last_inputs.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects int
		Drivers []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 3 {
		t.Fatal(e)
	}
	for i, row := range rows {
		g, e := m.RetainLastPrimaryInputs(i+2, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if g.ObjectCount() != row.Objects || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(i, drivers, row)
		}
	}
}
