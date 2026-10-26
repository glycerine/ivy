package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaAppendPropertyConesModes(t *testing.T) {
	m := NewGiaManager()
	m.Name = "cones"
	m.Spec = "cones.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	m.AppendCO(x)
	m.AppendCO(b)
	m.AppendCO(a.Not())
	m.AppendCO(x)
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for _, onlyRegs := range []bool{false, true} {
		cone := NewGiaManager()
		inputs := make([]GiaLit, 3)
		if onlyRegs {
			inputs = inputs[:2]
		}
		for i := range inputs {
			inputs[i], _ = cone.AppendCI()
		}
		v, _ := cone.AppendXorReal(inputs[0], inputs[1])
		cone.AppendCO(v.Not())
		constant := NewGiaManager()
		for range inputs {
			constant.AppendCI()
		}
		constant.AppendCO(GiaConst1)
		g, e := m.AppendPropertyCones([]*GiaManager{cone, constant, cone}, onlyRegs, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if g.registers != 2 || g.constraints != 1 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source")
		}
		for assignment := 0; assignment < 8; assignment++ {
			p, a, b := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
			cv := p == a
			if onlyRegs {
				cv = a == b
			}
			want := []bool{p != a, cv, true, cv, b, !a, p != a}
			got, e := g.Evaluate([]bool{p, a, b})
			if e != nil || !slices.Equal(got, want) {
				t.Fatal(onlyRegs, assignment, got, want, e)
			}
		}
		for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
			g, e := m.AppendPropertyCones([]*GiaManager{cone}, onlyRegs, options)
			if g != nil || e == nil {
				t.Fatal(options, e)
			}
		}
		for _, threshold := range []int{1, 4, 8, 12} {
			calls := 0
			g, e := m.AppendPropertyCones([]*GiaManager{cone}, onlyRegs, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
			if g != nil || !errors.Is(e, ErrCancelled) {
				t.Fatal(threshold, calls, e)
			}
		}
	}
	if _, e := m.AppendPropertyCones([]*GiaManager{nil}, false, GiaCofactorOptions{}); e == nil {
		t.Fatal("nil cone")
	}
	bad := NewGiaManager()
	bad.AppendCO(GiaConst0)
	if _, e := m.AppendPropertyCones([]*GiaManager{bad}, false, GiaCofactorOptions{}); e == nil {
		t.Fatal("dimensions")
	}
}
func FuzzGiaAppendPropertyCones(f *testing.F) {
	f.Add(uint8(3), uint8(7))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		m.AppendCO(p)
		m.AppendCO(a.Not())
		m.AppendCO(b)
		m.SetRegisterCount(2)
		onlyRegs := phase&1 != 0
		cone := NewGiaManager()
		u, _ := cone.AppendCI()
		v, _ := cone.AppendCI()
		if !onlyRegs {
			cone.AppendCI()
		}
		x, _ := cone.AppendXorReal(u.NotIf(phase&2 != 0), v)
		cone.AppendCO(x.NotIf(phase&4 != 0))
		g, e := m.AppendPropertyCones([]*GiaManager{cone}, onlyRegs, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		pv, av, bv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		uval, vval := pv, av
		if onlyRegs {
			uval, vval = av, bv
		}
		cv := ((uval != (phase&2 != 0)) != vval) != (phase&4 != 0)
		want := []bool{pv, cv, !av, bv}
		got, e := g.Evaluate([]bool{pv, av, bv})
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaAppendPropertyConesNative(t *testing.T) {
	raw, e := os.ReadFile("test_vectors/oracle/gia_append_cones.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects int
		Drivers []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 2 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		r, _ := m.AppendCI()
		x, _ := m.AppendAnd(a, b)
		m.AppendCO(x)
		m.AppendCO(b.Not())
		m.AppendCO(r)
		m.SetRegisterCount(2)
		c := NewGiaManager()
		u, _ := c.AppendCI()
		v, _ := c.AppendCI()
		if mode == 0 {
			c.AppendCI()
		}
		y, _ := c.AppendAnd(u, v)
		c.AppendCO(y.Not())
		g, e := m.AppendPropertyCones([]*GiaManager{c}, mode != 0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if g.ObjectCount() != row.Objects || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(mode, drivers, row, g.ObjectCount())
		}
	}
}
