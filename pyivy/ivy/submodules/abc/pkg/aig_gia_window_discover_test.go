package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func discoveryFixture(t *testing.T, mode, mm int) (*GiaManager, *GiaLUTNetwork, []GiaLit) {
	t.Helper()
	m := NewGiaManager()
	var signals []GiaLit
	for i := 0; i < 4; i++ {
		l, _ := m.AppendCI()
		signals = append(signals, l)
	}
	for i := 0; i < 8; i++ {
		a := signals[0]
		if i > 0 {
			a = signals[3+i]
		}
		b := signals[(i*3+1)%4]
		if mode == 1 && i > 2 {
			b = signals[4+i/2]
		}
		if mode == 2 {
			a = signals[(i*7+2)%(4+i)]
			b = signals[(i*11+1)%(4+i)]
		}
		if a == b {
			b = signals[(i+1)%4]
		}
		if a == b {
			b = signals[(i+2)%4]
		}
		l, e := m.AppendAnd(a, b)
		if e != nil {
			t.Fatal(e)
		}
		signals = append(signals, l)
	}
	m.AppendCO(signals[11])
	m.AppendCO(signals[7])
	var records []GiaLUTRecord
	for i := 0; i < 8; i++ {
		id := signals[4+i].ID()
		a, _ := m.objects[id].Fanin0(id)
		b, _ := m.objects[id].Fanin1(id)
		f := []GiaID{a.ID(), b.ID()}
		if mm != 0 && i&1 != 0 {
			f = []GiaID{1, 2, 3, 4}
		}
		records = append(records, GiaLUTRecord{Root: id, Fanins: f})
	}
	mapping, e := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if e != nil {
		t.Fatal(e)
	}
	n, e := mapping.Network(GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	return m, n, signals
}
func TestGiaWindowDiscoverNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_window_discover.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Mode, Mapping, Reverse, Limit, Pivot, Count int
		Roots, Nodes, Leaves, Ands                  []GiaID
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 480 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		m, n, s := discoveryFixture(t, row.Mode, row.Mapping)
		before := slices.Clone(m.objects)
		r, e := n.DiscoverWindow(m, s[4+row.Pivot].ID(), row.Limit, row.Reverse != 0, GiaLUTAnalysisOptions{})
		if e != nil || r.Found != (row.Count != 0) || len(r.Ands) != row.Count || !slices.Equal(r.Roots, row.Roots) || !slices.Equal(r.Nodes, row.Nodes) || !slices.Equal(r.Leaves, row.Leaves) || !slices.Equal(r.Ands, row.Ands) {
			t.Fatal(i, row, r, e)
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("graph changed")
		}
	}
}
func TestGiaWindowDiscoverSafety(t *testing.T) {
	m, n, s := discoveryFixture(t, 0, 0)
	for _, o := range []GiaLUTAnalysisOptions{{MaxObjects: 1}, {MaxVisits: 1}, {MaxObjects: -1}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		r, e := n.DiscoverWindow(m, s[11].ID(), 8, false, o)
		if e == nil || r.Found || r.Ands != nil {
			t.Fatal(r, e)
		}
	}
	for _, id := range []GiaID{0, 1, GiaID(m.ObjectCount())} {
		if _, e := n.DiscoverWindow(m, id, 8, false, GiaLUTAnalysisOptions{}); e == nil {
			t.Fatal(id)
		}
	}
	if _, e := n.DiscoverWindow(m, s[11].ID(), -1, false, GiaLUTAnalysisOptions{}); e == nil {
		t.Fatal("negative limit")
	}
}
func FuzzGiaWindowDiscover(f *testing.F) {
	f.Add(uint8(0), uint8(0), uint8(8), true)
	f.Fuzz(func(t *testing.T, mode, pivot, limit uint8, reverse bool) {
		m, n, s := discoveryFixture(t, int(mode%3), int(mode/3%2))
		r, e := n.DiscoverWindow(m, s[4+int(pivot%8)].ID(), int(limit%12), reverse, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if !r.Found {
			return
		}
		if len(r.Ands) > int(limit%12) || !slices.IsSorted(r.Roots) || !slices.IsSorted(r.Ands) || !slices.IsSorted(r.Nodes) || !slices.IsSorted(r.Leaves) {
			t.Fatal(r)
		}
		_, e = m.GenerateWindowCuts(r.Leaves, r.Ands, 4, GiaWindowCutOptions{})
		if e != nil {
			t.Fatal(r, e)
		}
		result, e := n.RemapUnitWindow(m, r.Leaves, r.Ands, r.Roots, 0, GiaUnitWindowSearchOptions{})
		if e != nil {
			t.Fatal(r, e)
		}
		if result.Replacement.Network != nil {
			checkGiaLUTComposition(t, m, result.Replacement.Network, result.Replacement.Functions)
		}
	})
}
