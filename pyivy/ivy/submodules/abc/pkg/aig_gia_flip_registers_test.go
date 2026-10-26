package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaFlipRegistersStateEquivalence(t *testing.T) {
	m := NewGiaManager()
	m.Name = "flip"
	m.Spec = "flip.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	buf, _ := m.AppendBuffer(b.Not())
	mux, _ := m.AppendMuxReal(a, x, buf)
	m.AppendCO(mux)
	m.AppendCO(b)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for mask := 0; mask < 4; mask++ {
		phases := []bool{mask&1 != 0, mask&2 != 0}
		g, e := m.FlipRegisters(phases, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if g.registers != 2 || g.constraints != 1 || g.Name != m.Name || g.Spec != m.Spec || g.XorCount() != 1 || g.BufferCount() != 1 || g.MuxCount() != 1 || g.ObjectCount() != m.ObjectCount() || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source")
		}
		for assignment := 0; assignment < 8; assignment++ {
			input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			transformed := slices.Clone(input)
			for i, phase := range phases {
				transformed[1+i] = transformed[1+i] != phase
			}
			want, e := m.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			got, e := g.Evaluate(transformed)
			if e != nil {
				t.Fatal(e)
			}
			if !slices.Equal(got[:2], want[:2]) {
				t.Fatal("properties", mask, assignment)
			}
			for i, phase := range phases {
				if got[2+i] != (want[2+i] != phase) {
					t.Fatal("next state", mask, assignment)
				}
			}
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.FlipRegisters([]bool{true, false}, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 12} {
		calls := 0
		g, e := m.FlipRegisters([]bool{true, false}, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if _, e := m.FlipRegisters(nil, GiaCofactorOptions{}); e == nil {
		t.Fatal("dimensions")
	}
}
func FuzzGiaRegisterPhase(f *testing.F) {
	f.Add(uint8(5), uint8(7))
	f.Fuzz(func(t *testing.T, mask, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		m.AppendCO(b.Not())
		m.AppendCO(x)
		m.AppendCO(a.Not())
		m.SetRegisterCount(2)
		phases := []bool{mask&1 != 0, mask&2 != 0}
		g, e := m.FlipRegisters(phases, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		pv, av, bv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		got, e := g.Evaluate([]bool{pv, av != phases[0], bv != phases[1]})
		want := []bool{!bv, (pv != av) != phases[0], (!av) != phases[1]}
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaFlipRegistersNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	y, _ := m.AppendAnd(x.Not(), b)
	m.AppendCO(y)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.SetRegisterCount(2)
	raw, e := os.ReadFile("test_vectors/oracle/gia_flip_registers.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows [][]GiaLit
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 4 {
		t.Fatal(e)
	}
	for mask, row := range rows {
		g, e := m.FlipRegisters([]bool{mask&1 != 0, mask&2 != 0}, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if !slices.Equal(drivers, row) {
			t.Fatal(mask, drivers, row)
		}
	}
}
