package abc

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"slices"
	"strconv"
	"testing"
)

type timManagerVector struct {
	Epoch  int
	Use    int
	CI, CO [][6]int64
	Boxes  [][9]int
	Tables [][]uint32
}

type timDuplicateVector struct {
	Unit           int              `json:"unit"`
	Arrivals       []uint32         `json:"arrivals"`
	RequiredNative []uint32         `json:"required_native"`
	Before         timManagerVector `json:"before"`
	SourceAfter    timManagerVector `json:"source_after"`
	Duplicate      timManagerVector `json:"duplicate"`
}

type timRebuildVector struct {
	Name               string `json:"name"`
	Counts             [4]int `json:"counts"`
	Boxes              [][7]int
	Arrivals, Required []uint32
	Values             []int
}

func timManagerWords(m *TimManager) timManagerVector {
	result := timManagerVector{Epoch: m.epoch}
	if m.useTraversal {
		result.Use = 1
	}
	for _, terminal := range m.cis {
		words := timTerminalWords(terminal)
		result.CI = append(result.CI, words)
	}
	for _, terminal := range m.cos {
		words := timTerminalWords(terminal)
		result.CO = append(result.CO, words)
	}
	for _, box := range m.boxes {
		words := timBoxWords(box)
		result.Boxes = append(result.Boxes, words)
	}
	for _, table := range m.tables {
		if table.Inputs == 0 {
			result.Tables = append(result.Tables, nil)
			continue
		}
		words := []uint32{uint32(table.Inputs), uint32(table.Outputs)}
		for _, delay := range table.Delays {
			words = append(words, math.Float32bits(delay))
		}
		result.Tables = append(result.Tables, words)
	}
	return result
}

func makeTimDuplicateManager(t testing.TB) *TimManager {
	t.Helper()
	ciArr := []uint32{0x3fc00000, 0x80000000, 0x40a00000, 0x40c00000}
	ciReq := []uint32{0x41000000, 0x41100000, 0x41200000, 0x41300000}
	coArr := []uint32{0x41400000, 0x41500000, 0x41600000, 0x41700000}
	coReq := []uint32{0x41800000, 0x41900000, 0x42200000, 0x7fc12345}
	delays := []uint32{0x3fa00000, 0xce6e6b28, 0x7fc54321, 0xc0200000}
	m, err := NewTimManager(4, 4, 8)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = m.AddBox(TimBox{FirstInput: 0, Inputs: 2, FirstOutput: 2, Outputs: 2, DelayTable: 0, Black: true}, TimBoxOptions{}); err != nil {
		t.Fatal(err)
	}
	if err = m.SetBoxCopy(0, -5); err != nil {
		t.Fatal(err)
	}
	values := make([]float32, len(delays))
	for i, bits := range delays {
		values[i] = math.Float32frombits(bits)
	}
	if err = m.SetDelayTable(0, TimDelayTable{Inputs: 2, Outputs: 2, Delays: values}, TimDelayOptions{}); err != nil {
		t.Fatal(err)
	}
	m.epoch = 7
	m.useTraversal = false
	m.boxes[0].TravID = 55
	for i := 0; i < 4; i++ {
		m.cis[i].TravID = 10 + i
		m.cos[i].TravID = 20 + i
		m.cis[i].Arrival = math.Float32frombits(ciArr[i])
		m.cis[i].Required = math.Float32frombits(ciReq[i])
		m.cos[i].Arrival = math.Float32frombits(coArr[i])
		m.cos[i].Required = math.Float32frombits(coReq[i])
	}
	return m
}

