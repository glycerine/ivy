package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

func TestGiaEdgeBoxesNativeNetwork(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_edge_boxes_network.json")
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
		network, err := mapping.Network(GiaLUTAnalysisOptions{})
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
		r, err := network.EvaluateDelayWithBoxes(m, e, timing, GiaLUTAnalysisOptions{})
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
