package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func sameGiaWindow(a, b GiaDiscoveredWindow) bool {
	return a.Found == b.Found && slices.Equal(a.Roots, b.Roots) && slices.Equal(a.Nodes, b.Nodes) && slices.Equal(a.Leaves, b.Leaves) && slices.Equal(a.Ands, b.Ands)
}
func TestGiaDiscoverySessionNative(t *testing.T) {
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
	key := -1
	var m *GiaManager
	var n *GiaLUTNetwork
	var signals []GiaLit
	var session *giaWindowDiscoverySession
	var scratch *bool
	for i, row := range rows {
		next := row.Mode*2 + row.Mapping
		if key != next {
			m, n, signals = discoveryFixture(t, row.Mode, row.Mapping)
			session = &giaWindowDiscoverySession{}
			scratch = nil
			key = next
		}
		r, e := n.discoverWindow(m, signals[4+row.Pivot].ID(), row.Limit, row.Reverse != 0, GiaLUTAnalysisOptions{}, session)
		want := GiaDiscoveredWindow{Found: row.Count != 0, Roots: row.Roots, Nodes: row.Nodes, Leaves: row.Leaves, Ands: row.Ands}
		if e != nil || !sameGiaWindow(r, want) {
			t.Fatal(i, r, want, e)
		}
		if session.builds != 1 || session.refreshes != 1 {
			t.Fatal(session.builds, session.refreshes)
		}
		if scratch == nil {
			scratch = &session.no[0]
		} else if scratch != &session.no[0] {
			t.Fatal("scratch rebuilt")
		}
	}
}
func TestGiaDiscoverySessionRecovery(t *testing.T) {
	m, n, signals := discoveryFixture(t, 0, 0)
	session := &giaWindowDiscoverySession{}
	pivot := signals[11].ID()
	want, e := n.DiscoverWindow(m, pivot, 8, true, GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if _, e = n.discoverWindow(m, pivot, 8, true, GiaLUTAnalysisOptions{}, session); e != nil {
		t.Fatal(e)
	}
	base := slices.Clone(session.baseRefs)
	calls := 0
	_, e = n.discoverWindow(m, pivot, 8, true, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls == 70 }}, session)
	if !errors.Is(e, ErrCancelled) {
		t.Fatal(calls, e)
	}
	r, e := n.discoverWindow(m, pivot, 8, true, GiaLUTAnalysisOptions{}, session)
	if e != nil || !sameGiaWindow(r, want) || !slices.Equal(base, session.baseRefs) {
		t.Fatal(r, e)
	}
	var leaves, nodes []GiaID
	for _, l := range signals[:4] {
		leaves = append(leaves, l.ID())
	}
	for _, l := range signals[4:] {
		nodes = append(nodes, l.ID())
	}
	replacement, e := n.RemapUnitWindow(m, leaves, nodes, []GiaID{signals[7].ID(), signals[11].ID()}, 0, GiaUnitWindowSearchOptions{})
	if e != nil || replacement.Replacement.Network == nil {
		t.Fatal(replacement, e)
	}
	updated := replacement.Replacement.Network
	fresh, e := updated.DiscoverWindow(m, pivot, 8, false, GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	r, e = updated.discoverWindow(m, pivot, 8, false, GiaLUTAnalysisOptions{}, session)
	if e != nil || !sameGiaWindow(r, fresh) || session.refreshes != 2 || session.builds != 1 {
		t.Fatal(r, e, session.refreshes)
	}
	base = slices.Clone(session.baseRefs)
	bad, e := updated.WithWindow([]GiaID{pivot}, [][]GiaID{nil}, GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	other := signals[7].ID()
	if _, e = bad.discoverWindow(m, other, 8, false, GiaLUTAnalysisOptions{}, session); e == nil {
		t.Fatal("invalid refresh")
	}
	if session.network != updated || !slices.Equal(base, session.baseRefs) {
		t.Fatal("partial refresh")
	}
	if _, e = updated.discoverWindow(m, pivot, 8, false, GiaLUTAnalysisOptions{}, session); e != nil {
		t.Fatal(e)
	}
	if e = m.PatchCODriver(0, signals[0]); e != nil {
		t.Fatal(e)
	}
	if _, e = updated.discoverWindow(m, pivot, 8, false, GiaLUTAnalysisOptions{}, session); !errors.Is(e, ErrInvariant) {
		t.Fatal("changed graph accepted", e)
	}
}
func FuzzGiaDiscoverySession(f *testing.F) {
	f.Add(uint8(0), uint8(7), uint8(8), true)
	f.Fuzz(func(t *testing.T, mode, pivot, limit uint8, reverse bool) {
		m, n, signals := discoveryFixture(t, int(mode%3), int(mode/3%2))
		session := &giaWindowDiscoverySession{}
		for i := 0; i < 3; i++ {
			p := signals[4+(int(pivot)+i)%8].ID()
			bound := (int(limit) + i) % 12
			fresh, e := n.DiscoverWindow(m, p, bound, reverse, GiaLUTAnalysisOptions{})
			if e != nil {
				t.Fatal(e)
			}
			reused, e := n.discoverWindow(m, p, bound, reverse, GiaLUTAnalysisOptions{}, session)
			if e != nil || !sameGiaWindow(fresh, reused) {
				t.Fatal(fresh, reused, e)
			}
		}
		if session.builds != 1 || session.refreshes != 1 {
			t.Fatal(session)
		}
	})
}