func TestTimDuplicateNativeAndCorrections(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/tim_duplicate.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []timDuplicateVector
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[0].Unit != 0 || rows[1].Unit != 3 {
		t.Fatal(rows)
	}
	for index, row := range rows {
		source := makeTimDuplicateManager(t)
		before := timManagerWords(source)
		if !reflect.DeepEqual(before, row.Before) {
			t.Fatal(index, "native setup", before, row.Before)
		}
		duplicate, err := source.Duplicate(row.Unit)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(timManagerWords(source), before) {
			t.Fatal(index, "Go duplicate mutated source")
		}
		want := row.Duplicate
		want.CI = slices.Clone(row.Duplicate.CI)
		want.CO = slices.Clone(row.Duplicate.CO)
		if row.Unit != 0 {
			want.CI[0][4] = 0
			want.CI[1][4] = 0
			want.CO[2][5] = int64(math.Float32bits(TimEternity))
			want.CO[3][5] = int64(math.Float32bits(TimEternity))
		}
		if actual := timManagerWords(duplicate); !reflect.DeepEqual(actual, want) {
			t.Fatal(index, actual, want)
		}
		if reflect.DeepEqual(row.Before, row.SourceAfter) {
			t.Fatal("fixture does not expose native source mutation")
		}
		if row.Unit != 0 && reflect.DeepEqual(row.Duplicate, want) {
			t.Fatal("fixture does not expose native unit timing bug")
		}
		duplicate.cis[0].Arrival = 99
		duplicate.tables[0].Delays[0] = 88
		duplicate.boxes[0].Copy = 77
		if source.cis[0].Arrival == 99 || source.tables[0].Delays[0] == 88 || source.boxes[0].Copy == 77 {
			t.Fatal("duplicate aliases source")
		}
	}

	m := makeTimDuplicateManager(t)
	arrivals, err := m.PrimaryArrivalTimes()
	if err != nil || len(arrivals) != 2 || math.Float32bits(arrivals[0]) != rows[0].Arrivals[0] || math.Float32bits(arrivals[1]) != rows[0].Arrivals[1] {
		t.Fatal(arrivals, err)
	}
	required, err := m.PrimaryRequiredTimes()
	if err != nil || len(required) != 2 || math.Float32bits(required[0]) != 0x42200000 || math.Float32bits(required[1]) != 0x7fc12345 {
		t.Fatal(required, err)
	}
	if len(rows[0].RequiredNative) != 0 {
		t.Fatal("native required-time early-null bug not reproduced")
	}
	arrivals[0], required[0] = 0, 0
	if math.Float32bits(m.cis[0].Arrival) != 0x3fc00000 || math.Float32bits(m.cos[2].Required) != 0x42200000 {
		t.Fatal("timing slices alias manager")
	}
}

func TestTimDuplicateBoundsAndTrivialTimes(t *testing.T) {
	m, _ := NewTimManager(2, 2, 4)
	if arrivals, err := m.PrimaryArrivalTimes(); err != nil || arrivals != nil {
		t.Fatal(arrivals, err)
	}
	if required, err := m.PrimaryRequiredTimes(); err != nil || required != nil {
		t.Fatal(required, err)
	}
	m.cis[0].Arrival = float32(math.Copysign(0, -1))
	if arrivals, _ := m.PrimaryArrivalTimes(); arrivals != nil {
		t.Fatal("negative zero is native-trivial", arrivals)
	}
	m.cis[1].Arrival = float32(math.NaN())
	if arrivals, _ := m.PrimaryArrivalTimes(); len(arrivals) != 2 || !math.IsNaN(float64(arrivals[1])) {
		t.Fatal(arrivals)
	}
	var missing *TimManager
	if copy, err := missing.Duplicate(0); err == nil || copy != nil {
		t.Fatal(copy, err)
	}
	if values, err := missing.PrimaryArrivalTimes(); err == nil || values != nil {
		t.Fatal(values, err)
	}
	if values, err := missing.PrimaryRequiredTimes(); err == nil || values != nil {
		t.Fatal(values, err)
	}
	if strconv.IntSize == 64 {
		if copy, err := m.Duplicate(int(int64(math.MaxInt32) + 1)); err == nil || copy != nil {
			t.Fatal(copy, err)
		}
	}
}

