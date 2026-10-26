package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgeWindowFeedbackNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_edge_window_feedback.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Limit, Accepted, Maximum int
		Nodes                    []GiaID
		Clause                   []SatLit
		Delays                   []int
		Edges                    [][2]GiaID
		Original, Window         [][]GiaID
		Objects                  [][3]int
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 96 {
		t.Fatal(len(rows))
	}
	accepted, rejected := 0, 0
	for i, row := range rows {
		m := NewGiaManager()
		for _, obj := range row.Objects[1:] {
			switch obj[0] {
			case 1:
				_, e = m.AppendCI()
			case 2:
				_, e = m.AppendCO(GiaLit(obj[1]))
			case 3:
				_, e = m.AppendAnd(GiaLit(obj[1]), GiaLit(obj[2]))
			}
			if e != nil {
				t.Fatal(e)
			}
		}
		var records []GiaLUTRecord
		for id, f := range row.Original {
			if len(f) > 0 {
				records = append(records, GiaLUTRecord{Root: GiaID(id), Fanins: f})
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
		before := slices.Clone(m.objects)
		r, e := n.EvaluateEdgeWindow(m, row.Nodes, row.Window, row.Limit, GiaLUTAnalysisOptions{})
		if e != nil || r.Accepted != (row.Accepted != 0) || r.Timing.Maximum != row.Maximum || !slices.Equal(r.Timing.Delays, row.Delays) || !slices.Equal(r.Edges.Rows(), row.Edges) || !slices.Equal(r.Clause, row.Clause) {
			t.Fatal(i, r, e)
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("graph changed")
		}
		if r.Accepted {
			accepted++
		} else {
			rejected++
			s := NewSatSolver()
			s.SetVariableCount(len(row.Nodes))
			s.AddClause(r.Clause)
			var a []SatLit
			for j, f := range row.Window {
				if len(f) > 0 {
					l, _ := MakeSatLit(SatVar(j), false)
					a = append(a, l)
				}
			}
			answer, e := s.Solve(a, SatLimits{})
			if e != nil || answer.Status != SatUnsatisfiable {
				t.Fatal("feedback", answer, e)
			}
		}
	}
	if accepted == 0 || rejected == 0 {
		t.Fatal(accepted, rejected)
	}
}
func TestGiaEdgeWindowSearchIntegration(t *testing.T) {
	m, n, l, ids, _ := mappingStateFixture(t, 6, 0, 4)
	m.AppendCO(GiaLit(ids[len(ids)-1]) << 1)
	n.fanins = append(n.fanins, nil)
	n.fanouts = append(n.fanouts, nil)
	r, e := n.RemapEdgeWindow(m, l, ids, []GiaID{ids[len(ids)-1]}, 10, GiaUnitWindowSearchOptions{MaxIterations: 100})
	if e != nil || r.Replacement.Network == nil || len(r.Search.BestCuts) != 1 || r.Replacement.Timing.Maximum != 10 {
		t.Fatal(r, e)
	}
	checkGiaLUTComposition(t, m, r.Replacement.Network, r.Replacement.Functions)
	rejected := false
	for _, step := range r.Search.Steps {
		if len(step.Conflict) > 0 {
			rejected = true
		}
	}
	if !rejected {
		t.Fatal("timing feedback not exercised")
	}
	low, e := n.SearchEdgeWindow(m, l, ids, []GiaID{ids[len(ids)-1]}, 0, GiaUnitWindowSearchOptions{MaxIterations: 100})
	if e != nil || low.Improved || len(low.BestCuts) > 0 || len(low.Steps) == 0 || low.Steps[len(low.Steps)-1].Status != SatUnsatisfiable {
		t.Fatal(low, e)
	}
	for _, o := range []GiaLUTAnalysisOptions{{MaxObjects: 1}, {MaxVisits: 1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		x, e := n.EvaluateEdgeWindow(m, ids, r.Search.BestWindow, 10, o)
		if e == nil || x.Edges != nil || x.Clause != nil {
			t.Fatal(x, e)
		}
	}
	if x, e := n.WithValidatedEdgeWindow(m, ids, r.Search.BestWindow, 0, GiaLUTAnalysisOptions{}); e == nil || x.Network != nil {
		t.Fatal(x, e)
	}
}
func FuzzGiaEdgeWindowSearch(f *testing.F) {
	f.Add(uint8(5), uint8(0), uint8(10))
	f.Fuzz(func(t *testing.T, size, mode, bound uint8) {
		m, n, l, ids, _ := mappingStateFixture(t, 2+int(size%7), int(mode%2), 4)
		m.AppendCO(GiaLit(ids[len(ids)-1]) << 1)
		n.fanins = append(n.fanins, nil)
		n.fanouts = append(n.fanouts, nil)
		limit := int(bound % 31)
		r, e := n.RemapEdgeWindow(m, l, ids, []GiaID{ids[len(ids)-1]}, limit, GiaUnitWindowSearchOptions{MaxIterations: 100})
		if e != nil {
			t.Fatal(e)
		}
		if r.Replacement.Network != nil {
			if r.Replacement.Timing.Maximum > limit {
				t.Fatal(r)
			}
			checkGiaLUTComposition(t, m, r.Replacement.Network, r.Replacement.Functions)
		}
		if limit < 10 && len(r.Search.BestCuts) > 0 {
			t.Fatal("sub-first-level limit", r)
		}
	})
}
