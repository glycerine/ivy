package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

// Replay native timing decisions through the actual feedback loop, fixing
// each candidate cut so a model decision heuristic cannot alter the reference.
func TestGiaUnitWindowSearchNativeFeedback(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_lut_window_timing.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Start, Accepted  int
		Nodes, Roots     []GiaID
		Clause           []SatLit
		Original, Window [][]GiaID
		Objects          [][3]int
	}
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	for index, row := range rows {
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
		leaves := []GiaID{1, 2, 3, 4}
		cuts := GiaWindowCuts{}
		for i, f := range row.Window {
			c := GiaWindowCut{Node: i}
			for _, id := range f {
				if j := slices.Index(leaves, id); j >= 0 {
					c.Leaves[j/64] |= 1 << uint(j%64)
				} else {
					j := slices.Index(row.Nodes, id)
					if j < 0 {
						t.Fatal(id)
					}
					c.Nodes[j/64] |= 1 << uint(j%64)
				}
			}
			cuts.Ranges = append(cuts.Ranges, GiaWindowCutRange{Start: len(cuts.Cuts), Count: 2})
			cuts.Cuts = append(cuts.Cuts, c, GiaWindowCut{Node: i, Nodes: [2]uint64{1 << uint(i), 0}})
		}
		var roots []int
		for _, id := range row.Roots {
			roots = append(roots, slices.Index(row.Nodes, id))
		}
		solver, e := NewGiaWindowSolver(3, len(row.Nodes), cuts.Cuts, cuts.Ranges, roots, GiaWindowCNFOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for _, span := range cuts.Ranges {
			l, _ := MakeSatLit(SatVar(solver.first+span.Start), false)
			if _, e = solver.solver.AddClause([]SatLit{l}); e != nil {
				t.Fatal(e)
			}
		}
		r, e := runGiaUnitWindowSearch(n, m, leaves, row.Nodes, row.Roots, cuts, solver, len(row.Nodes), row.Start, GiaUnitWindowSearchOptions{MaxIterations: 2})
		if e != nil || len(r.Steps) != 2 {
			t.Fatal(index, r, e)
		}
		first := r.Steps[0]
		if first.Status != SatSatisfiable || first.Accepted != (row.Accepted != 0) || !slices.Equal(first.Conflict, row.Clause) || r.Steps[1].Status != SatUnsatisfiable {
			t.Fatal(index, r, row.Clause)
		}
	}
}
func TestGiaUnitWindowSearchIntegration(t *testing.T) {
	m, n, leaves, nodes, _ := mappingStateFixture(t, 6, 0, 4)
	before := slices.Clone(m.objects)
	original := make([][]GiaID, len(n.fanins))
	for i, f := range n.fanins {
		original[i] = slices.Clone(f)
	}
	r, e := n.SearchUnitWindow(m, leaves, nodes, []GiaID{nodes[len(nodes)-1]}, 0, GiaUnitWindowSearchOptions{})
	if e != nil || !r.Improved || r.BestNodes != 1 || len(r.BestCuts) != 1 {
		t.Fatal(r, e)
	}
	if !slices.Equal(before, m.objects) {
		t.Fatal("graph changed")
	}
	for i, f := range n.fanins {
		if !slices.Equal(f, original[i]) {
			t.Fatal("mapping changed")
		}
	}
	r, e = n.SearchUnitWindow(m, leaves, nodes, []GiaID{nodes[len(nodes)-1]}, 0, GiaUnitWindowSearchOptions{MaxIterations: 1})
	if e != nil || !r.IterationLimit || len(r.Steps) != 1 {
		t.Fatal(r, e)
	}
	for _, o := range []GiaUnitWindowSearchOptions{{MaxIterations: -1}, {MaxIterations: 10001}, {Stop: func() bool { return true }}, {Cuts: GiaWindowCutOptions{MaxCuts: 1}}, {CNF: GiaWindowCNFOptions{MaxClauses: 1}}, {Analysis: GiaLUTAnalysisOptions{MaxVisits: 1}}} {
		r, e = n.SearchUnitWindow(m, leaves, nodes, []GiaID{nodes[len(nodes)-1]}, 0, o)
		if e == nil || r.Steps != nil || r.BestCuts != nil {
			t.Fatal(r, e)
		}
	}
}
func FuzzGiaUnitWindowSearch(f *testing.F) {
	f.Add(uint8(4), uint8(0))
	f.Fuzz(func(t *testing.T, size, mode uint8) {
		m, n, l, ids, _ := mappingStateFixture(t, 1+int(size%8), int(mode%2), 4)
		r, e := n.SearchUnitWindow(m, l, ids, []GiaID{ids[len(ids)-1]}, 0, GiaUnitWindowSearchOptions{})
		if e != nil || r.BestNodes != 1 || len(r.BestCuts) != 1 {
			t.Fatal(r, e)
		}
		timing, e := n.EvaluateUnitWindow(m, ids, r.BestWindow, []GiaID{ids[len(ids)-1]}, 0, GiaLUTAnalysisOptions{})
		if e != nil || !timing.Accepted {
			t.Fatal(timing, e)
		}
	})
}
