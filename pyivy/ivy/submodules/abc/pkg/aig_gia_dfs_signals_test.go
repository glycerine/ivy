package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaEncounterAndSignalExtraction(t *testing.T) {
	m := NewGiaManager()
	m.Name = "dfs"
	m.Spec = "dfs.aig"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(c, b)
	buf, _ := m.AppendBuffer(a.Not())
	mux, _ := m.AppendMuxReal(buf, x, a)
	m.AppendCO(mux)
	m.AppendCO(b)
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	r, e := m.ExtractEncounterOrderedCone(mux.Not(), GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || r.Graph.registers != 0 || r.Graph.constraints != 0 || r.Graph.XorCount() != 1 || r.Graph.BufferCount() != 1 || r.Graph.MuxCount() != 1 || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/source")
	}
	signals := []GiaLit{mux.Not(), a, GiaConst1, x, x.Not()}
	g, e := m.ExtractSignalsRetainingInputs(signals, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	for k := 0; k < 8; k++ {
		source := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
		av, bv, cv := source[0], source[1], source[2]
		value := av
		if !av {
			value = cv != bv
		}
		projected := make([]bool, len(r.Inputs))
		for i, index := range r.Inputs {
			projected[i] = source[index]
		}
		v, e := r.Graph.Evaluate(projected)
		if e != nil || v[0] == value {
			t.Fatal(k, v, value, e)
		}
		got, e := g.Evaluate(source)
		want := []bool{!value, av, true, cv != bv, cv == bv}
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(k, got, want, e)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.ExtractEncounterOrderedCone(mux, options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
		g, e := m.ExtractSignalsRetainingInputs(signals, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 3, 6} {
		calls := 0
		r, e := m.ExtractEncounterOrderedCone(mux, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	constant, e := m.ExtractEncounterOrderedCone(GiaConst1, GiaCofactorOptions{})
	if e != nil || len(constant.Inputs) != 0 {
		t.Fatal(constant, e)
	}
	g, e = m.ExtractSignalsRetainingInputs(nil, GiaCofactorOptions{})
	if e != nil || g.InputCount() != 3 || g.OutputCount() != 0 {
		t.Fatal(e)
	}
}
func FuzzGiaSignalDFSExtraction(f *testing.F) {
	f.Add(uint8(5), uint8(7))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(c, b)
		y, _ := m.AppendMuxReal(a, x, c)
		root := y.NotIf(phase&1 != 0)
		r, e := m.ExtractEncounterOrderedCone(root, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		source := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		projected := make([]bool, len(r.Inputs))
		for i, index := range r.Inputs {
			projected[i] = source[index]
		}
		want := source[2]
		if source[0] {
			want = source[2] != source[1]
		}
		want = want != (phase&1 != 0)
		got, e := r.Graph.Evaluate(projected)
		if e != nil || got[0] != want {
			t.Fatal(got, want, e)
		}
		g, e := m.ExtractSignalsRetainingInputs([]GiaLit{root, c.Not()}, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		v, e := g.Evaluate(source)
		if e != nil || !slices.Equal(v, []bool{want, !source[2]}) {
			t.Fatal(v, e)
		}
	})
}

func TestGiaSignalDFSCopyNative(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	d, _ := m.AppendCI()
	x, _ := m.AppendAnd(b, c)
	y, _ := m.AppendAnd(a, d)
	root, _ := m.AppendAnd(x, y)
	encounter, e := m.ExtractEncounterOrderedCone(root, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if !slices.Equal(encounter.Inputs, []int{1, 2, 0, 3}) {
		t.Fatal(encounter.Inputs)
	}
	signals, e := m.ExtractSignalsRetainingInputs([]GiaLit{root.Not(), x, GiaConst1, y.Not(), x}, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_dfs_signals.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects int
		Cis     []GiaID
		Drivers []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 2 {
		t.Fatal(e)
	}
	for mode, g := range []*GiaManager{encounter.Graph, signals} {
		row := rows[mode]
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if g.ObjectCount() != row.Objects || !slices.Equal(g.cis, row.Cis) || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(mode, g.cis, drivers, row)
		}
	}
}

func TestGiaSignalListCancellationAndInvalidRoots(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	co, _ := m.AppendCO(x)
	for _, signals := range [][]GiaLit{{InvalidGiaLit}, {co}} {
		if g, e := m.ExtractSignalsRetainingInputs(signals, GiaCofactorOptions{}); g != nil || e == nil {
			t.Fatal(signals, e)
		}
	}
	for _, threshold := range []int{1, 3, 5} {
		calls := 0
		g, e := m.ExtractSignalsRetainingInputs([]GiaLit{x, x.Not()}, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
