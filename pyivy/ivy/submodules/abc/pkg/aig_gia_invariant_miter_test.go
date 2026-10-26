package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaInvariantMiterTruthAndLimits(t *testing.T) {
	m := NewGiaManager()
	m.Name = "induction"
	m.Spec = "state.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	mux, _ := m.AppendMuxReal(b, x, a)
	m.AppendCO(p)
	m.AppendCO(p.Not())
	m.AppendCO(mux)
	m.AppendCO(a.Not())
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	inv := NewGiaManager()
	u, _ := inv.AppendCI()
	v, _ := inv.AppendCI()
	i, _ := inv.AppendXorReal(u, v)
	inv.AppendCO(i.Not())
	before := slices.Clone(m.objects)
	beforeInv := slices.Clone(inv.objects)
	g, e := GiaInvariantMiter(m, inv, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if g.registers != 0 || g.constraints != 0 || g.InputCount() != 3 || g.OutputCount() != 1 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) || !slices.Equal(beforeInv, inv.objects) {
		t.Fatal("metadata/source")
	}
	for assignment := 0; assignment < 8; assignment++ {
		p, a, b := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		next := a
		if b {
			next = p != a
		}
		want := (a == b) && !(next == !a)
		got, e := g.Evaluate([]bool{p, a, b})
		if e != nil || got[0] != want {
			t.Fatal(assignment, got, want, e)
		}
	}
	for _, constant := range []GiaLit{GiaConst0, GiaConst1} {
		c := NewGiaManager()
		c.AppendCI()
		c.AppendCI()
		c.AppendCO(constant)
		g, e := GiaInvariantMiter(m, c, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		v, e := g.Evaluate([]bool{false, false, false})
		if e != nil || v[0] {
			t.Fatal("constant", v, e)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := GiaInvariantMiter(m, inv, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 9, 12} {
		calls := 0
		g, e := GiaInvariantMiter(m, inv, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaInvariantMiter(f *testing.F) {
	f.Add(uint8(3), uint8(6))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		m.AppendCO(x.NotIf(phase&1 != 0))
		m.AppendCO(b.NotIf(phase&2 != 0))
		m.SetRegisterCount(2)
		inv := NewGiaManager()
		u, _ := inv.AppendCI()
		v, _ := inv.AppendCI()
		i, _ := inv.AppendAnd(u.NotIf(phase&4 != 0), v.NotIf(phase&8 != 0))
		inv.AppendCO(i.NotIf(phase&16 != 0))
		g, e := GiaInvariantMiter(m, inv, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		pv, av, bv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		predicate := func(a, b bool) bool { return ((a != (phase&4 != 0)) && (b != (phase&8 != 0))) != (phase&16 != 0) }
		want := predicate(av, bv) && !predicate((pv != av) != (phase&1 != 0), bv != (phase&2 != 0))
		got, e := g.Evaluate([]bool{pv, av, bv})
		if e != nil || got[0] != want {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaInvariantMiterNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	m.AppendCO(b)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.SetRegisterCount(2)
	inv := NewGiaManager()
	u, _ := inv.AppendCI()
	v, _ := inv.AppendCI()
	i, _ := inv.AppendAnd(u, v.Not())
	inv.AppendCO(i.Not())
	g, e := GiaInvariantMiter(m, inv, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_invariant_miter.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct {
		Objects int
		Mask    uint
	}
	if e = json.Unmarshal(raw, &native); e != nil {
		t.Fatal(e)
	}
	mask := uint(0)
	for k := 0; k < 8; k++ {
		v, e := g.Evaluate([]bool{k&1 != 0, k&2 != 0, k&4 != 0})
		if e != nil {
			t.Fatal(e)
		}
		if v[0] {
			mask |= 1 << k
		}
	}
	if g.ObjectCount() != native.Objects || mask != native.Mask {
		t.Fatal(mask, native, g.ObjectCount())
	}
}

func TestGiaInvariantMiterInvalidInputs(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	m.AppendCO(a)
	m.SetRegisterCount(1)
	inv := NewGiaManager()
	u, _ := inv.AppendCI()
	inv.AppendCO(u)
	for _, pair := range [][2]*GiaManager{{nil, inv}, {m, nil}, {NewGiaManager(), inv}, {m, NewGiaManager()}, {m, m}} {
		if g, e := GiaInvariantMiter(pair[0], pair[1], GiaCofactorOptions{}); g != nil || e == nil {
			t.Fatal(pair, e)
		}
	}
	inv.SetConstraintCount(1)
	if g, e := GiaInvariantMiter(m, inv, GiaCofactorOptions{}); g != nil || e == nil {
		t.Fatal("invariant constraint", e)
	}
}
