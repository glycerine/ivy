package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaExistentialAllOutputsCoordinateSemantics(t *testing.T) {
	m := NewGiaManager()
	m.Name = "coordinates"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	mux, _ := m.AppendMuxReal(b, x, p)
	m.AppendCO(mux)
	m.AppendCO(p)
	m.AppendCO(p.Not())
	m.AppendCO(x)
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	g, e := m.ExistentialAllOutputs(0, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if g.registers != 2 || g.constraints != 1 || g.InputCount() != 3 || g.Name != m.Name || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/source")
	}
	for k := 0; k < 8; k++ {
		input := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
		v, e := g.Evaluate(input)
		if e != nil || !slices.Equal(v, []bool{true, true, true, true}) {
			t.Fatal(k, v, e)
		}
	}
	// The independently quantified p and !p outputs can both be true, despite
	// having no common witness. This contract must not become a relation proof.
	if _, e := m.ExistentialAllOutputs(1, GiaCofactorOptions{}); e == nil {
		t.Fatal("RO is not a PI")
	}
	empty := NewGiaManager()
	empty.AppendCI()
	g, e = empty.ExistentialAllOutputs(0, GiaCofactorOptions{})
	if e != nil || g.OutputCount() != 0 || g.InputCount() != 1 {
		t.Fatal("empty outputs", e)
	}
}
func FuzzGiaExistentialAllOutputs(f *testing.F) {
	f.Add(uint8(5), uint8(3))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendAnd(p, a)
		y, _ := m.AppendXorReal(p, b)
		m.AppendCO(x.NotIf(phase&1 != 0))
		m.AppendCO(y)
		m.AppendCO(p.Not())
		m.SetRegisterCount(2)
		g, e := m.ExistentialAllOutputs(0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		pv, av, bv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		first := av
		if phase&1 != 0 {
			first = true
		}
		got, e := g.Evaluate([]bool{pv, av, bv})
		want := []bool{first, true, true}
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaExistentialAllOutputsNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	y, _ := m.AppendAnd(p.Not(), b)
	m.AppendCO(x)
	m.AppendCO(y)
	m.AppendCO(a.Not())
	m.SetRegisterCount(2)
	g, e := m.ExistentialAllOutputs(0, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_quantify_all.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct {
		Objects int
		Masks   []uint
	}
	if e = json.Unmarshal(raw, &native); e != nil {
		t.Fatal(e)
	}
	masks := make([]uint, 3)
	for k := 0; k < 8; k++ {
		v, e := g.Evaluate([]bool{k&1 != 0, k&2 != 0, k&4 != 0})
		if e != nil {
			t.Fatal(e)
		}
		for i, value := range v {
			if value {
				masks[i] |= 1 << k
			}
		}
	}
	if g.ObjectCount() != native.Objects || !slices.Equal(masks, native.Masks) {
		t.Fatal(masks, native, g.ObjectCount())
	}
}

func TestGiaExistentialAllOutputsLimits(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	m.AppendCO(x)
	m.AppendCO(p.Not())
	m.SetRegisterCount(1)
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.ExistentialAllOutputs(0, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 3, 5, 7} {
		calls := 0
		g, e := m.ExistentialAllOutputs(0, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
