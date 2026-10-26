package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

func TestGiaLUTBoxUnitTimingReference(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/gia_lut_box_unit_timing.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Outputs1, Outputs2, Maximum, EdgeCount, Start  int
		Primary                                        [2]uint32
		Tables                                         [][]uint32
		CIArrivals, COArrivals, CIRequired, CORequired []uint32
		Delays, Packed, Required, NativeRequired       []int
		Edges                                          [][2]GiaID
		Objects                                        [][3]int
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 48 {
		t.Fatal(len(rows))
	}
	differences := 0
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
		r, err := network.UnitTimingWithBoxes(m, timing, row.Start, GiaLUTAnalysisOptions{})
		if err != nil || r.Timing.Maximum != row.Maximum || !slices.Equal(r.Timing.Arrivals, row.Delays) || !slices.Equal(r.Timing.Required, row.Required) {
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
		if !slices.Equal(row.Required, row.NativeRequired) {
			differences++
		}
		for i, want := range row.CIRequired {
			p, _ := r.Manager.CI(i)
			if math.Float32bits(p.Required) != want {
				t.Fatal(index, "CI required", i, p, want)
			}
		}
		for i, want := range row.CORequired {
			p, _ := r.Manager.CO(i)
			if math.Float32bits(p.Required) != want {
				t.Fatal(index, "CO required", i, p, want)
			}
		}
		if r.Manager.epoch != 2 {
			t.Fatal("two timing epochs")
		}
		if !slices.Equal(before, m.objects) || !slices.Equal(timeBefore, timSnapshot(timing)) || !slices.Equal(e.Rows(), row.Edges) {
			t.Fatal("source changed")
		}
	}
	if differences == 0 {
		t.Fatal("missing native counter-clobber regression")
	}
	t.Logf("%d required arrays differ from original boxed timing", differences)

}

func TestGiaLUTBoxUnitTimingBoundsAndOwnership(t *testing.T) {
	m, mapping, timing, _ := buildGiaEdgeBoxTest(t)
	n, _ := mapping.Network(GiaLUTAnalysisOptions{})
	before := timSnapshot(timing)
	r, err := n.UnitTimingWithBoxes(m, timing, 0, GiaLUTAnalysisOptions{MaxVisits: 27})
	if err != nil || r.Timing.Maximum != 4 || r.Timing.Required[4] != 1 || r.Timing.Required[1] != 0 || r.Manager.epoch != 2 {
		t.Fatal(r, err)
	}
	input, _ := r.Manager.CO(0)
	if input.Required != 1.5 {
		t.Fatal("float required before integer conversion", input)
	}
	for _, options := range []GiaLUTAnalysisOptions{{MaxVisits: 26}, {MaxObjects: 1}, {MaxVisits: -1}, {MaxObjects: -1}, {Stop: func() bool { return true }}} {
		if r, err := n.UnitTimingWithBoxes(m, timing, 0, options); r.Manager != nil || r.Timing.Required != nil || err == nil {
			t.Fatal(r, err)
		}
	}
	for _, threshold := range []int{1, 10, 30, 60} {
		calls := 0
		r, err := n.UnitTimingWithBoxes(m, timing, 0, GiaLUTAnalysisOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if calls >= threshold && (err == nil || r.Manager != nil) {
			t.Fatal("cancelled result escaped", threshold, r, err)
		}
	}
	if !slices.Equal(before, timSnapshot(timing)) {
		t.Fatal("source timing changed")
	}
	r.Manager.tables[0].Delays[0] = 99
	if timing.tables[0].Delays[0] != 1.5 {
		t.Fatal("table clone alias")
	}
	noBoxes, _ := NewTimManager(3, 2, 0)
	noBoxes.InitPIArrivalAll(99)
	noBoxes.SetDelayTable(0, TimDelayTable{Inputs: 1, Outputs: 1, Delays: []float32{1}}, TimDelayOptions{})
	fallback, err := n.UnitTimingWithBoxes(m, noBoxes, 0, GiaLUTAnalysisOptions{MaxVisits: 13})
	if err != nil || fallback.Timing.Maximum != 1 || fallback.Manager.epoch != 0 {
		t.Fatal(fallback, err)
	}
	if _, err := n.UnitTimingWithBoxes(m, noBoxes, 0, GiaLUTAnalysisOptions{MaxVisits: 12}); err == nil {
		t.Fatal("fallback aggregate budget")
	}
}
func FuzzGiaLUTBoxUnitTiming(f *testing.F) {
	f.Add(uint16(5), int16(-7), int16(3), uint8(0))
	f.Add(uint16(255), int16(20), int16(-4), uint8(255))
	f.Fuzz(func(t *testing.T, a uint16, b, delay int16, start uint8) {
		m, mapping, timing, _ := buildGiaEdgeBoxTest(t)
		n, _ := mapping.Network(GiaLUTAnalysisOptions{})
		arrivalA, arrivalB := float32(a%80)/4, float32(b)/4
		arc := float32(delay) / 8
		if start&128 != 0 {
			arc = TimNoArc
		}
		timing.InitPIArrival(0, arrivalA)
		timing.InitPIArrival(1, arrivalB)
		timing.SetDelayTable(0, TimDelayTable{Inputs: 1, Outputs: 1, Delays: []float32{arc}}, TimDelayOptions{})
		before := timSnapshot(timing)
		r, err := n.UnitTimingWithBoxes(m, timing, int(start%20), GiaLUTAnalysisOptions{})
		if err != nil {
			t.Fatal(err)
		}
		input := max(0, int(arrivalA)+1, int(arrivalB)+1)
		boxArrival := int(float32(float32(input) + arc))
		if arc == TimNoArc {
			boxArrival = -1000000000
		}
		output := max(0, boxArrival+1, int(arrivalA)+1)
		maximum := max(int(start%20), input, output)
		required := float32(float32(maximum-1) - arc)
		if arc == TimNoArc {
			required = TimEternity
		}
		inputRequired := int(required)
		if r.Timing.Maximum != maximum || r.Timing.Arrivals[4] != input || r.Timing.Arrivals[5] != output || r.Timing.Required[4] != inputRequired || r.Timing.Required[2] != min(1000000000, inputRequired-1) || r.Timing.Required[1] != min(maximum-1, inputRequired-1) {
			t.Fatal("independent boxed recurrence", r.Timing, maximum, inputRequired)
		}
		if !slices.Equal(before, timSnapshot(timing)) {
			t.Fatal("source timing changed")
		}
	})
}
