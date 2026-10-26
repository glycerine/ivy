package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaDuplicateGateOutputsNative(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	r, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(x, r)
	m.AppendAnd(b, r)
	m.AppendCO(y.Not())
	m.AppendCO(x)
	m.SetRegisterCount(1)
	g, mapping, e := m.DuplicateWithGateOutputs(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_dup_outputs.json")
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
	for _, id := range g.cos {
		l, _ := g.objects[id].Fanin0(id)
		drivers = append(drivers, l)
	}
	if g.ObjectCount() != native.Objects || !slices.Equal(drivers, native.Drivers) || mapping[m.cos[1]].ID() != g.cos[len(g.cos)-1] {
		t.Fatal(drivers, native)
	}
	for assignment := 0; assignment < 8; assignment++ {
		a, b, r := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		got, e := g.Evaluate([]bool{a, b, r})
		want := []bool{!((a && b) && r), a && b, (a && b) && r, b && r, a && b}
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, e)
		}
	}
}
func TestGiaDuplicateGateOutputsConstraintsAndLimits(t *testing.T) {
	m := NewGiaManager()
	m.Name = "outs"
	m.Spec = "outs.aig"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	buf, _ := m.AppendBuffer(x.Not())
	mux, _ := m.AppendMuxReal(c, buf, b)
	m.AppendCO(mux)
	m.AppendCO(a)
	m.AppendCO(b)
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	g, _, e := m.DuplicateWithGateOutputs(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if g.RegisterCount() != 1 || g.ConstraintCount() != 1 || g.Name != m.Name || g.Spec != m.Spec || g.XorCount() != 1 || g.BufferCount() != 1 || g.MuxCount() != 1 || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/source")
	}
	for assignment := 0; assignment < 8; assignment++ {
		a, b, c := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		xor := a != b
		value := b
		if c {
			value = !xor
		}
		got, e := g.Evaluate([]bool{a, b, c})
		want := []bool{value, xor, !xor, value, a, b}
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}, {MaxObjects: g.ObjectCount() - 1}, {MaxMappingEntries: m.ObjectCount() - 1}} {
		g, mapping, e := m.DuplicateWithGateOutputs(options)
		if g != nil || mapping != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	if _, _, e = m.DuplicateWithGateOutputs(GiaCofactorOptions{MaxObjects: g.ObjectCount()}); e != nil {
		t.Fatal(e)
	}
	for _, threshold := range []int{1, 4, 8, 13} {
		calls := 0
		g, mapping, e := m.DuplicateWithGateOutputs(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || mapping != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	empty := NewGiaManager()
	g, _, e = empty.DuplicateWithGateOutputs(GiaCofactorOptions{})
	if e != nil || g.OutputCount() != 0 {
		t.Fatal(e)
	}
}
