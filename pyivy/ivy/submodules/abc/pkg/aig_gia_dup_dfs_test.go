package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaDuplicateDFSAndSelectedState(t *testing.T) {
	m := NewGiaManager()
	m.Name = "dfs-copy"
	m.Spec = "dfs-copy.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, b)
	buf, _ := m.AppendBuffer(a.Not())
	mux, _ := m.AppendMuxReal(b, x, p)
	dangling, _ := m.AppendAnd(p, a)
	for _, l := range []GiaLit{mux, buf, GiaConst0, a.Not()} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for index := -1; index < 3; index++ {
		var g *GiaManager
		var mapping []GiaLit
		var e error
		if index < 0 {
			g, mapping, e = m.DuplicateDFS(GiaCofactorOptions{})
		} else {
			g, mapping, e = m.SelectPropertyWithState(index, GiaCofactorOptions{})
		}
		if e != nil {
			t.Fatal(e)
		}
		expectedConstraints := 0
		if index < 0 {
			expectedConstraints = 1
		}
		if g.registers != 1 || g.constraints != expectedConstraints || g.InputCount() != 3 || g.Name != m.Name || g.Spec != m.Spec || mapping[dangling.ID()] != InvalidGiaLit || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source/map")
		}
		for k := 0; k < 8; k++ {
			input := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
			want, e := m.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			if index >= 0 {
				want = []bool{want[index], want[3]}
			}
			got, e := g.Evaluate(input)
			if e != nil || !slices.Equal(got, want) {
				t.Fatal(index, k, got, want, e)
			}
		}
		if index >= 0 {
			for po := 0; po < 3; po++ {
				if po != index && mapping[m.cos[po]] != InvalidGiaLit {
					t.Fatal("dropped CO map")
				}
			}
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, mapping, e := m.DuplicateDFS(options)
		if g != nil || mapping != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 12} {
		calls := 0
		g, mapping, e := m.DuplicateDFS(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || mapping != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	for _, index := range []int{-1, 3} {
		if _, _, e := m.SelectPropertyWithState(index, GiaCofactorOptions{}); e == nil {
			t.Fatal(index)
		}
	}
}
func FuzzGiaDFSStateSelection(f *testing.F) {
	f.Add(uint8(2), uint8(7))
	f.Fuzz(func(t *testing.T, selected, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		for _, l := range []GiaLit{x, b.Not(), p, a.Not(), b} {
			m.AppendCO(l)
		}
		m.SetRegisterCount(2)
		index := int(selected % 3)
		g, _, e := m.SelectPropertyWithState(index, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		pv, av, bv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		source := []bool{pv != av, !bv, pv}
		want := []bool{source[index], !av, bv}
		got, e := g.Evaluate([]bool{pv, av, bv})
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaDFSDuplicationNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, b)
	y, _ := m.AppendAnd(a, b)
	m.AppendAnd(p, a)
	for _, l := range []GiaLit{x, y, GiaConst0, a.Not()} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	raw, e := os.ReadFile("test_vectors/oracle/gia_dup_dfs.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects, Constraints int
		Drivers              []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 4 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		var g *GiaManager
		var e error
		if mode == 0 {
			g, _, e = m.DuplicateDFS(GiaCofactorOptions{})
		} else {
			g, _, e = m.SelectPropertyWithState(mode-1, GiaCofactorOptions{})
		}
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if g.ObjectCount() != row.Objects || g.constraints != row.Constraints || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(mode, drivers, row)
		}
	}
}
