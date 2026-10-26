package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaStatePropertyReplacement(t *testing.T) {
	m := NewGiaManager()
	m.Name = "state"
	m.Spec = "state.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	m.AppendCO(p.Not())
	m.AppendCO(p)
	m.AppendCO(x)
	m.AppendCO(b.Not())
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	property := NewGiaManager()
	u, _ := property.AppendCI()
	v, _ := property.AppendCI()
	buf, _ := property.AppendBuffer(u.Not())
	mux, _ := property.AppendMuxReal(v, buf, u)
	property.AppendCO(mux)
	property.AppendCO(v)
	property.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	beforeProp := slices.Clone(property.objects)
	g, e := m.WithStateProperties(property, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if g.registers != 2 || g.constraints != 2 || g.Name != m.Name || g.Spec != m.Spec || g.InputCount() != 3 || !slices.Equal(before, m.objects) || !slices.Equal(beforeProp, property.objects) {
		t.Fatal("metadata/source")
	}
	for k := 0; k < 8; k++ {
		p, a, b := k&1 != 0, k&2 != 0, k&4 != 0
		main := a
		if b {
			main = !a
		}
		want := []bool{main, b, p, p != a, !b}
		got, e := g.Evaluate([]bool{p, a, b})
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(k, got, want, e)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.WithStateProperties(property, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 12} {
		calls := 0
		g, e := m.WithStateProperties(property, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if _, e := m.WithStateProperties(nil, GiaCofactorOptions{}); e == nil {
		t.Fatal("nil")
	}
	if _, e := m.WithStateProperties(NewGiaManager(), GiaCofactorOptions{}); e == nil {
		t.Fatal("dimensions")
	}
}
func FuzzGiaStateProperties(f *testing.F) {
	f.Add(uint8(5), uint8(7))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		m.AppendCO(p)
		m.AppendCO(x)
		m.AppendCO(b.Not())
		m.SetRegisterCount(2)
		property := NewGiaManager()
		u, _ := property.AppendCI()
		v, _ := property.AppendCI()
		y, _ := property.AppendAnd(u.NotIf(phase&1 != 0), v)
		property.AppendCO(y.NotIf(phase&2 != 0))
		g, e := m.WithStateProperties(property, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		pv, av, bv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		main := (((av != (phase&1 != 0)) && bv) != (phase&2 != 0))
		want := []bool{main, pv != av, !bv}
		got, e := g.Evaluate([]bool{pv, av, bv})
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaStatePropertiesNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	m.AppendCO(p)
	m.AppendCO(x)
	m.AppendCO(b.Not())
	m.SetRegisterCount(2)
	property := NewGiaManager()
	u, _ := property.AppendCI()
	v, _ := property.AppendCI()
	y, _ := property.AppendAnd(u.Not(), v)
	property.AppendCO(y)
	property.AppendCO(v.Not())
	g, e := m.WithStateProperties(property, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_replace_properties.json")
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
	if g.ObjectCount() != native.Objects || !slices.Equal(drivers, native.Drivers) {
		t.Fatal(drivers, native)
	}
}
