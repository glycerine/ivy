package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func unitPassFixture(t *testing.T, mode int) (*GiaManager, *GiaLUTNetwork) {
	t.Helper()
	m := NewGiaManager()
	var sig []GiaLit
	for i := 0; i < 4; i++ {
		l, _ := m.AppendCI()
		sig = append(sig, l)
	}
	a, _ := m.AppendAnd(sig[0], sig[1])
	sig = append(sig, a)
	b, _ := m.AppendAnd(sig[2], sig[3])
	sig = append(sig, b)
	c, _ := m.AppendAnd(sig[4], sig[5])
	sig = append(sig, c)
	for i := 3; i < 10; i++ {
		a := sig[3+i]
		b := sig[(i+mode)%4]
		if mode == 1 {
			b = sig[4+i/2]
		}
		if a == b {
			b = sig[0]
		}
		l, e := m.AppendAnd(a, b.NotIf(mode == 2 && i&1 != 0))
		if e != nil {
			t.Fatal(e)
		}
		sig = append(sig, l)
	}
	m.AppendCO(sig[8])
	m.AppendCO(sig[13])
	var records []GiaLUTRecord
	for _, l := range sig[4:] {
		a, _ := m.objects[l.ID()].Fanin0(l.ID())
		b, _ := m.objects[l.ID()].Fanin1(l.ID())
		records = append(records, GiaLUTRecord{Root: l.ID(), Fanins: []GiaID{a.ID(), b.ID()}})
	}
	mapping, e := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if e != nil {
		t.Fatal(e)
	}
	n, e := mapping.Network(GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	return m, n
}
func TestGiaUnitRemapPassNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_unit_remap_pass.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Mode, Limit, Reverse, ImprovementLimit, Count int
		Fanins                                        [][]GiaID
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 36 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		m, n := unitPassFixture(t, row.Mode)
		before := slices.Clone(m.objects)
		original, e := n.UnitTiming(m, 0, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		r, e := n.RemapUnitPass(m, 0, GiaUnitRemapPassOptions{WindowLimit: row.Limit, Reverse: row.Reverse != 0, MaxImprovements: row.ImprovementLimit})
		if e != nil {
			t.Fatal(i, e)
		}
		if len(r.Network.Roots()) != row.Count || r.Timing.Maximum > original.Maximum {
			t.Fatal(i, r, row.Count)
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
		if !slices.Equal(m.objects, before) || len(n.Roots()) != 10 {
			t.Fatal("source changed")
		}
		native := &GiaLUTNetwork{fanins: row.Fanins}
		functions, e := native.ValidateFunctions(m, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal("native result", i, e)
		}
		checkGiaLUTComposition(t, m, native, functions)
		for j, attempt := range r.Attempts {
			if j > 0 && attempt.Pivot <= r.Attempts[j-1].Pivot {
				t.Fatal("pivot order", r.Attempts)
			}
		}
		if row.ImprovementLimit == 1 && r.Improvements > 1 {
			t.Fatal("improvement limit")
		}
	}
}
func TestGiaUnitRemapPassLimits(t *testing.T) {
	m, n := unitPassFixture(t, 0)
	for _, o := range []GiaUnitRemapPassOptions{{WindowLimit: -1}, {WindowLimit: 129}, {MaxImprovements: -1}, {MaxPivots: -1}, {MaxWindows: -1}, {Search: GiaUnitWindowSearchOptions{Cuts: GiaWindowCutOptions{MaxCuts: -1}}}, {MaxPivots: 1, WindowLimit: 3}, {MaxWindows: 1, WindowLimit: 3}, {Search: GiaUnitWindowSearchOptions{Stop: func() bool { return true }}}} {
		r, e := n.RemapUnitPass(m, 0, o)
		if e == nil || r.Network != nil || r.Attempts != nil {
			t.Fatal(o, r, e)
		}
	}
	r, e := n.RemapUnitPass(m, 0, GiaUnitRemapPassOptions{WindowLimit: 3})
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for _, a := range r.Attempts {
		if a.Reason == "repeated" {
			found = true
		}
	}
	if !found {
		t.Fatal("missing repeated-window suppression", r.Attempts)
	}
}
func FuzzGiaUnitRemapPass(f *testing.F) {
	f.Add(uint8(0), uint8(16), true, uint8(0))
	f.Fuzz(func(t *testing.T, mode, limit uint8, reverse bool, cap uint8) {
		m, n := unitPassFixture(t, int(mode%3))
		before, _ := n.UnitTiming(m, 0, GiaLUTAnalysisOptions{})
		r, e := n.RemapUnitPass(m, 0, GiaUnitRemapPassOptions{WindowLimit: 1 + int(limit%20), Reverse: reverse, MaxImprovements: int(cap % 3)})
		if e != nil {
			t.Fatal(e)
		}
		if r.Timing.Maximum > before.Maximum || len(r.Network.Roots()) > len(n.Roots()) {
			t.Fatal(r)
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
	})
}
