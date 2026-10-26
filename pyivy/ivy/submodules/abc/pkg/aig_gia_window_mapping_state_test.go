package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func mappingStateFixture(t *testing.T, nn, mode, width int) (*GiaManager, *GiaLUTNetwork, []GiaID, []GiaID, GiaWindowCuts) {
	t.Helper()
	m := NewGiaManager()
	var all []GiaLit
	var leaves, nodes []GiaID
	for i := 0; i < 5; i++ {
		l, _ := m.AppendCI()
		all = append(all, l)
		if i < 4 {
			leaves = append(leaves, l.ID())
		}
	}
	for i := 0; i < nn; i++ {
		a := all[0]
		if i > 0 {
			a = all[5+i-1]
		}
		l, e := m.AppendAnd(a, all[(i*3+1)%4])
		if e != nil {
			t.Fatal(e)
		}
		all = append(all, l)
		nodes = append(nodes, l.ID())
	}
	n := &GiaLUTNetwork{fanins: make([][]GiaID, len(m.objects)), fanouts: make([][]GiaID, len(m.objects))}
	for _, id := range nodes {
		if mode == 1 {
			n.fanins[id] = slices.Clone(leaves)
		} else {
			a, _ := m.objects[id].Fanin0(id)
			b, _ := m.objects[id].Fanin1(id)
			n.fanins[id] = []GiaID{a.ID(), b.ID()}
			if mode == 2 {
				n.fanins[id] = append(n.fanins[id], all[4].ID())
			}
		}
	}
	cuts, e := m.GenerateWindowCuts(leaves, nodes, width, GiaWindowCutOptions{})
	if e != nil {
		t.Fatal(e)
	}
	return m, n, leaves, nodes, cuts
}
func TestGiaWindowMappingStateNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_window_mapping_state.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Nodes, Mode, Width, First int
		Roots, Initial, Preferred []int
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 18 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m, n, leaves, nodes, cuts := mappingStateFixture(t, row.Nodes, row.Mode, row.Width)
		r, e := n.WindowMappingState(m, leaves, nodes, []GiaID{leaves[0], nodes[len(nodes)-1], nodes[len(nodes)-1]}, cuts, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		var preferred []int
		for i, node := range r.InitialNodes {
			preferred = append(preferred, node)
			if r.PreferredCuts[i] >= 0 {
				preferred = append(preferred, row.First+r.PreferredCuts[i])
			}
		}
		if !slices.Equal(r.RootVariables, row.Roots) || !slices.Equal(r.InitialNodes, row.Initial) || !slices.Equal(preferred, row.Preferred) {
			t.Fatal(row, r, preferred)
		}
		s, e := NewGiaWindowSolver(4, len(nodes), cuts.Cuts, cuts.Ranges, r.RootVariables, GiaWindowCNFOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if e = s.SetMappingPreferences(r); e != nil {
			t.Fatal(e)
		}
		for v, p := range s.solver.polarity {
			if p != slices.Contains(preferred, v) {
				t.Fatal(v, p)
			}
		}
		a, e := s.Solve(len(nodes), SatLimits{})
		if e != nil || a.SAT.Status != SatSatisfiable {
			t.Fatal(a, e)
		}
	}
}
func TestGiaWindowMappingStateSafety(t *testing.T) {
	m, n, l, ids, c := mappingStateFixture(t, 3, 0, 4)
	for _, o := range []GiaLUTAnalysisOptions{{MaxObjects: 1}, {MaxVisits: 1}, {MaxObjects: -1}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		r, e := n.WindowMappingState(m, l, ids, nil, c, o)
		if e == nil || r.InitialNodes != nil {
			t.Fatal(r, e)
		}
	}
	bad := GiaWindowCuts{Cuts: c.Cuts, Ranges: slices.Clone(c.Ranges)}
	bad.Ranges[0].Count = len(c.Cuts) + 1
	if _, e := n.WindowMappingState(m, l, ids, nil, bad, GiaLUTAnalysisOptions{}); e == nil {
		t.Fatal("range")
	}
	if _, e := n.WindowMappingState(m, append(slices.Clone(l), ids[0]), ids, nil, c, GiaLUTAnalysisOptions{}); e == nil {
		t.Fatal("duplicate")
	}
	s := NewSatSolver()
	s.SetVariableCount(3)
	if e := s.SetPreferredVariables([]SatVar{1, 1}); e != nil {
		t.Fatal(e)
	}
	if e := s.SetPreferredVariables([]SatVar{0, 3}); e == nil || !s.polarity[1] || s.polarity[0] {
		t.Fatal("atomic hints")
	}
	if e := s.SetPreferredVariables(nil); e != nil || s.polarity[1] {
		t.Fatal("reset")
	}
}
func FuzzGiaWindowMappingState(f *testing.F) {
	f.Add(uint8(3), uint8(2))
	f.Fuzz(func(t *testing.T, size, mode uint8) {
		m, n, l, ids, c := mappingStateFixture(t, 1+int(size%12), int(mode%3), 4)
		r, e := n.WindowMappingState(m, l, ids, []GiaID{ids[len(ids)-1]}, c, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if len(r.InitialNodes) != len(ids) || !slices.Equal(r.RootVariables, []int{len(ids) - 1}) {
			t.Fatal(r)
		}
		for i, k := range r.PreferredCuts {
			if k >= 0 && (c.Cuts[k].Node != i || k < c.Ranges[i].Start || k >= c.Ranges[i].Start+c.Ranges[i].Count) {
				t.Fatal(i, k)
			}
		}
	})
}
