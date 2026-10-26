package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

type boxRemapPassFixture struct {
	Unit, Mode, Reverse, ImprovementLimit, PriorMode, Baseline, Count, Maximum int
	Primary                                                                    [2]uint32
	Tables                                                                     [][]uint32
	PriorEdges                                                                 [][2]GiaID
	CIArrivals, COArrivals                                                     []uint32
	Fanins                                                                     [][]GiaID
}

func buildBoxRemapPassFixture(t *testing.T, row boxRemapPassFixture) (*GiaManager, *GiaLUTNetwork, *TimManager) {
	t.Helper()
	m := NewGiaManager()
	var ci []GiaLit
	for i := 0; i < 6; i++ {
		l, _ := m.AppendCI()
		ci = append(ci, l)
	}
	var gate [3][10]GiaLit
	gate[0][0], _ = m.AppendAnd(ci[0], ci[1])
	for i := 1; i < 10; i++ {
		gate[0][i], _ = m.AppendAnd(gate[0][i-1], ci[i%2])
	}
	for c := 1; c < 3; c++ {
		input := []GiaLit{ci[2*c], ci[2*c+1], ci[0], ci[1]}
		gate[c][0], _ = m.AppendAnd(input[0], input[1])
		gate[c][1], _ = m.AppendAnd(input[2], input[3])
		gate[c][2], _ = m.AppendAnd(gate[c][0], gate[c][1])
		for i := 3; i < 10; i++ {
			gate[c][i], _ = m.AppendAnd(gate[c][i-1], input[i%4])
		}
	}
	m.AppendCO(gate[0][9])
	m.AppendCO(ci[0])
	m.AppendCO(gate[1][9])
	m.AppendCO(ci[2])
	m.AppendCO(gate[2][9])
	var records []GiaLUTRecord
	for _, part := range gate {
		for _, l := range part {
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
	manager, _ := NewTimManager(6, 5, 0)
	manager.AddBox(TimBox{Inputs: 2, FirstOutput: 2, Outputs: 2}, TimBoxOptions{})
	manager.AddBox(TimBox{FirstInput: 2, Inputs: 2, FirstOutput: 4, Outputs: 2, DelayTable: 1}, TimBoxOptions{})
	for k, values := range row.Tables {
		table := TimDelayTable{Inputs: 2, Outputs: 2, Delays: make([]float32, len(values))}
		for i, v := range values {
			table.Delays[i] = math.Float32frombits(v)
		}
		if e = manager.SetDelayTable(k, table, TimDelayOptions{}); e != nil {
			t.Fatal(e)
		}
	}
	for i, v := range row.Primary {
		manager.InitPIArrival(i, math.Float32frombits(v))
	}
	return m, n, manager
}
func TestGiaBoxRemapPassReference(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_box_remap_pass.json")
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
		options := GiaUnitRemapPassOptions{WindowLimit: 16, Reverse: row.Reverse != 0, MaxImprovements: row.ImprovementLimit}
		var result *GiaLUTNetwork
		var functions []GiaLUTFunction
		var updated *TimManager
		var maximum, builds, resets, improvements int
		if row.Unit != 0 {
			r, e := n.RemapUnitPassWithBoxes(m, manager, 100, options)
			if e != nil {
				t.Fatal(i, e)
			}
			result, functions, updated, maximum = r.Pass.Network, r.Pass.Functions, r.Manager, r.Pass.Timing.Maximum
			builds, resets, improvements = r.Pass.SolverBuilds, r.Pass.WindowResets, r.Pass.Improvements
		} else {
			prior := &GiaEdgeAssignment{rows: slices.Clone(row.PriorEdges)}
			r, e := n.RemapEdgePassWithBoxes(m, manager, prior, 0, options)
			if e != nil || r.Pass.DelayLimit != row.Baseline {
				t.Fatal(i, r, e)
			}
			result, functions, updated, maximum = r.Pass.Network, r.Pass.Functions, r.Manager, r.Pass.Timing.Maximum
			builds, resets, improvements = r.Pass.SolverBuilds, r.Pass.WindowResets, r.Pass.Improvements
			if maximum > r.Pass.DelayLimit {
				t.Fatal("edge limit", i, r)
			}
			if !slices.Equal(prior.Rows(), row.PriorEdges) {
				t.Fatal("prior changed")
			}
		}
		if len(result.Roots()) != row.Count || maximum != row.Maximum || builds != 1 || resets < improvements-1 {
			t.Fatal(i, len(result.Roots()), maximum, row, builds, resets)
		}
		for id, f := range row.Fanins {
			if !slices.Equal(result.fanins[id], f) {
				t.Fatal(i, "mapping", id, result.fanins[id], f)
			}
		}
		checkGiaLUTComposition(t, m, result, functions)
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
		if !slices.Equal(before, timSnapshot(manager)) || len(n.Roots()) != 30 {
			t.Fatal("source changed")
		}
	}
}
func TestGiaBoxRemapPassPackedAndFailure(t *testing.T) {
	b, e := os.ReadFile("test_vectors/oracle/gia_box_remap_pass.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []boxRemapPassFixture
	json.Unmarshal(b, &rows)
	m, n, manager := buildBoxRemapPassFixture(t, rows[0])
	mapping, e := n.PackedMapping(0, GiaLUTAnalysisOptions{})
	if e != nil {
		t.Fatal(e)
	}
	prior := &GiaEdgeAssignment{rows: slices.Clone(rows[0].PriorEdges)}
	before := timSnapshot(manager)
	options := GiaUnitRemapPassOptions{WindowLimit: 16}
	unit, e := mapping.RemapUnitPassWithBoxes(m, manager, 100, options)
	if e != nil || len(unit.Mapping.Roots()) != 3 {
		t.Fatal(unit, e)
	}
	edge, e := mapping.RemapEdgePassWithBoxes(m, manager, prior, 0, options)
	if e != nil || len(edge.Mapping.Roots()) != 3 {
		t.Fatal(edge, e)
	}
	for _, o := range []GiaUnitRemapPassOptions{{MaxPackedEntries: 1}, {MaxPivots: 1}, {Search: GiaUnitWindowSearchOptions{Stop: func() bool { return true }}}} {
		u, e := mapping.RemapUnitPassWithBoxes(m, manager, 100, o)
		if e == nil || u.Mapping != nil || u.Pass.Manager != nil {
			t.Fatal(u, e)
		}
		v, e := mapping.RemapEdgePassWithBoxes(m, manager, prior, 0, o)
		if e == nil || v.Mapping != nil || v.Pass.Manager != nil {
			t.Fatal(v, e)
		}
	}
	if !slices.Equal(before, timSnapshot(manager)) {
		t.Fatal("timing changed")
	}
}
func FuzzGiaBoxRemapPass(f *testing.F) {
	b, e := os.ReadFile("test_vectors/oracle/gia_box_remap_pass.json")
	if e != nil {
		f.Fatal(e)
	}
	var rows []boxRemapPassFixture
	if e = json.Unmarshal(b, &rows); e != nil {
		f.Fatal(e)
	}
	f.Add(uint8(0), uint8(0))
	f.Fuzz(func(t *testing.T, index, cap uint8) {
		row := rows[int(index)%len(rows)]
		m, n, manager := buildBoxRemapPassFixture(t, row)
		before := timSnapshot(manager)
		o := GiaUnitRemapPassOptions{WindowLimit: 16, Reverse: row.Reverse != 0, MaxImprovements: int(cap % 4)}
		if row.Unit != 0 {
			r, e := n.RemapUnitPassWithBoxes(m, manager, 100, o)
			if e != nil || r.Pass.Timing.Maximum > 100 {
				t.Fatal(r, e)
			}
			checkGiaLUTComposition(t, m, r.Pass.Network, r.Pass.Functions)
		} else {
			prior := &GiaEdgeAssignment{rows: slices.Clone(row.PriorEdges)}
			r, e := n.RemapEdgePassWithBoxes(m, manager, prior, 0, o)
			if e != nil || r.Pass.Timing.Maximum > r.Pass.DelayLimit {
				t.Fatal(r, e)
			}
			checkGiaLUTComposition(t, m, r.Pass.Network, r.Pass.Functions)
		}
		if !slices.Equal(before, timSnapshot(manager)) {
			t.Fatal("timing changed")
		}
	})
}

func TestGiaBoxRemapPassNoBoxFallback(t *testing.T) {
	m, n := unitPassFixture(t, 0)
	manager, e := NewTimManager(len(m.cis), len(m.cos), 0)
	if e != nil {
		t.Fatal(e)
	}
	manager.InitPIArrival(0, 1000)
	o := GiaUnitRemapPassOptions{WindowLimit: 16}
	plain, e := n.RemapUnitPass(m, 0, o)
	if e != nil {
		t.Fatal(e)
	}
	boxed, e := n.RemapUnitPassWithBoxes(m, manager, 0, o)
	if e != nil || boxed.Manager.epoch != manager.epoch || boxed.Pass.Timing.Maximum != plain.Timing.Maximum || !slices.Equal(boxed.Pass.Timing.Arrivals, plain.Timing.Arrivals) || len(boxed.Pass.Network.Roots()) != len(plain.Network.Roots()) {
		t.Fatal(boxed, plain, e)
	}
	prior := &GiaEdgeAssignment{rows: make([][2]GiaID, m.ObjectCount())}
	edge, e := n.RemapEdgePass(m, prior, 0, o)
	if e != nil {
		t.Fatal(e)
	}
	boxedge, e := n.RemapEdgePassWithBoxes(m, manager, prior, 0, o)
	if e != nil || boxedge.Manager.epoch != manager.epoch || boxedge.Pass.DelayLimit != edge.DelayLimit || boxedge.Pass.Timing.Maximum != edge.Timing.Maximum || !slices.Equal(boxedge.Pass.Edges.Rows(), edge.Edges.Rows()) {
		t.Fatal(boxedge, edge, e)
	}
}
