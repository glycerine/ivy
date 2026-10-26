package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaGeneralTrimProjection(t *testing.T) {
	m := NewGiaManager()
	m.Name = "trim"
	m.Spec = "trim.aig"
	p, _ := m.AppendCI()
	m.AppendCI()
	a, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	m.AppendCO(GiaConst0)
	m.AppendCO(GiaConst1)
	m.AppendCO(x)
	m.AppendCO(p)
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for _, trimInputs := range []bool{false, true} {
		for _, trimOutputs := range []bool{false, true} {
			for _, phase := range []int{-1, 0, 1} {
				r, e := m.Trim(GiaTrimOptions{TrimInputs: trimInputs, TrimOutputs: trimOutputs, RemoveConstant: phase})
				if e != nil {
					t.Fatal(e)
				}
				if r.Graph.constraints != 0 || r.Graph.registers != 1 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || !slices.Equal(r.Registers, []int{0}) || !slices.Equal(before, m.objects) {
					t.Fatal("metadata/source")
				}
				expected := []int{0, 1, 2, 3}
				if trimOutputs {
					expected = []int{2, 3}
					if phase == 0 {
						expected = []int{1, 2, 3}
					} else if phase == 1 {
						expected = []int{0, 2, 3}
					}
				}
				if !slices.Equal(r.Outputs, expected) {
					t.Fatal(trimOutputs, phase, r.Outputs, expected)
				}
				for k := 0; k < 8; k++ {
					source := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
					input := make([]bool, len(r.Inputs))
					for i, index := range r.Inputs {
						if index >= 0 {
							input[i] = source[index]
						}
					}
					want, e := m.Evaluate(source)
					if e != nil {
						t.Fatal(e)
					}
					got, e := r.Graph.Evaluate(input)
					if e != nil {
						t.Fatal(e)
					}
					for i, index := range r.Outputs {
						if got[i] != want[index] {
							t.Fatal(k, got, want)
						}
					}
				}
			}
		}
	}
	for _, options := range []GiaTrimOptions{{RemoveConstant: 2}, {GiaStructuralSweepOptions: GiaStructuralSweepOptions{GiaTernaryHistoryOptions: GiaTernaryHistoryOptions{GiaCofactorOptions: GiaCofactorOptions{MaxObjects: 1}}}}, {GiaStructuralSweepOptions: GiaStructuralSweepOptions{GiaTernaryHistoryOptions: GiaTernaryHistoryOptions{GiaCofactorOptions: GiaCofactorOptions{MaxMappingEntries: 1}}}}} {
		r, e := m.Trim(options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 3, 6, 9} {
		calls := 0
		r, e := m.Trim(GiaTrimOptions{GiaStructuralSweepOptions: GiaStructuralSweepOptions{GiaTernaryHistoryOptions: GiaTernaryHistoryOptions{GiaCofactorOptions: GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }}}}})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func TestGiaGeneralTrimDualSweepAndDummies(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	for _, l := range []GiaLit{GiaConst1, GiaConst0, a, GiaConst0, p, b, a, p} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(2)
	r, e := m.Trim(GiaTrimOptions{TrimInputs: true, TrimOutputs: true, DualOutputs: true})
	if e != nil {
		t.Fatal(e)
	}
	if !slices.Equal(r.Outputs, []int{4, 5, 7}) || !slices.Equal(r.Registers, []int{1}) || !slices.Equal(r.Inputs, []int{0, 2}) {
		t.Fatal(r)
	}
	for sequence := 0; sequence < 64; sequence++ {
		state := []bool{false}
		previous := false
		for frame := 0; frame < 6; frame++ {
			p := sequence&(1<<frame) != 0
			out, next, e := r.Graph.Step([]bool{p}, state)
			if e != nil || !slices.Equal(out, []bool{p, previous}) || !slices.Equal(next, []bool{p}) {
				t.Fatal(frame, out, next, e)
			}
			state = next
			previous = p
		}
	}
	dummy := NewGiaManager()
	dummy.AppendCO(GiaConst0)
	dummy.AppendCO(GiaConst1)
	r, e = dummy.Trim(GiaTrimOptions{TrimInputs: true, TrimOutputs: true, DualOutputs: true})
	if e != nil || !slices.Equal(r.Inputs, []int{-1}) || !slices.Equal(r.Outputs, []int{-1, -1}) {
		t.Fatal("dual dummies", r, e)
	}
	single := NewGiaManager()
	single.AppendCO(GiaConst0)
	r, e = single.Trim(GiaTrimOptions{TrimOutputs: true})
	if e != nil || !slices.Equal(r.Outputs, []int{-1}) {
		t.Fatal(r, e)
	}
}
func FuzzGiaGeneralTrimming(f *testing.F) {
	f.Add(uint8(3), uint8(5))
	f.Fuzz(func(t *testing.T, flags, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		m.AppendCI()
		r, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, r)
		m.AppendCO(GiaConst0)
		m.AppendCO(GiaConst1)
		m.AppendCO(x)
		m.AppendCO(p)
		m.SetRegisterCount(1)
		result, e := m.Trim(GiaTrimOptions{TrimInputs: flags&1 != 0, TrimOutputs: flags&2 != 0, RemoveConstant: int(flags%3) - 1})
		if e != nil {
			t.Fatal(e)
		}
		source := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		input := make([]bool, len(result.Inputs))
		for i, index := range result.Inputs {
			if index >= 0 {
				input[i] = source[index]
			}
		}
		want := []bool{false, true, source[0] != source[2], source[0]}
		got, e := result.Graph.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		for i, index := range result.Outputs {
			if index >= 0 && got[i] != want[index] {
				t.Fatal(got, want)
			}
		}
	})
}

