package abc

import (
	"encoding/json"
	"math"
	"os"
	"reflect"
	"slices"
	"testing"
)

type timConfigureVector struct {
	Name              string
	Boxes             [][2]int
	Tables            [][]uint32
	Arrival, Required uint32
}

func makeTimConfigureManager(t testing.TB) *TimManager {
	t.Helper()
	inputs := []int{1, 2, 1, 1}
	outputs := []int{2, 1, 2, 1}
	tables := []int{0, 2, 0, -1}
	m, err := NewTimManager(7, 6, 13)
	if err != nil {
		t.Fatal(err)
	}
	firstCI, firstCO := 1, 0
	for box := range inputs {
		if _, err = m.AddBox(TimBox{FirstInput: firstCO, Inputs: inputs[box], FirstOutput: firstCI, Outputs: outputs[box], DelayTable: tables[box], Black: box&1 != 0}, TimBoxOptions{}); err != nil {
			t.Fatal(err)
		}
		firstCI += outputs[box]
		firstCO += inputs[box]
	}
	return m
}

func timConfigureSummary(name string, m *TimManager) timConfigureVector {
	result := timConfigureVector{Name: name, Boxes: make([][2]int, len(m.boxes)), Tables: make([][]uint32, len(m.tables))}
	for i, box := range m.boxes {
		black := 0
		if box.Black {
			black = 1
		}
		result.Boxes[i] = [2]int{box.DelayTable, black}
	}
	for i, table := range m.tables {
		if table.Inputs == 0 {
			continue
		}
		result.Tables[i] = []uint32{uint32(table.Inputs), uint32(table.Outputs)}
		for _, delay := range table.Delays {
			result.Tables[i] = append(result.Tables[i], math.Float32bits(delay))
		}
	}
	result.Arrival = math.Float32bits(m.cis[0].Arrival)
	result.Required = math.Float32bits(m.cos[5].Required)
	return result
}

func TestTimConfigureNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/tim_configure.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture []timConfigureVector
	if err = json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture) != 2 {
		t.Fatal(len(fixture))
	}
	arrivals, required := []float32{1.5}, []float32{9.25}
	unit := makeTimConfigureManager(t)
	if err = unit.Configure(nil, arrivals, required, TimDelayOptions{}); err != nil {
		t.Fatal(err)
	}
	libraryValues0 := []float32{2, 3}
	libraryValues2 := []float32{-4, 7}
	library := []TimLibraryBox{
		{Inputs: 1, Outputs: 2, Delays: libraryValues0, Black: true},
		{},
		{Inputs: 2, Outputs: 1, Delays: libraryValues2},
	}
	configured := makeTimConfigureManager(t)
	if err = configured.Configure(library, arrivals, required, TimDelayOptions{}); err != nil {
		t.Fatal(err)
	}
	actual := []timConfigureVector{timConfigureSummary("unit", unit), timConfigureSummary("library", configured)}
	if unit.DelayTableCount() != 4 || configured.DelayTableCount() != 4 || (*TimManager)(nil).DelayTableCount() != 0 {
		t.Fatal("delay-table count")
	}
	if !reflect.DeepEqual(actual, fixture) {
		t.Fatal(actual, fixture)
	}
	arrivals[0], required[0], libraryValues0[0], libraryValues2[0] = 99, 99, 99, 99
	if configured.cis[0].Arrival != 1.5 || configured.cos[5].Required != 9.25 || configured.tables[0].Delays[0] != 2 || configured.tables[2].Delays[0] != -4 {
		t.Fatal("configuration aliases caller storage")
	}
}

