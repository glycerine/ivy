package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

func TestGiaAreaUnitNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_area_unit_pass.json")
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
	regressions := 0
	for i, row := range rows {
		m, n := wideUnitFixture(t, row.Variables, row.Mode)
		before, _ := n.UnitTiming(m, 0, GiaLUTAnalysisOptions{})
		r, e := n.RemapUnitPass(m, 0, GiaUnitRemapPassOptions{WindowLimit: 16, Reverse: row.Reverse != 0, MaxImprovements: row.ImprovementLimit, Search: GiaUnitWindowSearchOptions{LutSize: row.Width, AreaOnly: true}})
		if e != nil || len(r.Network.Roots()) != row.Count {
			t.Fatal(i, r, e, row.Count)
		}
		if r.Timing.Maximum > before.Maximum {
			regressions++
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
		native := &GiaLUTNetwork{fanins: row.Fanins}
		functions, e := native.ValidateFunctions(m, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for j, f := range functions {
			if uint64(f.Root) != row.Truths[j][0] || f.Truth != row.Truths[j][1] {
				t.Fatal(i, f)
			}
		}
	}
	if regressions == 0 {
		t.Fatal("area-only did not permit timing tradeoff")
	}
}
func TestGiaAreaEdgeNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_area_edge_pass.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Mode, Maximum, Limit, Reverse, ImprovementLimit, Count int
		PriorEdges, NativeEdges                                [][2]GiaID
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 72 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		m, n := unitPassFixture(t, row.Mode)
		prior := &GiaEdgeAssignment{rows: slices.Clone(row.PriorEdges)}
		r, e := n.RemapEdgePass(m, prior, row.Maximum, GiaUnitRemapPassOptions{WindowLimit: 16, Reverse: row.Reverse != 0, MaxImprovements: row.ImprovementLimit, Search: GiaUnitWindowSearchOptions{AreaOnly: true}})
		if e != nil || len(r.Network.Roots()) != row.Count || r.TimingConstrained || r.DelayLimit != row.Limit || !slices.Equal(r.Edges.Rows(), row.NativeEdges) || !slices.Equal(prior.Rows(), row.PriorEdges) {
			t.Fatal(i, r, e)
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
	}
}
func TestGiaAreaBoxNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_area_box_pass.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []boxRemapPassFixture
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 36 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		m, n, manager := buildBoxRemapPassFixture(t, row)
		before := timSnapshot(manager)
		o := GiaUnitRemapPassOptions{WindowLimit: 16, Reverse: row.Reverse != 0, MaxImprovements: row.ImprovementLimit, Search: GiaUnitWindowSearchOptions{AreaOnly: true}}
		var network *GiaLUTNetwork
		var functions []GiaLUTFunction
		var updated *TimManager
		var maximum int
		if row.Unit != 0 {
			r, e := n.RemapUnitPassWithBoxes(m, manager, 100, o)
			if e != nil {
				t.Fatal(i, e)
			}
			network, functions, updated, maximum = r.Pass.Network, r.Pass.Functions, r.Manager, r.Pass.Timing.Maximum
		} else {
			prior := &GiaEdgeAssignment{rows: slices.Clone(row.PriorEdges)}
			r, e := n.RemapEdgePassWithBoxes(m, manager, prior, 0, o)
			if e != nil || r.Pass.TimingConstrained || r.Pass.DelayLimit != 0 || !slices.Equal(r.Pass.Edges.Rows(), row.PriorEdges) {
				t.Fatal(i, r, e)
			}
			network, functions, updated, maximum = r.Pass.Network, r.Pass.Functions, r.Manager, r.Pass.Timing.Maximum
		}
		if len(network.Roots()) != row.Count || maximum != row.Maximum {
			t.Fatal(i, len(network.Roots()), maximum, row)
		}
		checkGiaLUTComposition(t, m, network, functions)
		for j, want := range row.CIArrivals {
			p, _ := updated.CI(j)
			if math.Float32bits(p.Arrival) != want {
				t.Fatal(i, "CI", j, p, want)
			}
		}
		for j, want := range row.COArrivals {
			p, _ := updated.CO(j)
			if math.Float32bits(p.Arrival) != want {
				t.Fatal(i, "CO", j, p, want)
			}
		}
		if !slices.Equal(before, timSnapshot(manager)) {
			t.Fatal("timing changed")
		}
	}
}
func FuzzGiaAreaUnitPass(f *testing.F) {
	f.Add(uint8(0), uint8(4), true)
	f.Fuzz(func(t *testing.T, kind, width uint8, area bool) {
		m, n := wideUnitFixture(t, 5+int(kind%2), int(kind/2%2))
		before, _ := n.UnitTiming(m, 0, GiaLUTAnalysisOptions{})
		r, e := n.RemapUnitPass(m, 0, GiaUnitRemapPassOptions{WindowLimit: 16, Search: GiaUnitWindowSearchOptions{AreaOnly: area, LutSize: 1 + int(width%6)}})
		if e != nil || !area && r.Timing.Maximum > before.Maximum {
			t.Fatal(r, e)
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
	})
}
func FuzzGiaAreaEdgePass(f *testing.F) {
	f.Add(uint8(0), uint8(1), true)
	f.Fuzz(func(t *testing.T, kind, limit uint8, area bool) {
		m, n := unitPassFixture(t, int(kind%3))
		prior, _, e := n.SelectEdges(m, true, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		before := prior.Rows()
		r, e := n.RemapEdgePass(m, prior, int(limit%41), GiaUnitRemapPassOptions{WindowLimit: 16, Search: GiaUnitWindowSearchOptions{AreaOnly: area}})
		if e != nil || r.TimingConstrained == area {
			t.Fatal(r, e)
		}
		if area && !slices.Equal(r.Edges.Rows(), before) {
			t.Fatal("prior rows changed")
		}
		if !area && r.Improvements > 0 && r.Timing.Maximum > r.DelayLimit {
			t.Fatal(r)
		}
		checkGiaLUTComposition(t, m, r.Network, r.Functions)
	})
}
