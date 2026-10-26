package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

type boxEdgeWindowFixture struct {
	Outputs1, Outputs2, Maximum, EdgeCount, Limit, Accepted int
	Primary                                                 [2]uint32
	Tables                                                  [][]uint32
	CIArrivals, COArrivals                                  []uint32
	Delays, Packed                                          []int
	Edges                                                   [][2]GiaID
	Objects                                                 [][3]int
	Nodes                                                   []GiaID
	Window                                                  [][]GiaID
	Clause                                                  []SatLit
}

func buildBoxEdgeWindowFixture(t *testing.T, row boxEdgeWindowFixture) (*GiaManager, *GiaLUTNetwork, *TimManager) {
	t.Helper()
	m := NewGiaManager()
	var e error
	for _, obj := range row.Objects[1:] {
		switch obj[0] {
		case 1:
			_, e = m.AppendCI()
		case 2:
			_, e = m.AppendCO(GiaLit(obj[1]))
		case 3:
			_, e = m.AppendAnd(GiaLit(obj[1]), GiaLit(obj[2]))
		case 5:
			_, e = m.AppendBuffer(GiaLit(obj[1]))
		}
		if e != nil {
			t.Fatal(e)
		}
	}
	mapping, e := GiaLUTMappingFromPacked(m.ObjectCount(), row.Packed, 0)
	if e != nil {
		t.Fatal(e)
	}
	n, e := mapping.Network(GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	timing, e := NewTimManager(len(m.cis), len(m.cos), 0)
	if e != nil {
		t.Fatal(e)
	}
	timing.AddBox(TimBox{Inputs: 2, FirstOutput: 2, Outputs: row.Outputs1}, TimBoxOptions{})
	timing.AddBox(TimBox{FirstInput: 2, Inputs: 2, FirstOutput: 2 + row.Outputs1, Outputs: row.Outputs2, DelayTable: 1}, TimBoxOptions{})
	for k, values := range row.Tables {
		outs := row.Outputs1
		if k == 1 {
			outs = row.Outputs2
		}
		table := TimDelayTable{Inputs: 2, Outputs: outs, Delays: make([]float32, len(values))}
		for i, v := range values {
			table.Delays[i] = math.Float32frombits(v)
		}
		if e = timing.SetDelayTable(k, table, TimDelayOptions{}); e != nil {
			t.Fatal(e)
		}
	}
	for i, v := range row.Primary {
		timing.InitPIArrival(i, math.Float32frombits(v))
	}
	return m, n, timing
}
func TestGiaBoxEdgeWindowNative(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_box_edge_window.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []boxEdgeWindowFixture
	if e = json.Unmarshal(b, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 96 {
		t.Fatal(len(rows))
	}
	for i, row := range rows {
		m, n, timing := buildBoxEdgeWindowFixture(t, row)
		before := timSnapshot(timing)
		r, e := n.EvaluateEdgeWindowWithBoxes(m, timing, row.Nodes, row.Window, row.Limit, GiaLUTAnalysisOptions{})
		if e != nil {
			t.Fatal(i, e)
		}
		v := r.Evaluation
		if v.Accepted != (row.Accepted != 0) || v.Timing.Maximum != row.Maximum || v.Timing.Edges != row.EdgeCount || !slices.Equal(v.Timing.Delays, row.Delays) || !slices.Equal(v.Edges.Rows(), row.Edges) || !slices.Equal(v.Clause, row.Clause) {
			t.Fatal(i, r, e)
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
		if !slices.Equal(before, timSnapshot(timing)) {
			t.Fatal("timing changed")
		}
		if v.Accepted {
			applied, e := n.WithValidatedEdgeWindowWithBoxes(m, timing, row.Nodes, row.Window, row.Limit, GiaLUTAnalysisOptions{})
			if e != nil {
				t.Fatal(i, e)
			}
			checkGiaLUTComposition(t, m, applied.Network, applied.Functions)
		}
	}
}
func boxWindowRoots(m *GiaManager) []GiaID {
	var roots []GiaID
	for _, id := range m.cos {
		driver, _ := m.objects[id].Fanin0(id)
		if m.objects[driver.ID()].IsAnd() && !slices.Contains(roots, driver.ID()) {
			roots = append(roots, driver.ID())
		}
	}
	return roots
}
func TestGiaBoxEdgeWindowSearch(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_box_edge_window.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []boxEdgeWindowFixture
	json.Unmarshal(b, &rows)
	m, n, timing := buildBoxEdgeWindowFixture(t, rows[0])
	before := timSnapshot(timing)
	nodes := n.Roots()
	r, e := n.RemapEdgeWindowWithBoxes(m, timing, m.cis, nodes, boxWindowRoots(m), 100, GiaUnitWindowSearchOptions{MaxIterations: 100})
	if e != nil || r.Replacement.Network == nil || !r.Search.Improved || r.Replacement.Timing.Maximum > 100 {
		t.Fatal(r, e)
	}
	checkGiaLUTComposition(t, m, r.Replacement.Network, r.Replacement.Functions)
	if !slices.Equal(before, timSnapshot(timing)) {
		t.Fatal("source timing changed")
	}
	for _, o := range []GiaLUTAnalysisOptions{{MaxObjects: 1}, {MaxVisits: 1}, {Stop: func() bool { return true }}} {
		v, e := n.EvaluateEdgeWindowWithBoxes(m, timing, nodes, r.Search.BestWindow, 100, o)
		if e == nil || v.Manager != nil || v.Evaluation.Edges != nil {
			t.Fatal(v, e)
		}
	}
	if _, e = n.EvaluateEdgeWindowWithBoxes(m, nil, nodes, r.Search.BestWindow, 100, GiaLUTAnalysisOptions{}); e == nil {
		t.Fatal("nil timing")
	}
}
func FuzzGiaBoxEdgeWindow(f *testing.F) {
	b, e := os.ReadFile("test_vectors/oracle/gia_box_edge_window.json")
	if e != nil {
		f.Fatal(e)
	}
	var rows []boxEdgeWindowFixture
	if e = json.Unmarshal(b, &rows); e != nil {
		f.Fatal(e)
	}
	f.Add(uint8(0), uint8(100))
	f.Fuzz(func(t *testing.T, index, limit uint8) {
		row := rows[int(index)%len(rows)]
		m, n, timing := buildBoxEdgeWindowFixture(t, row)
		before := timSnapshot(timing)
		maximum := int(limit % 101)
		r, e := n.RemapEdgeWindowWithBoxes(m, timing, m.cis, n.Roots(), boxWindowRoots(m), maximum, GiaUnitWindowSearchOptions{MaxIterations: 100})
		if e != nil {
			t.Fatal(e)
		}
		if r.Replacement.Network != nil {
			if r.Replacement.Timing.Maximum > maximum {
				t.Fatal(r)
			}
			checkGiaLUTComposition(t, m, r.Replacement.Network, r.Replacement.Functions)
		}
		if !slices.Equal(before, timSnapshot(timing)) {
			t.Fatal("timing changed")
		}
	})
}

func TestGiaBoxEdgeWindowNoBoxFallback(t *testing.T) {
	m, n := unitPassFixture(t, 0)
	timing, e := NewTimManager(len(m.cis), len(m.cos), 0)
	if e != nil {
		t.Fatal(e)
	}
	timing.InitPIArrival(0, 500)
	nodes := n.Roots()
	window := make([][]GiaID, len(nodes))
	for i, id := range nodes {
		window[i] = slices.Clone(n.fanins[id])
	}
	plain, e := n.EvaluateEdgeWindow(m, nodes, window, 100, GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	boxed, e := n.EvaluateEdgeWindowWithBoxes(m, timing, nodes, window, 100, GiaLUTAnalysisOptions{})
	if e != nil || boxed.Manager.epoch != timing.epoch || boxed.Evaluation.Accepted != plain.Accepted || boxed.Evaluation.Timing.Maximum != plain.Timing.Maximum || !slices.Equal(boxed.Evaluation.Timing.Delays, plain.Timing.Delays) || !slices.Equal(boxed.Evaluation.Edges.Rows(), plain.Edges.Rows()) {
		t.Fatal(boxed, plain, e)
	}
}
