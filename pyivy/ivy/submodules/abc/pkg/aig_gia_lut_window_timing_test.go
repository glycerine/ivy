package abc

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
)

func TestGiaUnitWindowNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_lut_window_timing.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Start, ReferenceMax, Accepted int
		Nodes, Roots                  []GiaID
		Clause                        []SatLit
		Arrivals, Required            []int
		Original, Window              [][]GiaID
		Objects                       [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 32 {
		t.Fatal(len(rows))
	}
	accepted, rejected := 0, 0
	for index, row := range rows {
		m := NewGiaManager()
		for _, obj := range row.Objects[1:] {
			switch obj[0] {
			case 1:
				_, err = m.AppendCI()
			case 2:
				_, err = m.AppendCO(GiaLit(obj[1]))
			case 3:
				_, err = m.AppendAnd(GiaLit(obj[1]), GiaLit(obj[2]))
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		records := make([]GiaLUTRecord, 0)
		for id, fanins := range row.Original {
			if len(fanins) > 0 {
				records = append(records, GiaLUTRecord{Root: GiaID(id), Fanins: fanins})
			}
		}
		mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
		if err != nil {
			t.Fatal(err)
		}
		n, _ := mapping.Network(GiaLUTAnalysisOptions{})
		before := slices.Clone(m.objects)
		r, err := n.EvaluateUnitWindow(m, row.Nodes, row.Window, row.Roots, row.Start, GiaLUTAnalysisOptions{})
		if err != nil || r.Accepted != (row.Accepted != 0) || r.ReferenceMaximum != row.ReferenceMax || !slices.Equal(r.Arrivals, row.Arrivals) || !slices.Equal(r.Required, row.Required) || !slices.Equal(r.Clause, row.Clause) {
			t.Fatal(index, r, row.Clause, err)
		}
		if r.Accepted {
			accepted++
		} else {
			rejected++
		}
		if !slices.Equal(before, m.objects) {
			t.Fatal("source graph changed")
		}
	}
	if accepted == 0 || rejected == 0 {
		t.Fatal("missing timing decisions")
	}
}
func buildGiaUnitWindowTest(t *testing.T) (*GiaManager, *GiaLUTNetwork, []GiaID, [][]GiaID) {
	t.Helper()
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	d, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(c, d)
	z, _ := m.AppendAnd(x, y)
	v, _ := m.AppendAnd(z, a)
	root, _ := m.AppendAnd(v, y)
	m.AppendCO(root)
	m.AppendCO(z)
	nodes := []GiaID{x.ID(), y.ID(), z.ID(), v.ID(), root.ID()}
	window := [][]GiaID{{1, 2}, {3, 4}, {5, 6}, {1, 7}, {6, 8}}
	records := make([]GiaLUTRecord, 0)
	for _, id := range nodes {
		records = append(records, GiaLUTRecord{Root: id, Fanins: []GiaID{1, 2, 3, 4}})
	}
	mapping, err := NewGiaLUTMapping(m.ObjectCount(), records, 0)
	if err != nil {
		t.Fatal(err)
	}
	n, _ := mapping.Network(GiaLUTAnalysisOptions{})
	return m, n, nodes, window
}
func TestGiaUnitWindowClauseAndLimits(t *testing.T) {
	m, n, nodes, window := buildGiaUnitWindowTest(t)
	r, err := n.EvaluateUnitWindow(m, nodes, window, []GiaID{9, 7}, 0, GiaLUTAnalysisOptions{MaxVisits: 60})
	if err != nil || r.Accepted || !slices.Equal(r.Path, []GiaID{9, 8, 7, 5}) || !slices.Equal(r.Clause, []SatLit{9, 7, 5, 1}) || r.Boundary != 1 {
		t.Fatal(r, err)
	}
	solver := NewSatSolver()
	if ok, err := solver.AddClause(r.Clause); !ok || err != nil {
		t.Fatal(ok, err)
	}
	assumptions := make([]SatLit, 5)
	for i := range assumptions {
		assumptions[i], _ = MakeSatLit(SatVar(i), false)
	}
	answer, err := solver.Solve(assumptions, SatLimits{})
	if err != nil || answer.Status != SatUnsatisfiable {
		t.Fatal(answer, err)
	}
	assumptions[r.Clause[0].Var()] = r.Clause[0]
	answer, err = solver.Solve(assumptions, SatLimits{})
	if err != nil || answer.Status != SatSatisfiable {
		t.Fatal(answer, err)
	}
	reordered := []GiaID{6, 5, 7, 8, 9}
	permuted := [][]GiaID{window[1], window[0], window[2], window[3], window[4]}
	r, err = n.EvaluateUnitWindow(m, reordered, permuted, []GiaID{9}, 0, GiaLUTAnalysisOptions{})
	if err != nil || !slices.Equal(r.Clause, []SatLit{9, 7, 5, 3}) {
		t.Fatal("selection-variable indexing", r, err)
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: 59}, {MaxObjects: 1}, {MaxVisits: -1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		if r, err := n.EvaluateUnitWindow(m, nodes, window, []GiaID{9}, 0, options); r.Arrivals != nil || r.Clause != nil || err == nil {
			t.Fatal(r, err)
		}
	}
	if _, err := n.EvaluateUnitWindow(m, []GiaID{7, 5}, [][]GiaID{{5}, {1}}, []GiaID{7}, 0, GiaLUTAnalysisOptions{}); err == nil {
		t.Fatal("non-topological proposal")
	}
	calls := 0
	if r, err := n.EvaluateUnitWindow(m, nodes, window, []GiaID{9}, 0, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= 20 }}); r.Arrivals != nil || err == nil {
		t.Fatal(r, err)
	}
}
func FuzzGiaUnitWindow(f *testing.F) {
	f.Add(uint8(0))
	f.Add(uint8(255))
	f.Fuzz(func(t *testing.T, flags uint8) {
		m, n, nodes, window := buildGiaUnitWindowTest(t)
		for i := range window {
			if flags&(1<<uint(i)) != 0 {
				window[i] = []GiaID{1, 2, 3, 4}
			}
		}
		if flags&32 != 0 {
			nodes[0], nodes[1] = nodes[1], nodes[0]
			window[0], window[1] = window[1], window[0]
		}
		roots := []GiaID{9, 7}
		if flags&64 != 0 {
			roots[0], roots[1] = roots[1], roots[0]
		}
		r, err := n.EvaluateUnitWindow(m, nodes, window, roots, int(flags>>6), GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		original, err := n.UnitTiming(m, int(flags>>6), GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		arrivals := slices.Clone(original.Arrivals)
		seen := make([]bool, m.ObjectCount())
		var value func(GiaID) int
		value = func(id GiaID) int {
			index := slices.Index(nodes, id)
			if index < 0 {
				return original.Arrivals[id]
			}
			if seen[id] {
				return arrivals[id]
			}
			seen[id] = true
			arrivals[id] = 0
			for _, fanin := range window[index] {
				arrivals[id] = max(arrivals[id], value(fanin)+1)
			}
			return arrivals[id]
		}
		for _, id := range nodes {
			value(id)
		}
		accepted := true
		bad := InvalidGiaID
		for _, id := range roots {
			if arrivals[id] > original.Required[id] {
				accepted = false
				bad = id
				break
			}
		}
		if r.Accepted != accepted || r.ViolatingRoot != bad || !slices.Equal(r.Arrivals, arrivals) || !slices.Equal(r.Required, original.Required) {
			t.Fatal("independent proposal timing", r)
		}
		if !accepted {
			if len(r.Path) != len(r.Clause) || len(r.Path) == 0 {
				t.Fatal("empty conflict")
			}
			for i, id := range r.Path {
				index := slices.Index(nodes, id)
				if !r.Clause[i].Negative() || int(r.Clause[i].Var()) != index {
					t.Fatal("selection clause")
				}
				next := r.Boundary
				if i+1 < len(r.Path) {
					next = r.Path[i+1]
				}
				first := InvalidGiaID
				for _, fanin := range window[index] {
					if arrivals[fanin]+1 == arrivals[id] {
						first = fanin
						break
					}
				}
				if first != next {
					t.Fatal("critical-fanin order")
				}
			}
		}
	})
}