func makeTimRebuildManager(t testing.TB) *TimManager {
	t.Helper()
	inputs := []int{2, 1, 2, 1}
	outputs := []int{1, 2, 1, 2}
	copies := []int{-1, 2, -1, 0}
	m, err := NewTimManager(8, 8, 16)
	if err != nil {
		t.Fatal(err)
	}
	firstCI, firstCO := 2, 0
	for box := range inputs {
		_, err = m.AddBox(TimBox{FirstInput: firstCO, Inputs: inputs[box], FirstOutput: firstCI, Outputs: outputs[box], DelayTable: box, Black: box&1 != 0}, TimBoxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		if err = m.SetBoxCopy(box, copies[box]); err != nil {
			t.Fatal(err)
		}
		delays := make([]float32, inputs[box]*outputs[box])
		for k := range delays {
			delays[k] = float32(10*box + k + 1)
		}
		if err = m.SetDelayTable(box, TimDelayTable{Inputs: inputs[box], Outputs: outputs[box], Delays: delays}, TimDelayOptions{}); err != nil {
			t.Fatal(err)
		}
		firstCI += outputs[box]
		firstCO += inputs[box]
	}
	m.cis[0].Arrival = math.Float32frombits(0x3fc00000)
	m.cis[1].Arrival = math.Float32frombits(0xc0100000)
	m.cos[6].Required = math.Float32frombits(0x42c80000)
	m.cos[7].Required = math.Float32frombits(0x7fc01234)
	return m
}

func timRebuildSummary(name string, m *TimManager) timRebuildVector {
	pis, pos := m.PrimaryCounts()
	result := timRebuildVector{Name: name, Counts: [4]int{len(m.cis), len(m.cos), pis, pos}, Boxes: make([][7]int, 0, len(m.boxes))}
	for _, box := range m.boxes {
		black := 0
		if box.Black {
			black = 1
		}
		result.Boxes = append(result.Boxes, [7]int{box.FirstInput, box.Inputs, box.FirstOutput, box.Outputs, box.DelayTable, box.Copy, black})
	}
	for _, terminal := range m.cis {
		if terminal.Box < 0 {
			result.Arrivals = append(result.Arrivals, math.Float32bits(terminal.Arrival))
		}
	}
	for _, terminal := range m.cos {
		if terminal.Box < 0 {
			result.Required = append(result.Required, math.Float32bits(terminal.Required))
		}
	}
	return result
}

func TestTimTrimReduceAlignNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/tim_rebuild.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Managers []timRebuildVector
	}
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Managers) != 6 {
		t.Fatal(len(fixture.Managers))
	}
	source := makeTimRebuildManager(t)
	before := timManagerWords(source)
	trimAll, err := source.TrimBoxes([]bool{true, true, true, true})
	if err != nil {
		t.Fatal(err)
	}
	trimEven, err := source.TrimBoxes([]bool{true, false, true, false})
	if err != nil {
		t.Fatal(err)
	}
	trimNone, err := source.TrimBoxes([]bool{false, false, false, false})
	if err != nil {
		t.Fatal(err)
	}
	reordered, err := source.ReduceBoxes([]int{3, 1}, 1)
	if err != nil {
		t.Fatal(err)
	}
	alignment, err := AlignTimBoxes(source, reordered)
	if err != nil {
		t.Fatal(err)
	}
	duplicated, err := source.ReduceBoxes([]int{0, 0}, 0)
	if err != nil {
		t.Fatal(err)
	}
	actual := []timRebuildVector{
		timRebuildSummary("trim_all", trimAll),
		timRebuildSummary("trim_even", trimEven),
		timRebuildSummary("trim_none", trimNone),
		timRebuildSummary("reduce_reorder", reordered),
		{Name: "alignment", Values: make([]int, len(alignment))},
		timRebuildSummary("reduce_duplicate", duplicated),
	}
	for i, value := range alignment {
		if value {
			actual[4].Values[i] = 1
		}
	}
	if !reflect.DeepEqual(actual, fixture.Managers) {
		t.Fatal(actual, fixture.Managers)
	}
	if !reflect.DeepEqual(before, timManagerWords(source)) {
		t.Fatal("rebuild mutated source")
	}
	trimEven.tables[0].Delays[0] = 99
	trimEven.cis[0].Arrival = 88
	if source.tables[0].Delays[0] == 99 || source.cis[0].Arrival == 88 {
		t.Fatal("rebuild aliases source")
	}
}