func TestTimConfigureErrorsAreAtomic(t *testing.T) {
	library := []TimLibraryBox{{Inputs: 1, Outputs: 2, Delays: []float32{2, 3}}, {}, {Inputs: 2, Outputs: 1, Delays: []float32{-4, 7}}}
	tests := []struct {
		mutate   func(*TimManager)
		library  []TimLibraryBox
		arr, req []float32
		options  TimDelayOptions
	}{
		{library: library, arr: []float32{}, req: []float32{1}},
		{library: library, arr: []float32{1}, req: []float32{}},
		{library: library[:2], arr: []float32{1}, req: []float32{1}},
		{library: []TimLibraryBox{{Inputs: 2, Outputs: 2, Delays: make([]float32, 4)}, {}, library[2]}, arr: []float32{1}, req: []float32{1}},
		{library: []TimLibraryBox{{Inputs: 1, Outputs: 2, Delays: []float32{1}}, {}, library[2]}, arr: []float32{1}, req: []float32{1}},
		{library: library, arr: []float32{1}, req: []float32{1}, options: TimDelayOptions{MaxTables: 3}},
		{library: library, arr: []float32{1}, req: []float32{1}, options: TimDelayOptions{MaxValues: 4}},
		{library: library, arr: []float32{1}, req: []float32{1}, options: TimDelayOptions{MaxTables: -1}},
		{mutate: func(m *TimManager) {
			m.SetDelayTable(0, TimDelayTable{Inputs: 1, Outputs: 1, Delays: []float32{1}}, TimDelayOptions{})
		}, library: library, arr: []float32{1}, req: []float32{1}},
	}
	for index, test := range tests {
		m := makeTimConfigureManager(t)
		if test.mutate != nil {
			test.mutate(m)
		}
		before := timManagerWords(m)
		if err := m.Configure(test.library, test.arr, test.req, test.options); err == nil || !reflect.DeepEqual(before, timManagerWords(m)) {
			t.Fatal(index, err)
		}
	}
	var missing *TimManager
	if err := missing.Configure(nil, nil, nil, TimDelayOptions{}); err == nil {
		t.Fatal("nil manager")
	}
}

func FuzzTimConfigure(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, uint8(3))
	f.Add([]byte{255, 0, 17}, uint8(0))
	f.Fuzz(func(t *testing.T, data []byte, flags uint8) {
		if len(data) == 0 {
			return
		}
		boxCount := 1 + int(flags%4)
		inputs, outputs := make([]int, boxCount), make([]int, boxCount)
		totalInputs, totalOutputs := 0, 0
		for i := 0; i < boxCount; i++ {
			inputs[i] = 1 + int(data[(2*i)%len(data)]%3)
			outputs[i] = 1 + int(data[(2*i+1)%len(data)]%3)
			totalInputs += inputs[i]
			totalOutputs += outputs[i]
		}
		m, _ := NewTimManager(1+totalOutputs, totalInputs+1, 32)
		library := make([]TimLibraryBox, boxCount)
		firstCI, firstCO := 1, 0
		for box := 0; box < boxCount; box++ {
			m.AddBox(TimBox{FirstInput: firstCO, Inputs: inputs[box], FirstOutput: firstCI, Outputs: outputs[box], DelayTable: box}, TimBoxOptions{})
			values := make([]float32, inputs[box]*outputs[box])
			for i := range values {
				values[i] = float32(int8(data[(box+i)%len(data)])) / 2
			}
			library[box] = TimLibraryBox{Inputs: inputs[box], Outputs: outputs[box], Delays: values, Black: data[box%len(data)]&1 != 0}
			firstCI += outputs[box]
			firstCO += inputs[box]
		}
		before := timManagerWords(m)
		if err := m.Configure(library, []float32{float32(int8(data[0]))}, []float32{float32(int8(data[len(data)-1]))}, TimDelayOptions{MaxTables: 8, MaxValues: 64}); err != nil {
			t.Fatal(err)
		}
		for box := range m.boxes {
			if m.boxes[box].DelayTable != box || m.boxes[box].Black != library[box].Black || !slices.Equal(m.tables[box].Delays, library[box].Delays) {
				t.Fatal("library model", box)
			}
		}
		if before.Tables != nil && len(before.Tables) != 0 {
			t.Fatal("unexpected setup tables")
		}
	})
}
