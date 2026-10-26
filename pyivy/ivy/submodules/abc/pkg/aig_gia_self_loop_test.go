package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaSelfLoopNative(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	r, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, r)
	m.AppendCO(x.Not())
	m.AppendCO(a.Not())
	m.SetRegisterCount(1)
	g, e := m.WithRegisterHoldControl(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_self_loop.json")
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
		t.Fatal(drivers, native, g.ObjectCount())
	}
	for k := 0; k < 8; k++ {
		control, a, r := k&1 != 0, k&2 != 0, k&4 != 0
		next := r
		if control {
			next = !a
		}
		got, e := g.Evaluate([]bool{control, a, r})
		if e != nil || !slices.Equal(got, []bool{!(a && r), next}) {
			t.Fatal(k, got, e)
		}
	}
}
func TestGiaSelfLoopExtendedLimits(t *testing.T) {
	m := NewGiaManager()
	m.Name = "hold"
	m.Spec = "hold.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	buf, _ := m.AppendBuffer(x.Not())
	mux, _ := m.AppendMuxReal(b, buf, p)
	m.AppendCO(mux)
	m.AppendCO(p)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	g, e := m.WithRegisterHoldControl(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if g.registers != 2 || g.constraints != 1 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/source")
	}
	for k := 0; k < 16; k++ {
		ctrl, p, a, b := k&1 != 0, k&2 != 0, k&4 != 0, k&8 != 0
		v := p
		if b {
			v = !(p != a)
		}
		na, nb := a, b
		if ctrl {
			na = p != a
			nb = !a
		}
		got, e := g.Evaluate([]bool{ctrl, p, a, b})
		want := []bool{v, p, na, nb}
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(k, got, want, e)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.WithRegisterHoldControl(options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 13} {
		calls := 0
		g, e := m.WithRegisterHoldControl(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	empty := NewGiaManager()
	g, e = empty.WithRegisterHoldControl(GiaCofactorOptions{})
	if e != nil || g.InputCount() != 1 || g.OutputCount() != 0 {
		t.Fatal("empty", e)
	}
}

func FuzzGiaRegisterHold(f *testing.F) {
	f.Add(uint8(5), uint8(13))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		m.AppendCO(b.NotIf(phase&1 != 0))
		m.AppendCO(x.NotIf(phase&2 != 0))
		m.AppendCO(a.NotIf(phase&4 != 0))
		m.SetRegisterCount(2)
		g, e := m.WithRegisterHoldControl(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		ctrl, pv, av, bv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0
		na, nb := av, bv
		if ctrl {
			na = (pv != av) != (phase&2 != 0)
			nb = av != (phase&4 != 0)
		}
		want := []bool{bv != (phase&1 != 0), na, nb}
		got, e := g.Evaluate([]bool{ctrl, pv, av, bv})
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	})
}