func TestTimTrimReduceAlignErrors(t *testing.T) {
	m := makeTimRebuildManager(t)
	for _, mask := range [][]bool{nil, {true}, {true, true, true, true, true}} {
		if result, err := m.TrimBoxes(mask); err == nil || result != nil {
			t.Fatal(mask, err)
		}
	}
	for _, test := range []struct {
		boxes []int
		diff  int
	}{{nil, -1}, {nil, 3}, {[]int{-1}, 0}, {[]int{4}, 0}, {[]int{0, 1, 2, 3, 0}, 0}} {
		if result, err := m.ReduceBoxes(test.boxes, test.diff); err == nil || result != nil {
			t.Fatal(test, err)
		}
	}
	if result, err := AlignTimBoxes(m, m); err == nil || result != nil {
		t.Fatal(result, err)
	}
	implementation, _ := m.ReduceBoxes([]int{0, 1}, 0)
	implementation.boxes[0].Copy = -1
	if result, err := AlignTimBoxes(m, implementation); err == nil || result != nil {
		t.Fatal(result, err)
	}
	implementation.boxes[0].Copy = 1
	implementation.boxes[1].Copy = 1
	if result, err := AlignTimBoxes(m, implementation); err == nil || result != nil {
		t.Fatal(result, err)
	}
	noncanonical, _ := NewTimManager(3, 3, 6)
	noncanonical.AddBox(TimBox{FirstInput: 1, Inputs: 1, FirstOutput: 1, Outputs: 1}, TimBoxOptions{})
	if result, err := noncanonical.TrimBoxes([]bool{false}); err == nil || result != nil {
		t.Fatal(result, err)
	}
}

func FuzzTimDuplicate(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6}, int32(0))
	f.Add([]byte{255, 0, 127, 128}, int32(3))
	f.Fuzz(func(t *testing.T, data []byte, unit int32) {
		if len(data) == 0 {
			return
		}
		m, _ := NewTimManager(4, 4, 8)
		m.AddBox(TimBox{Inputs: 2, FirstOutput: 2, Outputs: 2}, TimBoxOptions{})
		delays := make([]float32, 4)
		for i := range delays {
			delays[i] = float32(int8(data[i%len(data)])) / 4
			if data[i%len(data)]&15 == 0 {
				delays[i] = TimNoArc
			}
		}
		m.SetDelayTable(0, TimDelayTable{Inputs: 2, Outputs: 2, Delays: delays}, TimDelayOptions{})
		m.epoch = int(data[0])
		m.useTraversal = data[0]&1 != 0
		for i := range m.cis {
			m.cis[i].TravID = int(data[i%len(data)])
			m.cos[i].TravID = int(data[(i+1)%len(data)])
			m.cis[i].Arrival = float32(int8(data[(i+2)%len(data)]))
			m.cos[i].Required = float32(int8(data[(i+3)%len(data)]))
		}
		before := timManagerWords(m)
		copy, err := m.Duplicate(int(unit))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, timManagerWords(m)) {
			t.Fatal("source mutation")
		}
		if copy.epoch != 0 || !copy.useTraversal {
			t.Fatal(copy.epoch, copy.useTraversal)
		}
		for _, terminal := range append(slices.Clone(copy.cis), copy.cos...) {
			if terminal.TravID != 0 {
				t.Fatal("stale traversal", terminal)
			}
		}
		for i, delay := range copy.tables[0].Delays {
			want := delays[i]
			if unit != 0 && want != TimNoArc {
				want = float32(unit)
			}
			if math.Float32bits(delay) != math.Float32bits(want) {
				t.Fatal(i, delay, want)
			}
		}
		if unit != 0 {
			for _, terminal := range copy.cis[:2] {
				if terminal.Arrival != 0 {
					t.Fatal(terminal)
				}
			}
			for _, terminal := range copy.cos[2:] {
				if terminal.Required != TimEternity {
					t.Fatal(terminal)
				}
			}
		}
	})
}

