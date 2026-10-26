package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaStructuralSweepModesAndTraces(t *testing.T) {
	m := NewGiaManager()
	m.Name = "sweep"
	m.Spec = "sweep.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	unused, _ := m.AppendCI()
	x, _ := m.AppendXorReal(b, c)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.AppendCO(a)
	m.AppendCO(p)
	m.AppendCO(p)
	m.AppendCO(unused)
	m.SetRegisterCount(4)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for _, constant := range []bool{false, true} {
		for _, equiv := range []bool{false, true} {
			r, e := m.SweepZeroInitializedRegisters(GiaStructuralSweepOptions{ReduceConstants: constant, MergeEquivalent: equiv})
			if e != nil {
				t.Fatal(e)
			}
			if r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || r.Graph.constraints != 1 || !slices.Equal(before, m.objects) || !slices.Equal(r.PrimaryInputs, []int{0}) {
				t.Fatal("metadata/source", r)
			}
			for sequence := 0; sequence < 64; sequence++ {
				sourceState := make([]bool, 4)
				targetState := make([]bool, r.Graph.registers)
				for frame := 0; frame < 6; frame++ {
					input := []bool{sequence&(1<<frame) != 0}
					want, next, e := m.Step(input, sourceState)
					if e != nil {
						t.Fatal(e)
					}
					got, nextTarget, e := r.Graph.Step(input, targetState)
					if e != nil || !slices.Equal(got, want) {
						t.Fatal(constant, equiv, sequence, frame, got, want, e)
					}
					for i, index := range r.Registers {
						if nextTarget[i] != next[index] {
							t.Fatal("state projection", i, index)
						}
					}
					sourceState, targetState = next, nextTarget
				}
			}
			if !equiv && !slices.Equal(r.Registers, []int{0, 1, 2}) {
				t.Fatal("initial cleanup projection", r.Registers)
			}
			if equiv && !slices.Equal(r.Registers, []int{0}) && !constant {
				t.Fatal("structural projection", r.Registers)
			}
		}
	}
	for _, options := range []GiaStructuralSweepOptions{{MaxIterations: -1}, {GiaTernaryHistoryOptions: GiaTernaryHistoryOptions{MaxStates: -1}}, {GiaTernaryHistoryOptions: GiaTernaryHistoryOptions{GiaCofactorOptions: GiaCofactorOptions{MaxObjects: 1}}}, {GiaTernaryHistoryOptions: GiaTernaryHistoryOptions{GiaCofactorOptions: GiaCofactorOptions{MaxMappingEntries: 1}}}, {MergeEquivalent: true, MaxIterations: 1}, {ReduceConstants: true, GiaTernaryHistoryOptions: GiaTernaryHistoryOptions{MaxStates: 1}}} {
		r, e := m.SweepZeroInitializedRegisters(options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 5, 15, 30} {
		calls := 0
		r, e := m.SweepZeroInitializedRegisters(GiaStructuralSweepOptions{ReduceConstants: true, MergeEquivalent: true, GiaTernaryHistoryOptions: GiaTernaryHistoryOptions{GiaCofactorOptions: GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }}}})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	comb := NewGiaManager()
	a, _ = comb.AppendCI()
	b, _ = comb.AppendCI()
	comb.AppendAnd(a, b)
	comb.AppendCO(a)
	r, e := comb.SweepZeroInitializedRegisters(GiaStructuralSweepOptions{})
	if e != nil || r.Graph.AndCount() != 0 || r.Graph.InputCount() != 2 {
		t.Fatal("combinational cleanup", r, e)
	}
}
func FuzzGiaStructuralSweep(f *testing.F) {
	f.Add(uint8(3), uint8(42))
	f.Fuzz(func(t *testing.T, mode, sequence uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(b, c)
		m.AppendCO(x)
		m.AppendCO(a.Not())
		m.AppendCO(a)
		m.AppendCO(p)
		m.AppendCO(p)
		m.SetRegisterCount(3)
		r, e := m.SweepZeroInitializedRegisters(GiaStructuralSweepOptions{ReduceConstants: mode&1 != 0, MergeEquivalent: mode&2 != 0})
		if e != nil {
			t.Fatal(e)
		}
		state := make([]bool, r.Graph.registers)
		for frame := 0; frame < 8; frame++ {
			p := sequence&(1<<frame) != 0
			outputs, next, e := r.Graph.Step([]bool{p}, state)
			if e != nil || !slices.Equal(outputs, []bool{false, true}) {
				t.Fatal(frame, outputs, e)
			}
			state = next
		}
	})
}

func TestGiaStructuralSweepNative(t *testing.T) {
	raw, e := os.ReadFile("test_vectors/oracle/gia_structural_sweep.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects, Registers int
		Drivers            []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 4 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		unused, _ := m.AppendCI()
		x, _ := m.AppendAnd(b, c)
		m.AppendCO(x)
		m.AppendCO(a.Not())
		m.AppendCO(a)
		m.AppendCO(p)
		m.AppendCO(p)
		m.AppendCO(unused)
		m.SetRegisterCount(4)
		r, e := m.SweepZeroInitializedRegisters(GiaStructuralSweepOptions{ReduceConstants: mode/2 != 0, MergeEquivalent: mode%2 != 0})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range r.Graph.cos {
			l, _ := r.Graph.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if r.Graph.ObjectCount() != row.Objects || r.Graph.registers != row.Registers || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(mode, drivers, row, r.Graph.ObjectCount(), r.Graph.registers)
		}
	}
}
