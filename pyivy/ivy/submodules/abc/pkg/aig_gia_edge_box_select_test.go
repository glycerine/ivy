package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgeBoxSelectNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_edge_box_select.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Outputs1, Outputs2, Maximum, EdgeCount, UseTwo int
		Primary                                        [2]uint32
		Tables                                         [][]uint32
		CIArrivals, COArrivals                         []uint32
		Delays, Packed                                 []int
		Edges                                          [][2]GiaID
		Objects                                        [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 32 {
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
		r, err := mapping.SelectEdgesWithBoxes(m, timing, row.UseTwo != 0, GiaLUTAnalysisOptions{})
		if err != nil || r.Timing.Maximum != row.Maximum || r.Timing.Edges != row.EdgeCount || !slices.Equal(r.Timing.Delays, row.Delays) || !slices.Equal(r.Edges.Rows(), row.Edges) {
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

func TestGiaEdgeBoxSelectBoundsAndFallback(t *testing.T) {
	m, mapping, timing, _ := buildGiaEdgeBoxTest(t)
	before := timSnapshot(timing)
	r, err := mapping.SelectEdgesWithBoxes(m, timing, true, GiaLUTAnalysisOptions{MaxVisits: 14})
	if err != nil || r.Timing.Maximum != 11 || r.Manager.epoch != 1 {
		t.Fatal(r, err)
	}
	reevaluated, err := r.Edges.EvaluateDelayWithBoxes(m, mapping, timing, GiaLUTAnalysisOptions{})
	if err != nil || !slices.Equal(reevaluated.Timing.Delays, r.Timing.Delays) {
		t.Fatal("selection reevaluation", err)
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: 13}, {MaxObjects: 1}, {MaxVisits: -1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		if r, err := mapping.SelectEdgesWithBoxes(m, timing, false, options); r.Manager != nil || r.Edges != nil || err == nil {
			t.Fatal(r, err)
		}
	}
	negative := timing.Clone()
	negative.InitPIArrivalAll(-100)
	if r, err := mapping.SelectEdgesWithBoxes(m, negative, true, GiaLUTAnalysisOptions{}); r.Edges != nil || err == nil {
		t.Fatal("native negative-candidate assertion", r, err)
	}
	noBoxes, _ := NewTimManager(3, 2, 0)
	noBoxes.InitPIArrivalAll(99)
	fallback, err := mapping.SelectEdgesWithBoxes(m, noBoxes, true, GiaLUTAnalysisOptions{})
	if err != nil || fallback.Timing.Maximum != 10 || fallback.Manager.epoch != 0 || len(fallback.Order) != 0 {
		t.Fatal(fallback, err)
	}
	if !slices.Equal(before, timSnapshot(timing)) {
		t.Fatal("source timing changed")
	}
}
func FuzzGiaEdgeBoxSelect(f *testing.F) {
	f.Add(uint16(5), int16(-7), int16(3), true)
	f.Add(uint16(255), int16(20), int16(-4), false)
	f.Fuzz(func(t *testing.T, a uint16, b, delay int16, two bool) {
		m, mapping, timing, blank := buildGiaEdgeBoxTest(t)
		timing.InitPIArrival(0, float32(a%80)/4)
		timing.InitPIArrival(1, float32(b)/4)
		timing.SetDelayTable(0, TimDelayTable{Inputs: 1, Outputs: 1, Delays: []float32{float32(delay) / 8}}, TimDelayOptions{})
		before := timSnapshot(timing)
		r, err := mapping.SelectEdgesWithBoxes(m, timing, two, GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		actual, err := r.Edges.EvaluateDelayWithBoxes(m, mapping, timing, GiaLUTAnalysisOptions{})
		if err != nil || !slices.Equal(actual.Timing.Delays, r.Timing.Delays) || actual.Timing.Maximum != r.Timing.Maximum {
			t.Fatal("selected assignment timing", err)
		}
		baseline, err := blank.EvaluateDelayWithBoxes(m, mapping, timing, GiaLUTAnalysisOptions{})
		if err != nil || r.Timing.Maximum > baseline.Timing.Maximum {
			t.Fatal("delay regression", err)
		}
		for id := range m.objects {
			count, err := r.Edges.Count(GiaID(id))
			if err != nil || count > 2 {
				t.Fatal("capacity", id, count, err)
			}
		}
		if !slices.Equal(before, timSnapshot(timing)) {
			t.Fatal("source manager changed")
		}
	})
}