func TestGiaGeneralTrimNative(t *testing.T) {
	type record struct {
		Objects, Inputs, Registers int
		Drivers                    []GiaLit
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_trim.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct {
		Ordinary []record
		Dual     record
	}
	if e = json.Unmarshal(raw, &native); e != nil || len(native.Ordinary) != 12 {
		t.Fatal(e)
	}
	check := func(g *GiaManager, row record) {
		t.Helper()
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if g.ObjectCount() != row.Objects || g.InputCount() != row.Inputs || g.registers != row.Registers || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(drivers, row, g.ObjectCount(), g.InputCount(), g.registers)
		}
	}
	for mode, row := range native.Ordinary {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		m.AppendCI()
		r, _ := m.AppendCI()
		x, _ := m.AppendAnd(p, r)
		for _, l := range []GiaLit{GiaConst0, GiaConst1, x, p} {
			m.AppendCO(l)
		}
		m.SetRegisterCount(1)
		result, e := m.Trim(GiaTrimOptions{TrimInputs: mode/6 != 0, TrimOutputs: (mode%6)/3 != 0, RemoveConstant: mode%3 - 1})
		if e != nil {
			t.Fatal(e)
		}
		check(result.Graph, row)
	}
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	for _, l := range []GiaLit{GiaConst1, GiaConst0, a, GiaConst0, p, b, a, p} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(2)
	result, e := m.Trim(GiaTrimOptions{TrimInputs: true, TrimOutputs: true, DualOutputs: true})
	if e != nil {
		t.Fatal(e)
	}
	check(result.Graph, native.Dual)
}
func FuzzGiaDualTrimming(f *testing.F) {
	f.Add(uint8(3), uint8(42))
	f.Fuzz(func(t *testing.T, phase, sequence uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		for _, l := range []GiaLit{GiaConst1, GiaConst0, a, GiaConst0, p, b, a.NotIf(phase&2 != 0), p.NotIf(phase&1 != 0)} {
			m.AppendCO(l)
		}
		m.SetRegisterCount(2)
		r, e := m.Trim(GiaTrimOptions{TrimInputs: true, TrimOutputs: true, DualOutputs: true})
		if e != nil {
			t.Fatal(e)
		}
		sourceState := []bool{false, false}
		targetState := make([]bool, r.Graph.registers)
		for frame := 0; frame < 8; frame++ {
			pv := sequence&(1<<frame) != 0
			input := make([]bool, r.Graph.PrimaryInputCount())
			for i, index := range r.Inputs[:len(input)] {
				if index == 0 {
					input[i] = pv
				}
			}
			got, next, e := r.Graph.Step(input, targetState)
			if e != nil {
				t.Fatal(e)
			}
			sourceOutputs := []bool{true, false, sourceState[0], false, pv, sourceState[1]}
			for i, index := range r.Outputs[:len(got)] {
				want := false
				if index >= 0 {
					want = sourceOutputs[index]
				}
				if got[i] != want {
					t.Fatal(frame, i, index, got, sourceOutputs)
				}
			}
			sourceState = []bool{sourceState[0] != (phase&2 != 0), pv != (phase&1 != 0)}
			for i, index := range r.Registers {
				if next[i] != sourceState[index] {
					t.Fatal("dual reset/state projection", frame, i, index)
				}
			}
			targetState = next
		}
	})
}
