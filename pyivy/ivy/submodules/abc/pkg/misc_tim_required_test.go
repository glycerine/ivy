package abc

import (
	"encoding/json"
	"math"
	"os"
	"slices"
	"testing"
)

func TestTimRequiredNativeAndCorrection(t *testing.T) {
	data, err := os.ReadFile("test_vectors/oracle/tim_required.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct {
		Inputs, Outputs, NativeDefined         int
		Table, Requirements, Native, Corrected []uint32
	}
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 128 {
		t.Fatal(len(rows))
	}
	differences, defined := 0, 0
	for index, row := range rows {
		m, _ := NewTimManager(row.Outputs+2, row.Inputs+2, 0)
		delays := make([]float32, len(row.Table))
		req := make([]float32, len(row.Requirements))
		for i, bits := range row.Table {
			delays[i] = math.Float32frombits(bits)
		}
		for i, bits := range row.Requirements {
			req[i] = math.Float32frombits(bits)
		}
		m.SetDelayTable(0, TimDelayTable{Inputs: row.Inputs, Outputs: row.Outputs, Delays: delays}, TimDelayOptions{})
		m.AddBox(TimBox{FirstInput: 1, Inputs: row.Inputs, FirstOutput: 1, Outputs: row.Outputs}, TimBoxOptions{})
		m.IncrementTraversal()
		for i, value := range req {
			m.SetCIRequired(i+1, value)
		}
		for input, want := range row.Corrected {
			value, err := m.CORequired(input + 1)
			if err != nil || math.Float32bits(value) != want {
				t.Fatal(index, input, value, want, err)
			}
		}
		if row.NativeDefined != 0 {
			defined++
			native := make([]uint32, row.Inputs)
			for input := range native {
				best := TimEternity
				for output, value := range req {
					if delays[output*row.Inputs+output] != TimNoArc {
						candidate := float32(value - delays[output*row.Inputs+input])
						if !(best < candidate) {
							best = candidate
						}
					}
				}
				native[input] = math.Float32bits(best)
			}
			if !slices.Equal(native, row.Native) {
				t.Fatal("native guard regression", index, native, row.Native)
			}
			if !slices.Equal(row.Native, row.Corrected) {
				differences++
			}
		} else if len(row.Native) != 0 {
			t.Fatal("undefined native reference was executed")
		}
	}
	if defined != 80 || differences == 0 {
		t.Fatal("missing native bug evidence", defined, differences)
	}
	t.Logf("%d of %d defined native cases differ after the disabled-arc correction", differences, defined)
}
func TestTimRequiredGuardCacheAndIEEE(t *testing.T) {
	m, _ := NewTimManager(2, 2, 4)
	m.AddBox(TimBox{Inputs: 2, Outputs: 2}, TimBoxOptions{})
	m.SetDelayTable(0, TimDelayTable{Inputs: 2, Outputs: 2, Delays: []float32{TimNoArc, 1, 2, TimNoArc}}, TimDelayOptions{})
	if v, err := m.CORequired(0); err != nil || v != TimEternity {
		t.Fatal("native epoch-zero cache", v, err)
	}
	m.IncrementTraversal()
	m.SetCIRequired(0, 10)
	m.SetCIRequired(1, 20)
	for input, want := range []float32{18, 9} {
		v, err := m.CORequired(input)
		if err != nil || v != want {
			t.Fatal("corrected cross-arc regression", input, v, err)
		}
	}
	clone := m.Clone()
	clone.SetTraversalEnabled(false)
	clone.SetCIRequired(1, 30)
	v, err := clone.CORequired(0)
	if err != nil || v != 28 || m.cos[0].Required != 18 {
		t.Fatal("clone/cache", v, err)
	}
	m.SetDelayTable(0, TimDelayTable{Inputs: 1, Outputs: 1, Delays: []float32{0}}, TimDelayOptions{})
	v, err = m.CORequired(0)
	if err != nil || v != 18 {
		t.Fatal("cached replacement", v, err)
	}
	m.IncrementTraversal()
	before := timSnapshot(m)
	if _, err = m.CORequired(0); err == nil || !slices.Equal(before, timSnapshot(m)) {
		t.Fatal("shape failure changed marks", err)
	}
	rectangular, _ := NewTimManager(2, 1, 3)
	rectangular.AddBox(TimBox{Inputs: 1, Outputs: 2}, TimBoxOptions{})
	rectangular.SetDelayTable(0, TimDelayTable{Inputs: 1, Outputs: 2, Delays: []float32{1, 3}}, TimDelayOptions{})
	rectangular.IncrementTraversal()
	rectangular.SetCIRequired(0, 10)
	rectangular.SetCIRequired(1, 8)
	v, err = rectangular.CORequired(0)
	if err != nil || v != 5 {
		t.Fatal("more outputs than inputs", v, err)
	}
	rectangular.SetTraversalEnabled(false)
	rectangular.SetCIRequired(0, float32(math.NaN()))
	rectangular.SetCIRequired(1, 8)
	v, err = rectangular.CORequired(0)
	if err != nil || v != 5 {
		t.Fatal("native min overwrites early NaN", v, err)
	}
	rectangular.SetCIRequired(0, 10)
	rectangular.SetCIRequired(1, float32(math.NaN()))
	v, err = rectangular.CORequired(0)
	if err != nil || !math.IsNaN(float64(v)) {
		t.Fatal("native min retains last NaN", v, err)
	}
}
func FuzzTimBoxRequired(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5}, uint8(3))
	f.Add([]byte{255, 0, 17, 42}, uint8(255))
	f.Fuzz(func(t *testing.T, data []byte, flags uint8) {
		if len(data) == 0 {
			return
		}
		ins, outs := 1+int(flags%4), 1+int((flags>>2)%4)
		m, _ := NewTimManager(outs, ins, 0)
		delays := make([]float32, ins*outs)
		req := make([]float32, outs)
		for i := range delays {
			v := data[i%len(data)]
			delays[i] = float32(int8(v)) / 8
			if v&31 == 0 {
				delays[i] = TimNoArc
			}
		}
		for i := range req {
			req[i] = float32(int8(data[(i+3)%len(data)])) / 4
		}
		m.SetDelayTable(0, TimDelayTable{Inputs: ins, Outputs: outs, Delays: delays}, TimDelayOptions{})
		m.AddBox(TimBox{Inputs: ins, Outputs: outs}, TimBoxOptions{})
		m.IncrementTraversal()
		for i, value := range req {
			m.SetCIRequired(i, value)
		}
		for input := 0; input < ins; input++ {
			expected := float64(TimEternity)
			for output, value := range req {
				delay := delays[output*ins+input]
				if delay != TimNoArc {
					expected = math.Min(expected, float64(float32(value-delay)))
				}
			}
			actual, err := m.CORequired(input)
			if err != nil || math.Float32bits(actual) != math.Float32bits(float32(expected)) {
				t.Fatal(input, actual, expected, err)
			}
		}
	})
}
