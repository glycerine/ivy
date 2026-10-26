package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

func TestTimArrivalNative(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/tim_arrival.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Inputs, Outputs, BoxTrav int
		Table, Arrivals, Result  []uint32
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 96 {
		t.Fatal(len(rows))
	}
	for index, row := range rows {
		m, _ := NewTimManager(row.Outputs+2, row.Inputs+2, 0)
		values := make([]float32, len(row.Table))
		for i, bits := range row.Table {
			values[i] = math.Float32frombits(bits)
		}
		if err = m.SetDelayTable(0, TimDelayTable{Inputs: row.Inputs, Outputs: row.Outputs, Delays: values}, TimDelayOptions{}); err != nil {
			t.Fatal(err)
		}
		m.AddBox(TimBox{FirstInput: 1, Inputs: row.Inputs, FirstOutput: 1, Outputs: row.Outputs, DelayTable: 0}, TimBoxOptions{})
		m.IncrementTraversal()
		for i, bits := range row.Arrivals {
			m.SetCOArrival(i+1, math.Float32frombits(bits))
		}
		for i, want := range row.Result {
			value, err := m.CIArrival(i + 1)
			if err != nil || math.Float32bits(value) != want {
				t.Fatal(index, i, value, want, err)
			}
		}
		b, _ := m.Box(0)
		if b.TravID != row.BoxTrav {
			t.Fatal(index, b)
		}
	}
}
func TestTimDelayStorageCacheAndErrors(t *testing.T) {
	m, _ := NewTimManager(2, 2, 4)
	m.AddBox(TimBox{Inputs: 2, Outputs: 2, DelayTable: 3}, TimBoxOptions{})
	if value, err := m.CIArrival(0); err != nil || value != 0 {
		t.Fatal("native epoch-zero cached result", value, err)
	}
	values := []float32{0.5, 1.25, TimNoArc, 2.5}
	table := TimDelayTable{Inputs: 2, Outputs: 2, Delays: values}
	if err := m.SetDelayTable(3, table, TimDelayOptions{MaxTables: 4, MaxValues: 4}); err != nil {
		t.Fatal(err)
	}
	values[0] = 99
	snapshot, err := m.BoxDelayTable(0)
	if err != nil || snapshot.Delays[0] != 0.5 {
		t.Fatal(snapshot, err)
	}
	snapshot.Delays[0] = 88
	m.IncrementTraversal()
	m.SetCOArrival(0, 1.5)
	m.SetCOArrival(1, 2.5)
	value, err := m.CIArrival(0)
	if err != nil || value != 3 {
		t.Fatal(value, err)
	}
	clone := m.Clone()
	clone.tables[3].Delays[0] = 77
	if m.tables[3].Delays[0] != 0.5 {
		t.Fatal("table clone alias")
	}
	m.SetTraversalEnabled(false)
	m.SetCOArrival(1, 7.75)
	value, err = m.CIArrival(0)
	if err != nil || value != 9 {
		t.Fatal("uncached recomputation", value, err)
	}
	m.SetTraversalEnabled(true)
	m.SetDelayTable(3, TimDelayTable{Inputs: 2, Outputs: 2, Delays: []float32{0, 0, 0, 0}}, TimDelayOptions{})
	value, err = m.CIArrival(0)
	if err != nil || value != 9 {
		t.Fatal("cached table mutation", value, err)
	}
	for _, bad := range []float32{float32(math.NaN()), float32(math.Inf(1)), float32(math.Inf(-1)), 2147483648} {
		m.SetDelayTable(3, TimDelayTable{Inputs: 2, Outputs: 2, Delays: []float32{bad, 0, 0, 0}}, TimDelayOptions{})
		m.IncrementTraversal()
		before := timSnapshot(m)
		b, _ := m.Box(0)
		if _, err := m.CIArrival(0); err == nil || !slices.Equal(before, timSnapshot(m)) {
			t.Fatal("nonfinite partial update", bad, err)
		}
		after, _ := m.Box(0)
		if after != b {
			t.Fatal("box marked on failure")
		}
	}
	for _, options := range []TimDelayOptions{{MaxTables: 3}, {MaxValues: 3}, {MaxTables: -1}, {MaxValues: -1}} {
		if err := m.SetDelayTable(3, table, options); err == nil {
			t.Fatal(options)
		}
	}
	if err := m.SetDelayTable(0, TimDelayTable{Inputs: 2, Outputs: 2, Delays: []float32{1}}, TimDelayOptions{}); err == nil {
		t.Fatal("bad shape")
	}
	if _, err := m.DelayTable(0); err == nil {
		t.Fatal("sparse hole")
	}
	m.SetDelayTable(3, TimDelayTable{Inputs: 1, Outputs: 1, Delays: []float32{0}}, TimDelayOptions{})
	if _, err := m.BoxDelayTable(0); err == nil {
		t.Fatal("box/table shape")
	}
}
func FuzzTimBoxArrival(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, uint8(3))
	f.Add([]byte{255, 0, 17, 42}, uint8(255))
	f.Fuzz(func(t *testing.T, data []byte, flags uint8) {
		if len(data) == 0 {
			return
		}
		ins, outs := 1+int(flags%4), 1+int((flags>>2)%4)
		m, _ := NewTimManager(outs, ins, 0)
		delays := make([]float32, ins*outs)
		arrivals := make([]float32, ins)
		for i := range delays {
			v := data[i%len(data)]
			delays[i] = float32(int8(v)) / 8
			if v&31 == 0 {
				delays[i] = TimNoArc
			}
		}
		for i := range arrivals {
			arrivals[i] = float32(int8(data[(i+3)%len(data)])) / 4
		}
		if err := m.SetDelayTable(0, TimDelayTable{Inputs: ins, Outputs: outs, Delays: delays}, TimDelayOptions{}); err != nil {
			t.Fatal(err)
		}
		m.AddBox(TimBox{Inputs: ins, Outputs: outs}, TimBoxOptions{})
		m.IncrementTraversal()
		for i, a := range arrivals {
			m.SetCOArrival(i, a)
		}
		for output := 0; output < outs; output++ {
			expected := int32(-1000000000)
			for input, a := range arrivals {
				delay := delays[output*ins+input]
				if delay != TimNoArc {
					candidate := int32(float32(a + delay))
					if candidate > expected {
						expected = candidate
					}
				}
			}
			actual, err := m.CIArrival(output)
			if err != nil || math.Float32bits(actual) != math.Float32bits(float32(expected)) {
				t.Fatal(output, actual, expected, err)
			}
		}
	})
}
