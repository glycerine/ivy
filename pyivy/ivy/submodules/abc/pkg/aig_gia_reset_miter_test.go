package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaResetComparisonMiterTraces(t *testing.T) {
	m := NewGiaManager()
	m.Name = "compare"
	m.Spec = "compare.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	m.AppendCO(b.Not())
	m.AppendCO(x)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for _, spec := range []string{"x00", "X01", "xxx", "xXX", "Xx1", "X1X"} {
		r, e := m.ResetComparisonMiter(spec, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if r.Graph.constraints != 0 || r.Graph.PrimaryOutputCount() != 2 || r.Graph.registers != 2*r.RegistersPerSide || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source", r)
		}
		for sequence := 0; sequence < 64; sequence++ {
			state0, state1 := make([]bool, 2), make([]bool, 2)
			for i := 0; i < 2; i++ {
				state0[i] = spec[1+i] == '1'
				state1[i] = state0[i]
			}
			inputs := make([]bool, r.Graph.PrimaryInputCount())
			for i := range inputs {
				inputs[i] = sequence&(1<<i) != 0
			}
			for i, index := range r.UnknownRegisters {
				state0[index] = inputs[r.FirstInputs[1+i]]
				state1[index] = inputs[r.SecondInputs[1+i]]
			}
			target := make([]bool, r.Graph.registers)
			for frame := 0; frame < 6; frame++ {
				inputs[r.FirstInputs[0]] = sequence&(1<<frame) != 0
				if r.SecondInputs[0] != r.FirstInputs[0] {
					inputs[r.SecondInputs[0]] = sequence&(1<<((frame+2)%6)) != 0
				}
				for i := range r.UnknownRegisters {
					if frame > 0 {
						inputs[r.FirstInputs[1+i]] = frame%2 == 0
						inputs[r.SecondInputs[1+i]] = frame%2 != 0
					}
				}
				p0, p1 := inputs[r.FirstInputs[0]], inputs[r.SecondInputs[0]]
				expected := []bool{state0[1] != state1[1], (p0 != state0[0]) != (p1 != state1[0])}
				got, next, e := r.Graph.Step(inputs, target)
				if e != nil || !slices.Equal(got, expected) {
					t.Fatal(spec, sequence, frame, got, expected, e)
				}
				state0 = []bool{p0 != state0[0], !state0[0]}
				state1 = []bool{p1 != state1[0], !state1[0]}
				target = next
			}
		}
	}
	for _, spec := range []string{"", "000", "x02", "x0?", "xxxx"} {
		if _, e := m.ResetComparisonMiter(spec, GiaCofactorOptions{}); e == nil {
			t.Fatal(spec)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.ResetComparisonMiter("xxx", options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 5, 15, 30} {
		calls := 0
		r, e := m.ResetComparisonMiter("xxx", GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaResetComparison(f *testing.F) {
	f.Add(uint8(3), uint8(42))
	f.Fuzz(func(t *testing.T, selection, sequence uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		m.AppendCO(a)
		m.AppendCO(x)
		m.SetRegisterCount(1)
		spec := []string{"x0", "X1", "xx", "xX", "XX"}[int(selection)%5]
		r, e := m.ResetComparisonMiter(spec, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		inputs := make([]bool, r.Graph.PrimaryInputCount())
		for i := range inputs {
			inputs[i] = sequence&(1<<i) != 0
		}
		a0, a1 := spec[1] == '1', spec[1] == '1'
		if len(r.UnknownRegisters) > 0 {
			a0 = inputs[r.FirstInputs[1]]
			a1 = inputs[r.SecondInputs[1]]
		}
		state := make([]bool, r.Graph.registers)
		for frame := 0; frame < 8; frame++ {
			inputs[r.FirstInputs[0]] = sequence&(1<<frame) != 0
			if r.SecondInputs[0] != r.FirstInputs[0] {
				inputs[r.SecondInputs[0]] = sequence&(1<<((frame+3)%8)) != 0
			}
			got, next, e := r.Graph.Step(inputs, state)
			if e != nil || got[0] != (a0 != a1) {
				t.Fatal(spec, frame, got, a0, a1, e)
			}
			a0 = inputs[r.FirstInputs[0]] != a0
			a1 = inputs[r.SecondInputs[0]] != a1
			state = next
		}
	})
}

func TestGiaResetComparisonNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	for _, l := range []GiaLit{b.Not(), x, x, a.Not()} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(2)
	raw, e := os.ReadFile("test_vectors/oracle/gia_reset_miter.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects, Inputs, Registers int
		Drivers                    []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 6 {
		t.Fatal(e)
	}
	specs := []string{"x00", "X01", "xxx", "xXX", "Xx1", "X1X"}
	for mode, row := range rows {
		r, e := m.ResetComparisonMiter(specs[mode], GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range r.Graph.cos {
			l, _ := r.Graph.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if r.Graph.ObjectCount() != row.Objects || r.Graph.InputCount() != row.Inputs || r.Graph.registers != row.Registers || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(specs[mode], drivers, row, r.Graph.ObjectCount())
		}
	}
}
