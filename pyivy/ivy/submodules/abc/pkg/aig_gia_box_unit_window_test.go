package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

type boxUnitWindowFixture struct {
	boxEdgeWindowFixture
	Required []int
	Roots    []GiaID
}

func TestGiaBoxUnitWindowCorrectedReference(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_box_unit_window.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []boxUnitWindowFixture
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 96 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		m, n, manager := buildBoxEdgeWindowFixture(t, row.boxEdgeWindowFixture)
		before := timSnapshot(manager)
		r, e := n.EvaluateUnitWindowWithBoxes(m, manager, row.Nodes, row.Window, row.Roots, row.Limit, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(i, e)
		}
		v := r.Evaluation
		if v.Accepted != (row.Accepted != 0) || v.ReferenceMaximum != row.Maximum || !slices.Equal(v.Arrivals, row.Delays) || !slices.Equal(v.Required, row.Required) || !slices.Equal(v.Clause, row.Clause) {
			t.Fatal(i, r, row.Clause, e)
		}
		for j, want := range row.CIArrivals {
			p, _ := r.Manager.CI(j)
			if math.Float32bits(p.Arrival) != want {
				t.Fatal(i, "CI", j, p, want)
			}
		}
		for j, want := range row.COArrivals {
			p, _ := r.Manager.CO(j)
			if math.Float32bits(p.Arrival) != want {
				t.Fatal(i, "CO", j, p, want)
			}
		}
		if !slices.Equal(before, timSnapshot(manager)) {
			t.Fatal("timing changed")
		}
		if v.Accepted {
			applied, e := n.WithValidatedUnitWindowWithBoxes(m, manager, row.Nodes, row.Window, row.Limit, GiaLUTAnalysisOptions{})
			if e != nil {
				t.Fatal(i, e)
			}
			checkGiaLUTComposition(t, m, applied.Network, applied.Functions)
		}
	}
}
func TestGiaBoxUnitWindowSearch(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_box_unit_window.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []boxUnitWindowFixture
	json.Unmarshal(b, &rows)
	m, n, manager := buildBoxEdgeWindowFixture(t, rows[0].boxEdgeWindowFixture)
	before := timSnapshot(manager)
	nodes := n.Roots()
	r, e := n.RemapUnitWindowWithBoxes(m, manager, m.cis, nodes, boxWindowRoots(m), 30, GiaUnitWindowSearchOptions{MaxIterations: 100})
	if e != nil || r.Replacement.Network == nil || !r.Search.Improved {
		t.Fatal(r, e)
	}
	checkGiaLUTComposition(t, m, r.Replacement.Network, r.Replacement.Functions)
	if !slices.Equal(before, timSnapshot(manager)) {
		t.Fatal("timing changed")
	}
	for _, o := range []GiaLUTAnalysisOptions{{MaxObjects: 1}, {MaxVisits: 1}, {MaxVisits: -1}, {Stop: func() bool { return true }}} {
		v, e := n.EvaluateUnitWindowWithBoxes(m, manager, nodes, r.Search.BestWindow, boxWindowRoots(m), 0, o)
		if e == nil || v.Manager != nil || v.Evaluation.Arrivals != nil {
			t.Fatal(v, e)
		}
	}
}
func TestGiaBoxUnitWindowGlobalGuard(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_box_unit_window.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []boxUnitWindowFixture
	json.Unmarshal(b, &rows)
	for _, row := range rows {
		if row.Accepted != 0 {
			continue
		}
		m, n, manager := buildBoxEdgeWindowFixture(t, row.boxEdgeWindowFixture)
		original, e := n.UnitTimingWithBoxes(m, manager, row.Limit, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		proposal, e := n.WithWindow(row.Nodes, row.Window, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		timing, e := proposal.UnitTimingWithBoxes(m, manager, row.Limit, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if timing.Timing.Maximum <= original.Timing.Maximum {
			continue
		}
		result, e := n.WithValidatedUnitWindowWithBoxes(m, manager, row.Nodes, row.Window, row.Limit, GiaLUTAnalysisOptions{})
		if e == nil || result.Network != nil || result.Timing.Manager != nil {
			t.Fatal("global regression accepted", result, e)
		}
		return
	}
	t.Fatal("missing global regression case")
}
func FuzzGiaBoxUnitWindow(f *testing.F) {
	b, e := os.ReadFile("test_vectors/oracle/gia_box_unit_window.json")
	if e != nil {
		f.Fatal(e)
	}
	var rows []boxUnitWindowFixture
	if e = json.Unmarshal(b, &rows); e != nil {
		f.Fatal(e)
	}
	f.Add(uint8(0), uint8(30))
	f.Fuzz(func(t *testing.T, index, start uint8) {
		row := rows[int(index)%len(rows)]
		m, n, manager := buildBoxEdgeWindowFixture(t, row.boxEdgeWindowFixture)
		before := timSnapshot(manager)
		delayStart := int(start % 31)
		original, e := n.UnitTimingWithBoxes(m, manager, delayStart, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(e)
		}
		r, e := n.RemapUnitWindowWithBoxes(m, manager, m.cis, n.Roots(), boxWindowRoots(m), delayStart, GiaUnitWindowSearchOptions{MaxIterations: 100})
		if e != nil {
			t.Fatal(e)
		}
		if r.Replacement.Network != nil {
			if r.Replacement.Timing.Timing.Maximum > original.Timing.Maximum {
				t.Fatal(r)
			}
			checkGiaLUTComposition(t, m, r.Replacement.Network, r.Replacement.Functions)
		}
		if !slices.Equal(before, timSnapshot(manager)) {
			t.Fatal("timing changed")
		}
	})
}

func TestGiaBoxUnitWindowNoBoxFallback(t *testing.T) {
	m, n := unitPassFixture(t, 0)
	manager, e := NewTimManager(len(m.cis), len(m.cos), 0)
	if e != nil {
		t.Fatal(e)
	}
	manager.InitPIArrival(0, 500)
	nodes := n.Roots()
	window := make([][]GiaID, len(nodes))
	for i, id := range nodes {
		window[i] = slices.Clone(n.fanins[id])
	}
	plain, e := n.EvaluateUnitWindow(m, nodes, window, boxWindowRoots(m), 0, GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	boxed, e := n.EvaluateUnitWindowWithBoxes(m, manager, nodes, window, boxWindowRoots(m), 0, GiaLUTAnalysisOptions{})
	if e != nil || boxed.Manager.epoch != manager.epoch || boxed.Evaluation.Accepted != plain.Accepted || boxed.Evaluation.ReferenceMaximum != plain.ReferenceMaximum || !slices.Equal(boxed.Evaluation.Arrivals, plain.Arrivals) || !slices.Equal(boxed.Evaluation.Required, plain.Required) || !slices.Equal(boxed.Evaluation.Clause, plain.Clause) {
		t.Fatal(boxed, plain, e)
	}
}
