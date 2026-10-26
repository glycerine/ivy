package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func giaTernaryReductionExample(t *testing.T) *GiaManager {
	t.Helper()
	m := NewGiaManager()
	m.Name = "history"
	m.Spec = "history.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	d, _ := m.AppendCI()
	x, e := m.AppendXorReal(b, c)
	if e != nil {
		t.Fatal(e)
	}
	y, e := m.AppendAnd(a, d)
	if e != nil {
		t.Fatal(e)
	}
	m.AppendCO(x)
	m.AppendCO(y)
	m.AppendCO(a)
	m.AppendCO(b.Not())
	m.AppendCO(c.Not())
	m.AppendCO(p)
	m.SetRegisterCount(4)
	m.SetConstraintCount(1)
	return m
}
func TestGiaTernaryRegisterReductionTrace(t *testing.T) {
	m := giaTernaryReductionExample(t)
	before := slices.Clone(m.objects)
	r, e := m.ReduceTernaryZeroInitializedRegisters(GiaTernaryHistoryOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if r.Constants != 1 || r.Equivalent != 1 || !slices.Equal(r.Replacements, []GiaLit{InvalidGiaLit, GiaConst0, InvalidGiaLit, GiaLit(m.cis[2]) << 1, InvalidGiaLit}) || r.Graph.constraints != 1 || r.Graph.registers != 4 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || !slices.Equal(before, m.objects) {
		t.Fatal(r)
	}
	for sequence := 0; sequence < 256; sequence++ {
		original, reduced := make([]bool, 4), make([]bool, 4)
		for frame := 0; frame < 8; frame++ {
			pi := []bool{sequence&(1<<frame) != 0}
			want, next, e := m.Step(pi, original)
			if e != nil {
				t.Fatal(e)
			}
			got, nextReduced, e := r.Graph.Step(pi, reduced)
			if e != nil || !slices.Equal(got, want) || !slices.Equal(next, nextReduced) {
				t.Fatal(sequence, frame, got, want, e)
			}
			original, reduced = next, nextReduced
		}
	}
	for _, options := range []GiaTernaryHistoryOptions{{MaxStates: 1}, {MaxStateEntries: 1}, {GiaCofactorOptions: GiaCofactorOptions{MaxObjects: 1}}, {GiaCofactorOptions: GiaCofactorOptions{MaxMappingEntries: 1}}} {
		r, e := m.ReduceTernaryZeroInitializedRegisters(options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 15, 30} {
		calls := 0
		r, e := m.ReduceTernaryZeroInitializedRegisters(GiaTernaryHistoryOptions{GiaCofactorOptions: GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }}})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func TestGiaTernaryReductionDoesNotMergeUnknown(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	m.AppendCI()
	m.AppendCO(a)
	m.AppendCO(a)
	m.AppendCO(b)
	m.AppendCO(p)
	m.SetRegisterCount(3)
	r, e := m.ReduceTernaryZeroInitializedRegisters(GiaTernaryHistoryOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if r.Constants != 2 || r.Equivalent != 0 || r.Replacements[3] != InvalidGiaLit {
		t.Fatal(r)
	}
	// Identical unknown histories are not an equality proof for independent PIs.
	m = NewGiaManager()
	p, _ = m.AppendCI()
	q, _ := m.AppendCI()
	a, _ = m.AppendCI()
	b, _ = m.AppendCI()
	m.AppendCO(a)
	m.AppendCO(p)
	m.AppendCO(q)
	m.SetRegisterCount(2)
	r, e = m.ReduceTernaryZeroInitializedRegisters(GiaTernaryHistoryOptions{})
	if e != nil || r.Constants+r.Equivalent != 0 {
		t.Fatal(r, e)
	}
}
func FuzzGiaTernaryRegisterReduction(f *testing.F) {
	f.Add(uint8(3), uint8(42))
	f.Fuzz(func(t *testing.T, phase, sequence uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(b, c)
		m.AppendCO(x)
		m.AppendCO(a.NotIf(phase&1 != 0))
		m.AppendCO(a)
		m.AppendCO(b.Not())
		m.AppendCO(c.Not())
		m.SetRegisterCount(3)
		r, e := m.ReduceTernaryZeroInitializedRegisters(GiaTernaryHistoryOptions{MaxStates: 16})
		if e != nil {
			t.Fatal(e)
		}
		// PI p is deliberately unused; all transitions are deterministic from zero.
		_ = p
		state := []bool{false, false, false}
		for frame := 0; frame < 8; frame++ {
			pv := sequence&(1<<frame) != 0
			v, e := r.Graph.Evaluate([]bool{pv, state[0], state[1], state[2]})
			want := []bool{state[1] != state[2], state[0] != (phase&1 != 0), state[0], !state[1], !state[2]}
			if e != nil || !slices.Equal(v, want) {
				t.Fatal(frame, v, want, e)
			}
			state = []bool{state[0], !state[1], !state[2]}
		}
	})
}

func TestGiaTernaryRegisterReductionNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	d, _ := m.AppendCI()
	x, _ := m.AppendAnd(b, c)
	y, _ := m.AppendAnd(a, d)
	m.AppendCO(x)
	m.AppendCO(y)
	m.AppendCO(a)
	m.AppendCO(b.Not())
	m.AppendCO(c.Not())
	m.AppendCO(p)
	m.SetRegisterCount(4)
	r, e := m.ReduceTernaryZeroInitializedRegisters(GiaTernaryHistoryOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_reduce_ternary.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct {
		Objects int
		Map     []int64
		Drivers []GiaLit
	}
	if e = json.Unmarshal(raw, &native); e != nil {
		t.Fatal(e)
	}
	if len(native.Map) != len(r.Replacements) {
		t.Fatal("mapping dimension")
	}
	for i, value := range native.Map {
		if r.Replacements[i] != GiaLit(value) {
			t.Fatal(i, value, r.Replacements[i])
		}
	}
	drivers := make([]GiaLit, 0)
	for _, id := range r.Graph.cos {
		l, _ := r.Graph.objects[id].Fanin0(id)
		drivers = append(drivers, l)
	}
	if r.Graph.ObjectCount() != native.Objects || !slices.Equal(drivers, native.Drivers) {
		t.Fatal(drivers, native)
	}
}
