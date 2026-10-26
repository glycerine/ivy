package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaInputRemappingOrderedSubstitution(t *testing.T) {
	m := NewGiaManager()
	m.Name = "remap"
	m.Spec = "remap.aig"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	buf, _ := m.AppendBuffer(c.Not())
	mux, _ := m.AppendMuxReal(c, x, a)
	m.AppendCO(mux)
	m.AppendCO(buf)
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	replacements := []GiaLit{a.Not(), a.Not(), b.Not()}
	g, e := m.RemapInputs(replacements, nil, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if g.registers != 1 || g.constraints != 1 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/source")
	}
	for k := 0; k < 8; k++ {
		input := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
		av := !input[0]
		bv := !av
		cv := !bv
		main := av
		if cv {
			main = av != bv
		}
		want := []bool{main, !cv}
		got, e := g.Evaluate(input)
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(k, got, want, e)
		}
	}
	custom, e := m.RemapInputs([]GiaLit{InvalidGiaLit, GiaConst1, GiaConst0}, []GiaLit{a.Not(), x, buf}, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if custom.constraints != 0 || custom.registers != 1 {
		t.Fatal("custom roles")
	}
	for k := 0; k < 8; k++ {
		input := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
		v, e := custom.Evaluate(input)
		if e != nil || !slices.Equal(v, []bool{!input[0], !input[0], true}) {
			t.Fatal(v, e)
		}
	}
	for _, r := range [][]GiaLit{nil, {b, InvalidGiaLit, InvalidGiaLit}, {x, InvalidGiaLit, InvalidGiaLit}, {GiaLit(999), InvalidGiaLit, InvalidGiaLit}} {
		if _, e := m.RemapInputs(r, nil, GiaCofactorOptions{}); e == nil {
			t.Fatal(r)
		}
	}
	if _, e := m.RemapInputs(replacements, []GiaLit{}, GiaCofactorOptions{}); e == nil {
		t.Fatal("missing RIs")
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.RemapInputs(replacements, nil, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 3, 6, 8} {
		calls := 0
		g, e := m.RemapInputs(replacements, nil, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaInputRemapping(f *testing.F) {
	f.Add(uint8(5), uint8(7))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(a, b)
		y, _ := m.AppendMuxReal(c, x, a)
		m.AppendCO(y)
		r := []GiaLit{a.NotIf(phase&1 != 0), a.NotIf(phase&2 != 0), b.NotIf(phase&4 != 0)}
		g, e := m.RemapInputs(r, nil, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		av := (assignment&1 != 0) != (phase&1 != 0)
		bv := av != (phase&2 != 0)
		cv := bv != (phase&4 != 0)
		want := av
		if cv {
			want = av != bv
		}
		got, e := g.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
		if e != nil || got[0] != want {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaInputRemappingNative(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(x, c)
	m.AppendCO(y)
	m.AppendCO(x.Not())
	raw, e := os.ReadFile("test_vectors/oracle/gia_remap_inputs.json")
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
		var roots []GiaLit
		if mode != 0 {
			roots = []GiaLit{a.Not(), y.Not()}
		}
		g, e := m.RemapInputs([]GiaLit{InvalidGiaLit, a, GiaConst1}, roots, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if g.ObjectCount() != row.Objects || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(mode, drivers, row)
		}
	}
}
