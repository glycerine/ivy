package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgeBoxesNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_edge_boxes.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Outputs1, Outputs2, Maximum, EdgeCount int
		Primary                                [2]uint32
		Tables                                 [][]uint32
		CIArrivals, COArrivals                 []uint32
		Delays, Packed                         []int
		Edges                                  [][2]GiaID
		Objects                                [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 48 {
		t.Fatal(len(rows))
	}
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
			case 5:
				_, err = m.AppendBuffer(GiaLit(obj[1]))
			}
			if err != nil {
				t.Fatal(err)
			}
		}
		mapping, err := GiaLUTMappingFromPacked(m.ObjectCount(), row.Packed, 0)
		if err != nil {
			t.Fatal(err)
		}
		timing, _ := NewTimManager(len(m.cis), len(m.cos), 0)
		timing.AddBox(TimBox{Inputs: 2, FirstOutput: 2, Outputs: row.Outputs1}, TimBoxOptions{})
		timing.AddBox(TimBox{FirstInput: 2, Inputs: 2, FirstOutput: 2 + row.Outputs1, Outputs: row.Outputs2, DelayTable: 1}, TimBoxOptions{})
		for tableIndex, values := range row.Tables {
			table := TimDelayTable{Inputs: 2, Outputs: row.Outputs1, Delays: make([]float32, len(values))}
			if tableIndex == 1 {
				table.Outputs = row.Outputs2
			}
			for i, bits := range values {
				table.Delays[i] = math.Float32frombits(bits)
			}
			timing.SetDelayTable(tableIndex, table, TimDelayOptions{})
		}
		for i, bits := range row.Primary {
			timing.InitPIArrival(i, math.Float32frombits(bits))
		}
		e := &GiaEdgeAssignment{rows: slices.Clone(row.Edges)}
		before := slices.Clone(m.objects)
		timeBefore := timSnapshot(timing)
		r, err := e.EvaluateDelayWithBoxes(m, mapping, timing, GiaLUTAnalysisOptions{})
		if err != nil || r.Timing.Maximum != row.Maximum || r.Timing.Edges != row.EdgeCount || !slices.Equal(r.Timing.Delays, row.Delays) {
			t.Fatal(index, r, err)
		}
		for i, want := range row.CIArrivals {
			p, _ := r.Manager.CI(i)
			if math.Float32bits(p.Arrival) != want {
				t.Fatal(index, "CI", i, p, want)
			}
		}
		for i, want := range row.COArrivals {
			p, _ := r.Manager.CO(i)
			if math.Float32bits(p.Arrival) != want {
				t.Fatal(index, "CO", i, p, want)
			}
		}
		if !slices.Equal(before, m.objects) || !slices.Equal(timeBefore, timSnapshot(timing)) || !slices.Equal(e.Rows(), row.Edges) {
			t.Fatal("source changed")
		}
	}
}
func buildGiaEdgeBoxTest(t *testing.T) (*GiaManager, *GiaLUTMapping, *TimManager, *GiaEdgeAssignment) {
	t.Helper()
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	box, _ := m.AppendCI()
	input, _ := m.AppendAnd(a, b)
	root, _ := m.AppendAnd(box, a)
	dangling, _ := m.AppendAnd(a, b)
	m.AppendCO(input)
	m.AppendCO(root)
	mapping, err := NewGiaLUTMapping(m.ObjectCount(), []GiaLUTRecord{{Root: input.ID(), Fanins: []GiaID{a.ID(), b.ID()}}, {Root: root.ID(), Fanins: []GiaID{box.ID(), a.ID()}}, {Root: dangling.ID(), Fanins: []GiaID{a.ID(), b.ID()}}}, 0)
	if err != nil {
		t.Fatal(err)
	}
	timing, _ := NewTimManager(3, 2, 0)
	timing.AddBox(TimBox{Inputs: 1, FirstOutput: 2, Outputs: 1}, TimBoxOptions{})
	timing.SetDelayTable(0, TimDelayTable{Inputs: 1, Outputs: 1, Delays: []float32{1.5}}, TimDelayOptions{})
	timing.InitPIArrival(0, 1.25)
	timing.InitPIArrival(1, -4.5)
	edges, _ := GiaEdgesFromPairs(m.ObjectCount(), nil, GiaEdgeOptions{})
	return m, mapping, timing, edges
}
func TestGiaEdgeBoxesBoundsAndOwnership(t *testing.T) {
	m, mapping, timing, edges := buildGiaEdgeBoxTest(t)
	before := timSnapshot(timing)
	r, err := edges.EvaluateDelayWithBoxes(m, mapping, timing, GiaLUTAnalysisOptions{MaxVisits: 14})
	if err != nil || r.Timing.Maximum != 22 || r.Manager.epoch != 1 || r.Timing.Delays[m.cos[0]] != 0 {
		t.Fatal(r, err)
	}
	r.Manager.tables[0].Delays[0] = 99
	if timing.tables[0].Delays[0] != 1.5 {
		t.Fatal("returned table alias")
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: 13}, {MaxObjects: 1}, {MaxVisits: -1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		if r, err := edges.EvaluateDelayWithBoxes(m, mapping, timing, options); r.Manager != nil || r.Timing.Delays != nil || err == nil {
			t.Fatal(r, err)
		}
	}
	for _, threshold := range []int{1, 5, 15, 30, 50} {
		calls := 0
		r, err := edges.EvaluateDelayWithBoxes(m, mapping, timing, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if calls >= threshold && (err == nil || r.Manager != nil) {
			t.Fatal("cancelled result escaped", threshold, r, err)
		}
	}
	if !slices.Equal(before, timSnapshot(timing)) {
		t.Fatal("source timing changed")
	}
	bad := timing.Clone()
	bad.InitPIArrival(0, float32(math.Inf(1)))
	if r, err := edges.EvaluateDelayWithBoxes(m, mapping, bad, GiaLUTAnalysisOptions{}); r.Manager != nil || err == nil {
		t.Fatal("undefined CI conversion", r, err)
	}
}
func TestGiaEdgeBoxesNoBoxFallback(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	root, _ := m.AppendAnd(a, b)
	m.AppendCO(root)
	mapping, _ := NewGiaLUTMapping(m.ObjectCount(), []GiaLUTRecord{{Root: root.ID(), Fanins: []GiaID{a.ID(), b.ID()}}}, 0)
	timing, _ := NewTimManager(2, 1, 0)
	timing.InitPIArrivalAll(99)
	timing.SetDelayTable(0, TimDelayTable{Inputs: 1, Outputs: 1, Delays: []float32{1}}, TimDelayOptions{})
	e, _ := GiaEdgesFromPairs(m.ObjectCount(), nil, GiaEdgeOptions{})
	r, err := e.EvaluateDelayWithBoxes(m, mapping, timing, GiaLUTAnalysisOptions{MaxVisits: 3})
	if err != nil || r.Timing.Maximum != 10 || r.Manager.epoch != 0 {
		t.Fatal(r, err)
	}
	if _, err := e.EvaluateDelayWithBoxes(m, mapping, timing, GiaLUTAnalysisOptions{MaxVisits: 2}); err == nil {
		t.Fatal("combined fallback budget")
	}
}
func FuzzGiaEdgeBoxes(f *testing.F) {
	f.Add(uint8(0), int16(5), int16(-7), int16(3))
	f.Add(uint8(255), int16(-100), int16(20), int16(-4))
	f.Fuzz(func(t *testing.T, flags uint8, a, b, delay int16) {
		m, mapping, timing, edges := buildGiaEdgeBoxTest(t)
		timing.InitPIArrival(0, float32(a)/4)
		timing.InitPIArrival(1, float32(b)/4)
		timing.SetDelayTable(0, TimDelayTable{Inputs: 1, Outputs: 1, Delays: []float32{float32(delay) / 8}}, TimDelayOptions{})
		if flags&1 != 0 {
			edges, _ = GiaEdgesFromPairs(m.ObjectCount(), [][2]GiaID{{1, 4}, {3, 5}}, GiaEdgeOptions{})
		}
		r, err := edges.EvaluateDelayWithBoxes(m, mapping, timing, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		cost := 10
		if flags&1 != 0 {
			cost = 2
		}
		input := max(0, int(float32(a)/4)+cost, int(float32(b)/4)+10)
		boxArrival := int(float32(float32(input) + float32(delay)/8))
		want := max(0, boxArrival+cost, int(float32(a)/4)+10)
		want = max(want, input)
		if r.Timing.Maximum != want {
			t.Fatal(r.Timing.Maximum, want)
		}
	})
}
