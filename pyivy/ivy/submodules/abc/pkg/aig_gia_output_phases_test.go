package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaOutputPhasesRawLayout(t *testing.T) {
	m := NewGiaManager()
	m.Name = "phases"
	m.Spec = "phases.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	buf, _ := m.AppendBuffer(b.Not())
	mux, _ := m.AppendMuxReal(a, x, buf)
	m.AppendCO(mux)
	m.AppendCO(p)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	g, e := m.DuplicateAllOutputPhases(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if g.registers != 2 || g.constraints != 0 || g.OutputCount() != 8 || g.PrimaryOutputCount() != 6 || g.Name != m.Name || g.Spec != m.Spec || g.XorCount() != 1 || g.MuxCount() != 1 || g.BufferCount() != 1 || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/source")
	}
	for k := 0; k < 8; k++ {
		input := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
		original, e := m.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		got, e := g.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		for i, v := range original {
			if got[2*i] != v || got[2*i+1] == v {
				t.Fatal(k, i, got, original)
			}
		}
		if got[6] != !input[1] || got[7] != input[1] {
			t.Fatal("raw RI reassignment", got)
		}
	}
	if _, e := m.DuplicateAllOutputPhases(GiaCofactorOptions{MaxObjects: g.ObjectCount()}); e != nil {
		t.Fatal("exact", e)
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: g.ObjectCount() - 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.DuplicateAllOutputPhases(options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 11} {
		calls := 0
		g, e := m.DuplicateAllOutputPhases(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	empty := NewGiaManager()
	g, e = empty.DuplicateAllOutputPhases(GiaCofactorOptions{})
	if e != nil || g.ObjectCount() != 1 {
		t.Fatal(e)
	}
}
func FuzzGiaOutputPhases(f *testing.F) {
	f.Add(uint8(3), uint8(5))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		m.AppendCO(x.NotIf(phase&1 != 0))
		m.AppendCO(b.Not())
		m.AppendCO(p.NotIf(phase&2 != 0))
		m.SetRegisterCount(2)
		g, e := m.DuplicateAllOutputPhases(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		pv, av, bv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		want := []bool{(pv != av) != (phase&1 != 0), !bv, pv != (phase&2 != 0)}
		got, e := g.Evaluate([]bool{pv, av, bv})
		if e != nil {
			t.Fatal(e)
		}
		for i, v := range want {
			if got[2*i] != v || got[2*i+1] == v {
				t.Fatal(got, want)
			}
		}
	})
}

func TestGiaOutputPhasesNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	m.AppendCO(x.Not())
	m.AppendCO(b)
	m.AppendCO(p.Not())
	m.SetRegisterCount(2)
	g, e := m.DuplicateAllOutputPhases(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_output_phases.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct {
		Objects, Registers int
		Drivers            []GiaLit
	}
	if e = json.Unmarshal(raw, &native); e != nil {
		t.Fatal(e)
	}
	drivers := make([]GiaLit, 0)
	for _, id := range g.cos {
		l, _ := g.objects[id].Fanin0(id)
		drivers = append(drivers, l)
	}
	if g.ObjectCount() != native.Objects || g.registers != native.Registers || !slices.Equal(drivers, native.Drivers) {
		t.Fatal(drivers, native)
	}
}
