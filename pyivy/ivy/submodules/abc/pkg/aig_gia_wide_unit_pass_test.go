package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func wideUnitFixture(t *testing.T, vars, mode int) (*GiaManager, *GiaLUTNetwork) {
	t.Helper()
	m := NewGiaManager()
	var ci, gates []GiaLit
	for i := 0; i < vars; i++ {
		l, _ := m.AppendCI()
		ci = append(ci, l)
	}
	l, _ := m.AppendAnd(ci[0], ci[1])
	gates = append(gates, l)
	for i := 1; i < 10; i++ {
		l, e := m.AppendAnd(gates[i-1], ci[(i+1)%vars])
		if e != nil {
			t.Fatal(e)
		}
		gates = append(gates, l)
	}
	m.AppendCO(gates[4])
	m.AppendCO(gates[9])
	var records []GiaLUTRecord
	for _, l := range gates {
		a, _ := m.objects[l.ID()].Fanin0(l.ID())
		b, _ := m.objects[l.ID()].Fanin1(l.ID())
		fanins := []GiaID{a.ID(), b.ID()}
		if mode != 0 {
			fanins = nil
			for _, c := range ci {
				fanins = append(fanins, c.ID())
			}
		}
		records = append(records, GiaLUTRecord{Root: l.ID(), Fanins: fanins})
	}
	mapping, e := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if e != nil {
		t.Fatal(e)
	}
	n, e := mapping.AlternateNetwork(GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	return m, n
}
func TestGiaWideUnitPassNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_wide_unit_pass.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Variables, Width, Mode, Reverse, ImprovementLimit, Count int
		Fanins                                                   [][]GiaID
		Truths                                                   [][2]uint64
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 64 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		m, n := wideUnitFixture(t, row.Variables, row.Mode)
		original, e := n.UnitTiming(m, 0, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		r, e := n.RemapUnitPass(m, 0, GiaUnitRemapPassOptions{WindowLimit: 16, Reverse: row.Reverse != 0, MaxImprovements: row.ImprovementLimit, Search: GiaUnitWindowSearchOptions{LutSize: row.Width}})
		if e != nil || len(r.Network.Roots()) != row.Count || r.Timing.Maximum > original.Maximum {
			t.Fatal(i, r, e, row.Count)
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
		native := &GiaLUTNetwork{fanins: row.Fanins}
		functions, e := native.ValidateFunctions(m, GiaLUTAnalysisOptions{})
		if e != nil || len(functions) != len(row.Truths) {
			t.Fatal(i, e)
		}
		for j, f := range functions {
			if uint64(f.Root) != row.Truths[j][0] || f.Truth != row.Truths[j][1] {
				t.Fatal(i, f, row.Truths[j])
			}
		}
		checkGiaLUTComposition(t, m, native, functions)
	}
}
func TestGiaWideTruthSlotsAndEdgeLimits(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendBuffer(a)
	m.AppendCO(b)
	n := &GiaLUTNetwork{fanins: make([][]GiaID, m.ObjectCount())}
	n.fanins[b.ID()] = []GiaID{0, a.ID(), 0, a.ID(), 0, a.ID()}
	functions, e := n.ValidateFunctions(m, GiaLUTAnalysisOptions{})
	if e != nil || len(functions) != 1 || functions[0].Truth != 0xffffffff00000000 {
		t.Fatal(functions, e)
	}
	checkGiaLUTComposition(t, m, n, functions)
	n.fanins[b.ID()] = []GiaID{a.ID(), 0, a.ID(), 0, a.ID()}
	functions, e = n.ValidateFunctions(m, GiaLUTAnalysisOptions{})
	if e != nil || functions[0].Truth != 0xffff0000ffff0000 {
		t.Fatal(functions, e)
	}
	graph, network := unitPassFixture(t, 0)
	prior := &GiaEdgeAssignment{rows: make([][2]GiaID, graph.ObjectCount())}
	r, e := network.RemapEdgePass(graph, prior, 0, GiaUnitRemapPassOptions{Search: GiaUnitWindowSearchOptions{LutSize: 6}})
	if !errors.Is(e, ErrUnsupported) || r.Network != nil {
		t.Fatal(r, e)
	}
	for _, width := range []int{-1, 7} {
		v, e := network.RemapUnitPass(graph, 0, GiaUnitRemapPassOptions{WindowLimit: 1, Search: GiaUnitWindowSearchOptions{LutSize: width}})
		if !errors.Is(e, ErrDimension) || v.Network != nil {
			t.Fatal(v, e)
		}
	}
}
func TestGiaWideBoxUnitPacked(t *testing.T) {
	m := NewGiaManager()
	var ci []GiaLit
	for i := 0; i < 7; i++ {
		l, _ := m.AppendCI()
		ci = append(ci, l)
	}
	var parts [2][]GiaLit
	for c := 0; c < 2; c++ {
		inputs := ci[:6]
		if c == 1 {
			inputs = []GiaLit{ci[6], ci[0], ci[1], ci[2], ci[3], ci[4]}
		}
		l, _ := m.AppendAnd(inputs[0], inputs[1])
		parts[c] = append(parts[c], l)
		for i := 1; i < 10; i++ {
			l, e := m.AppendAnd(parts[c][i-1], inputs[(i+1)%6])
			if e != nil {
				t.Fatal(e)
			}
			parts[c] = append(parts[c], l)
		}
	}
	m.AppendCO(parts[0][9])
	m.AppendCO(ci[0])
	m.AppendCO(parts[1][9])
	var records []GiaLUTRecord
	for _, part := range parts {
		for _, l := range part {
			a, _ := m.objects[l.ID()].Fanin0(l.ID())
			b, _ := m.objects[l.ID()].Fanin1(l.ID())
			records = append(records, GiaLUTRecord{Root: l.ID(), Fanins: []GiaID{a.ID(), b.ID()}})
		}
	}
	mapping, e := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if e != nil {
		t.Fatal(e)
	}
	manager, _ := NewTimManager(7, 3, 0)
	manager.AddBox(TimBox{Inputs: 2, FirstOutput: 6, Outputs: 1}, TimBoxOptions{})
	manager.SetDelayTable(0, TimDelayTable{Inputs: 2, Outputs: 1, Delays: []float32{1.25, 0.5}}, TimDelayOptions{})
	before := timSnapshot(manager)
	r, e := mapping.RemapUnitPassWithBoxes(m, manager, 100, GiaUnitRemapPassOptions{WindowLimit: 16, Search: GiaUnitWindowSearchOptions{LutSize: 6}})
	if e != nil || len(r.Mapping.Roots()) != 2 {
		t.Fatal(r, e)
	}
	for _, id := range r.Mapping.Roots() {
		record, _ := r.Mapping.Record(id)
		if len(record.Fanins) != 6 {
			t.Fatal(record)
		}
	}
	checkGiaLUTComposition(t, m, r.Pass.Pass.Network, r.Pass.Pass.Functions)
	if !slices.Equal(before, timSnapshot(manager)) {
		t.Fatal("source changed")
	}
}
func FuzzGiaWideUnitPass(f *testing.F) {
	f.Add(uint8(0), uint8(6), true)
	f.Fuzz(func(t *testing.T, kind, width uint8, reverse bool) {
		m, n := wideUnitFixture(t, 5+int(kind%2), int(kind/2%2))
		before, e := n.UnitTiming(m, 0, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		packed, e := n.PackedMapping(0, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		r, e := packed.RemapUnitPass(m, 0, GiaUnitRemapPassOptions{WindowLimit: 16, Reverse: reverse, Search: GiaUnitWindowSearchOptions{LutSize: 1 + int(width%6)}})
		if e != nil || r.Pass.Timing.Maximum > before.Maximum {
			t.Fatal(r, e)
		}
		checkGiaLUTComposition(t, m, r.Pass.Network, r.Pass.Functions)
	})
}
