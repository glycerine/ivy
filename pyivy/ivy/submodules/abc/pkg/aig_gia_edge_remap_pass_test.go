package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgeRemapPassNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_edge_remap_pass.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Mode, Maximum, Limit, Baseline, PriorMode, Reverse, ImprovementLimit, Count, NativeDelay, FreshDelay int
		PriorEdges, NativeEdges                                                                              [][2]GiaID
		Fanins                                                                                               [][]GiaID
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 72 {
		t.Fatal(len(rows))
	}
	nativeLeaks := 0
	for i, row := range rows {
		m, n := unitPassFixture(t, row.Mode)
		prior := &GiaEdgeAssignment{rows: slices.Clone(row.PriorEdges)}
		before := prior.Rows()
		r, e := n.RemapEdgePass(m, prior, row.Maximum, GiaUnitRemapPassOptions{WindowLimit: 16, Reverse: row.Reverse != 0, MaxImprovements: row.ImprovementLimit})
		if e != nil || len(r.Network.Roots()) != row.Count || r.DelayLimit != row.Limit {
			t.Fatal(i, r, e)
		}
		if !slices.Equal(prior.Rows(), before) || len(n.Roots()) != 10 {
			t.Fatal("source changed")
		}
		if r.Improvements > 0 {
			if r.Timing.Maximum > r.DelayLimit {
				t.Fatal("late replacement", i, r)
			}
		} else {
			if !slices.Equal(r.Edges.Rows(), row.PriorEdges) || r.Timing.Maximum != row.Baseline {
				t.Fatal("trial rows leaked", i, r)
			}
			if !slices.Equal(row.NativeEdges, row.PriorEdges) {
				nativeLeaks++
			}
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
		native := &GiaLUTNetwork{fanins: row.Fanins}
		functions, e := native.ValidateFunctions(m, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		checkGiaLUTComposition(t, m, native, functions)
		if r.SolverBuilds != 1 || r.DiscoveryBuilds != 1 || r.DiscoveryQueries != len(r.Attempts) {
			t.Fatal("reuse", r)
		}
		if row.Count < 10 && row.FreshDelay > row.Limit {
			t.Fatal("native corrected delay", row)
		}
	}
	if nativeLeaks != 12 {
		t.Fatal("native regression count", nativeLeaks)
	}
}
func TestGiaEdgeRemapPassPackedAndLimits(t *testing.T) {
	m, n := unitPassFixture(t, 0)
	prior, _, e := n.SelectEdges(m, true, GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	mapping, e := n.PackedMapping(0, GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	before := mapping.Packed()
	rows := prior.Rows()
	r, e := mapping.RemapEdgePass(m, prior, 0, GiaUnitRemapPassOptions{WindowLimit: 16})
	if e != nil || len(r.Mapping.Roots()) != 2 || r.Pass.Timing.Maximum > r.Pass.DelayLimit {
		t.Fatal(r, e)
	}
	if !slices.Equal(before, mapping.Packed()) || !slices.Equal(rows, prior.Rows()) {
		t.Fatal("source changed")
	}
	r.Pass.Edges.rows[0][0] = 1
	if prior.rows[0][0] != 0 {
		t.Fatal("edge alias")
	}
	for _, o := range []GiaUnitRemapPassOptions{{MaxPackedEntries: -1}, {MaxPackedEntries: 1}, {MaxPivots: 1}, {Search: GiaUnitWindowSearchOptions{Stop: func() bool { return true }}}} {
		r, e := mapping.RemapEdgePass(m, prior, 0, o)
		if e == nil || r.Mapping != nil || r.Pass.Network != nil || r.Pass.Edges != nil {
			t.Fatal(r, e)
		}
	}
	if _, e = n.RemapEdgePass(m, nil, 0, GiaUnitRemapPassOptions{}); e == nil {
		t.Fatal("nil prior")
	}
}
func FuzzGiaEdgeRemapPass(f *testing.F) {
	f.Add(uint8(0), uint8(0), true, true)
	f.Fuzz(func(t *testing.T, mode, limit uint8, reverse, greedy bool) {
		m, n := unitPassFixture(t, int(mode%3))
		prior := &GiaEdgeAssignment{rows: make([][2]GiaID, m.ObjectCount())}
		if greedy {
			var e error
			prior, _, e = n.SelectEdges(m, true, GiaLUTAnalysisOptions{})
			if e != nil {
				t.Fatal(e)
			}
		}
		rows := prior.Rows()
		r, e := n.RemapEdgePass(m, prior, int(limit%41), GiaUnitRemapPassOptions{WindowLimit: 16, Reverse: reverse})
		if e != nil {
			t.Fatal(e)
		}
		if r.Improvements > 0 && r.Timing.Maximum > r.DelayLimit {
			t.Fatal(r)
		}
		if !slices.Equal(rows, prior.Rows()) {
			t.Fatal("prior changed")
		}
		if r.Improvements == 0 && !slices.Equal(rows, r.Edges.Rows()) {
			t.Fatal("unchanged assignment replaced")
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
	})
}