func FuzzTimRebuild(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8}, uint8(0x55))
	f.Add([]byte{255, 0, 17, 42}, uint8(0xaa))
	f.Fuzz(func(t *testing.T, data []byte, flags uint8) {
		if len(data) == 0 {
			return
		}
		boxCount := 1 + int(flags&3)
		nPis, nPos := int(flags>>2&3), int(flags>>4&3)
		inputs, outputs := make([]int, boxCount), make([]int, boxCount)
		totalInputs, totalOutputs := 0, 0
		for i := 0; i < boxCount; i++ {
			inputs[i] = 1 + int(data[(2*i)%len(data)]%3)
			outputs[i] = 1 + int(data[(2*i+1)%len(data)]%3)
			totalInputs += inputs[i]
			totalOutputs += outputs[i]
		}
		m, _ := NewTimManager(nPis+totalOutputs, totalInputs+nPos, 32)
		firstCI, firstCO := nPis, 0
		for box := 0; box < boxCount; box++ {
			m.AddBox(TimBox{FirstInput: firstCO, Inputs: inputs[box], FirstOutput: firstCI, Outputs: outputs[box], DelayTable: box, Black: data[box%len(data)]&1 != 0}, TimBoxOptions{})
			copyID := int(int8(data[(box+3)%len(data)]))
			if data[(box+1)%len(data)]&1 == 0 {
				copyID = -1
			}
			m.SetBoxCopy(box, copyID)
			delays := make([]float32, inputs[box]*outputs[box])
			for i := range delays {
				delays[i] = float32(int8(data[(box+i)%len(data)]))
			}
			m.SetDelayTable(box, TimDelayTable{Inputs: inputs[box], Outputs: outputs[box], Delays: delays}, TimDelayOptions{})
			firstCI += outputs[box]
			firstCO += inputs[box]
		}
		for i := 0; i < nPis; i++ {
			m.cis[i].Arrival = float32(int8(data[i%len(data)]))
		}
		for i := 0; i < nPos; i++ {
			m.cos[totalInputs+i].Required = float32(int8(data[(i+2)%len(data)]))
		}
		before := timManagerWords(m)
		mask := make([]bool, boxCount)
		selected := make([]int, 0, boxCount)
		for box := range mask {
			mask[box] = flags&(1<<uint((box+6)%8)) != 0
			if mask[box] {
				selected = append(selected, box)
			}
		}
		trimmed, err := m.TrimBoxes(mask)
		if err != nil {
			t.Fatal(err)
		}
		if len(selected) == boxCount {
			for box := range selected {
				if trimmed.boxes[box].Copy != m.boxes[box].Copy {
					t.Fatal("all-box copy", box)
				}
			}
		} else {
			for newID, oldID := range selected {
				wantCopy := m.boxes[oldID].Copy
				if wantCopy == -1 {
					wantCopy = oldID
				}
				if trimmed.boxes[newID].Copy != wantCopy || trimmed.boxes[newID].Inputs != inputs[oldID] || trimmed.boxes[newID].Outputs != outputs[oldID] {
					t.Fatal("trim model", newID, oldID, trimmed.boxes[newID])
				}
			}
		}
		if _, _, err = trimmed.timCanonicalLayout(); err != nil {
			t.Fatal(err)
		}
		order := make([]int, boxCount)
		for i := range order {
			order[i] = boxCount - 1 - i
		}
		diff := min(nPis, nPos)
		if diff > 0 {
			diff = int(flags>>6) % (diff + 1)
		}
		reduced, err := m.ReduceBoxes(order, diff)
		if err != nil {
			t.Fatal(err)
		}
		for newID, oldID := range order {
			wantCopy := m.boxes[oldID].Copy
			if wantCopy == -1 {
				wantCopy = oldID
			}
			if reduced.boxes[newID].Copy != wantCopy || reduced.boxes[newID].Inputs != inputs[oldID] || reduced.boxes[newID].Outputs != outputs[oldID] {
				t.Fatal("reduce model", newID, oldID, reduced.boxes[newID])
			}
		}
		if _, _, err = reduced.timCanonicalLayout(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, timManagerWords(m)) {
			t.Fatal("source mutation")
		}
	})
}
