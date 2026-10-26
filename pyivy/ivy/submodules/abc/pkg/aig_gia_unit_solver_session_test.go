package abc

import (
	"encoding/json"
	"os"
	"testing"
)

func unitReuseFixture(t *testing.T, mode int) (*GiaManager, *GiaLUTNetwork) {
	t.Helper()
	m := NewGiaManager()
	var ci []GiaLit
	for i := 0; i < 8; i++ {
		l, _ := m.AppendCI()
		ci = append(ci, l)
	}
	var parts [2][]GiaLit
	for c := 0; c < 2; c++ {
		sig := append([]GiaLit(nil), ci[4*c:4*c+4]...)
		a, _ := m.AppendAnd(sig[0], sig[1])
		sig = append(sig, a)
		b, _ := m.AppendAnd(sig[2], sig[3])
		sig = append(sig, b)
		x, _ := m.AppendAnd(sig[4], sig[5])
		sig = append(sig, x)
		for i := 3; i < 10+c; i++ {
			l, e := m.AppendAnd(sig[3+i], sig[(i+c)%4].NotIf(mode != 0 && i&1 != 0))
			if e != nil {
				t.Fatal(e)
			}
			sig = append(sig, l)
		}
		parts[c] = sig
	}
	for c := 0; c < 2; c++ {
		m.AppendCO(parts[c][8])
		m.AppendCO(parts[c][13+c])
	}
	var records []GiaLUTRecord
	for _, sig := range parts {
		for _, l := range sig[4:] {
			a, _ := m.objects[l.ID()].Fanin0(l.ID())
			b, _ := m.objects[l.ID()].Fanin1(l.ID())
			records = append(records, GiaLUTRecord{Root: l.ID(), Fanins: []GiaID{a.ID(), b.ID()}})
		}
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
func TestGiaUnitSolverReuseNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_unit_remap_reuse.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Mode, Reverse, ImprovementLimit, Count int
		Fanins                                 [][]GiaID
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 12 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		m, n := unitReuseFixture(t, row.Mode)
		r, e := n.RemapUnitPass(m, 0, GiaUnitRemapPassOptions{WindowLimit: 16, Reverse: row.Reverse != 0, MaxImprovements: row.ImprovementLimit})
		if e != nil || len(r.Network.Roots()) != row.Count || r.SolverBuilds != 1 {
			t.Fatal(i, r, e)
		}
		if row.ImprovementLimit != 1 && r.WindowResets < 1 {
			t.Fatal("solver was not reused", r)
		}
		if r.DiscoveryBuilds != 1 || r.DiscoveryQueries != len(r.Attempts) || r.ReferenceRefreshes < 1 {
			t.Fatal("discovery reuse", r)
		}
		if row.ImprovementLimit != 1 && r.ReferenceRefreshes < 2 {
			t.Fatal("references were not refreshed", r)
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
		native := &GiaLUTNetwork{fanins: row.Fanins}
		functions, e := native.ValidateFunctions(m, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkGiaLUTComposition(t, m, native, functions)
	}
}
func TestGiaUnitSolverSessionIsolation(t *testing.T) {
	session := &giaUnitWindowSolverSession{log: 2}
	cuts := []GiaWindowCut{{Node: 0}, {Node: 0}}
	ranges := []GiaWindowCutRange{{Count: 2}}
	s, e := session.prepare(1, cuts, ranges, []int{0}, GiaWindowCNFOptions{})
	if e != nil {
		t.Fatal(e)
	}
	base := s.solver.bookmark
	if _, e = s.solver.AddClause([]SatLit{1}); e != nil {
		t.Fatal(e)
	}
	r, e := s.Solve(1, SatLimits{})
	if e != nil || r.SAT.Status != SatUnsatisfiable {
		t.Fatal(r, e)
	}
	s, e = session.prepare(1, cuts, ranges, []int{0}, GiaWindowCNFOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if s.solver.bookmark != base || session.builds != 1 || session.resets != 1 {
		t.Fatal(session)
	}
	r, e = s.Solve(1, SatLimits{})
	if e != nil || r.SAT.Status != SatSatisfiable {
		t.Fatal("stale exclusion", r, e)
	}
	if _, e = session.prepare(1, cuts, ranges, []int{0}, GiaWindowCNFOptions{Stop: func() bool { return true }}); e == nil {
		t.Fatal("cancel")
	}
	if session.resets != 1 {
		t.Fatal("failed reset counted")
	}
	r, e = s.Solve(1, SatLimits{})
	if e != nil || r.SAT.Status != SatSatisfiable {
		t.Fatal(r, e)
	}
}
func FuzzGiaUnitSolverReuse(f *testing.F) {
	f.Add(uint8(0), true, uint8(0))
	f.Fuzz(func(t *testing.T, mode uint8, reverse bool, cap uint8) {
		m, n := unitReuseFixture(t, int(mode%2))
		r, e := n.RemapUnitPass(m, 0, GiaUnitRemapPassOptions{WindowLimit: 16, Reverse: reverse, MaxImprovements: int(cap % 3)})
		if e != nil || r.SolverBuilds != 1 {
			t.Fatal(r, e)
		}
		if cap%3 != 1 && r.WindowResets == 0 {
			t.Fatal("no reuse")
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
	})
}
