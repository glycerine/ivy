package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaRegisterReductionZeroInitialTraces(t *testing.T) {
	m := NewGiaManager()
	m.Name = "equiv"
	m.Spec = "equiv.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	d, _ := m.AppendCI()
	x, _ := m.AppendXorReal(b, c)
	mux, _ := m.AppendMuxReal(d, a, x)
	m.AppendCO(mux)
	m.AppendCO(b)
	m.AppendCO(GiaConst0)
	m.AppendCO(p)
	m.AppendCO(p)
	m.AppendCO(p.Not())
	m.SetRegisterCount(4)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	r, e := m.ReduceZeroInitializedRegisterInputs(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if r.Constants != 1 || r.Equivalent != 1 || r.Graph.registers != 4 || r.Graph.constraints != 1 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || !slices.Equal(before, m.objects) || !slices.Equal(r.Replacements, []GiaLit{InvalidGiaLit, GiaConst0, InvalidGiaLit, b, InvalidGiaLit}) {
		t.Fatal(r, e)
	}
	for sequence := 0; sequence < 256; sequence++ {
		sourceState, targetState := make([]bool, 4), make([]bool, 4)
		for frame := 0; frame < 8; frame++ {
			input := []bool{sequence&(1<<frame) != 0}
			want, next, e := m.Step(input, sourceState)
			if e != nil {
				t.Fatal(e)
			}
			got, nextTarget, e := r.Graph.Step(input, targetState)
			if e != nil || !slices.Equal(got, want) || !slices.Equal(nextTarget, next) {
				t.Fatal(sequence, frame, got, want, e)
			}
			sourceState, targetState = next, nextTarget
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.ReduceZeroInitializedRegisterInputs(options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 3, 7, 12} {
		calls := 0
		r, e := m.ReduceZeroInitializedRegisterInputs(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	unchanged := NewGiaManager()
	p, _ = unchanged.AppendCI()
	unchanged.AppendCO(p)
	r, e = unchanged.ReduceZeroInitializedRegisterInputs(GiaCofactorOptions{})
	if e != nil || r.Graph == unchanged || r.Constants+r.Equivalent != 0 {
		t.Fatal("independent unchanged graph", r, e)
	}
}
func FuzzGiaStructuralRegisterReduction(f *testing.F) {
	f.Add(uint8(5), uint8(42))
	f.Fuzz(func(t *testing.T, phase, sequence uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(b, c)
		m.AppendCO(x)
		m.AppendCO(a.NotIf(phase&1 != 0))
		m.AppendCO(GiaConst0)
		m.AppendCO(p.NotIf(phase&2 != 0))
		m.AppendCO(p.NotIf(phase&2 != 0))
		m.SetRegisterCount(3)
		r, e := m.ReduceZeroInitializedRegisterInputs(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		state := []bool{false, false, false}
		for frame := 0; frame < 8; frame++ {
			pv := sequence&(1<<frame) != 0
			v, e := r.Graph.Evaluate([]bool{pv, state[0], state[1], state[2]})
			if e != nil || v[0] != (state[1] != state[2]) || v[1] != (state[0] != (phase&1 != 0)) {
				t.Fatal(v, state, e)
			}
			next := pv != (phase&2 != 0)
			state = []bool{false, next, next}
			if !slices.Equal(v[2:], state) {
				t.Fatal(v, state)
			}
		}
	})
}

func TestGiaRegisterReductionNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	m.AppendCI()
	x, _ := m.AppendAnd(b, c)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.AppendCO(GiaConst0)
	m.AppendCO(p)
	m.AppendCO(p)
	m.AppendCO(p.Not())
	m.SetRegisterCount(4)
	r, e := m.ReduceZeroInitializedRegisterInputs(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_reduce_equiv.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct {
		Objects int
		Drivers []GiaLit
	}
	if e = json.Unmarshal(raw, &native); e != nil {
		t.Fatal(e)
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
