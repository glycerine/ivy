package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaTrimDirectOutputsProjection(t *testing.T) {
	m := NewGiaManager()
	m.Name = "trim"
	m.Spec = "trim.aig"
	p, _ := m.AppendCI()
	unused, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	buf, _ := m.AppendBuffer(b.Not())
	m.AppendCO(unused.Not())
	m.AppendCO(GiaConst1)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.AppendCO(buf)
	m.AppendCO(p)
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	r, e := m.TrimDirectInputOutputs(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if !slices.Equal(r.Inputs, []int{0, 2, 3}) || !slices.Equal(r.Outputs, []int{2, 3, 4, 5}) || r.Graph.constraints != 0 || r.Graph.registers != 2 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || r.Graph.XorCount() != 1 || r.Graph.BufferCount() != 1 || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/projections", r)
	}
	for k := 0; k < 16; k++ {
		source := make([]bool, 4)
		for i := range source {
			source[i] = k&(1<<i) != 0
		}
		input := make([]bool, len(r.Inputs))
		for i, index := range r.Inputs {
			input[i] = source[index]
		}
		want, e := m.Evaluate(source)
		if e != nil {
			t.Fatal(e)
		}
		got, e := r.Graph.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		for i, index := range r.Outputs {
			if got[i] != want[index] {
				t.Fatal(k, got, want)
			}
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.TrimDirectInputOutputs(options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 3, 6, 9} {
		calls := 0
		r, e := m.TrimDirectInputOutputs(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	dummy := NewGiaManager()
	u, _ := dummy.AppendCI()
	dummy.AppendCO(u.Not())
	dummy.AppendCO(GiaConst1)
	r, e = dummy.TrimDirectInputOutputs(GiaCofactorOptions{})
	if e != nil || !slices.Equal(r.Inputs, []int{-1}) || !slices.Equal(r.Outputs, []int{-1}) {
		t.Fatal(r, e)
	}
	for _, input := range []bool{false, true} {
		v, e := r.Graph.Evaluate([]bool{input})
		if e != nil || v[0] {
			t.Fatal(v, e)
		}
	}
}
func FuzzGiaTrimDirectOutputs(f *testing.F) {
	f.Add(uint8(2), uint8(5))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		u, _ := m.AppendCI()
		r, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, r)
		m.AppendCO(u.NotIf(phase&1 != 0))
		m.AppendCO(GiaConst1)
		m.AppendCO(x.NotIf(phase&2 != 0))
		m.AppendCO(p.Not())
		m.SetRegisterCount(1)
		result, e := m.TrimDirectInputOutputs(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		pv, rv := assignment&1 != 0, assignment&2 != 0
		got, e := result.Graph.Evaluate([]bool{pv, rv})
		want := []bool{(pv != rv) != (phase&2 != 0), !pv}
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaTrimDirectNative(t *testing.T) {
	raw, e := os.ReadFile("test_vectors/oracle/gia_trim_direct.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects, Inputs, Registers int
		Drivers                    []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 2 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		u, _ := m.AppendCI()
		if mode == 0 {
			r, _ := m.AppendCI()
			x, _ := m.AppendAnd(a, r)
			m.AppendCO(u)
			m.AppendCO(GiaConst1)
			m.AppendCO(x.Not())
			m.AppendCO(a)
			m.SetRegisterCount(1)
		} else {
			m.AppendCO(a.Not())
			m.AppendCO(u)
		}
		r, e := m.TrimDirectInputOutputs(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range r.Graph.cos {
			l, _ := r.Graph.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if r.Graph.ObjectCount() != row.Objects || r.Graph.InputCount() != row.Inputs || r.Graph.registers != row.Registers || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(mode, drivers, row)
		}
	}
}
