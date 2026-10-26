package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaOneHotAssumptionsExhaustive(t *testing.T) {
	for registers := 0; registers <= 3; registers++ {
		m := NewGiaManager()
		m.Name = "onehot"
		m.Spec = "input.aig"
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(a, b)
		buf, _ := m.AppendBuffer(x.Not())
		mux, _ := m.AppendMuxReal(c, buf, a)
		m.AppendCO(mux)
		m.AppendCO(a.Not())
		for i := 0; i < registers; i++ {
			m.AppendCO([]GiaLit{b, c, x}[i])
		}
		m.SetRegisterCount(registers)
		m.SetConstraintCount(1)
		before := slices.Clone(m.objects)
		g, e := m.WithOneHotAssumptions(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		selected := 3
		if registers > 0 {
			selected = registers
		}
		if g.registers != registers || g.constraints != 1+selected || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) || g.XorCount() != 1 || g.BufferCount() != 1 || g.MuxCount() != 1 {
			t.Fatal("metadata/source", registers)
		}
		for assignment := 0; assignment < 8; assignment++ {
			input := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
			want, e := m.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			got, e := g.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			if !slices.Equal(got[:2], want[:2]) || !slices.Equal(got[len(got)-registers:], want[len(want)-registers:]) {
				t.Fatal("outputs/registers", registers, assignment, got, want)
			}
			count := 0
			start := 0
			if registers > 0 {
				start = 3 - registers
			}
			for _, v := range input[start:] {
				if v {
					count++
				}
			}
			allowed := true
			for _, v := range got[2 : len(got)-registers] {
				allowed = allowed && v
			}
			if allowed != (count == 1) {
				t.Fatal("assumptions", registers, assignment, got)
			}
		}
	}
}
func TestGiaOneHotAssumptionsLimits(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	m.AppendCO(x)
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.WithOneHotAssumptions(options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 12} {
		calls := 0
		g, e := m.WithOneHotAssumptions(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if _, e := NewGiaManager().WithOneHotAssumptions(GiaCofactorOptions{}); e == nil {
		t.Fatal("zero inputs")
	}
}

func TestGiaOneHotAttachNativeAdaptation(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.SetRegisterCount(1)
	g, e := m.WithOneHotAssumptions(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_onehot_attach.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct{ Masks []uint }
	if e = json.Unmarshal(raw, &native); e != nil {
		t.Fatal(e)
	}
	masks := make([]uint, 3)
	for k := 0; k < 8; k++ {
		values, e := g.Evaluate([]bool{k&1 != 0, k&2 != 0, k&4 != 0})
		if e != nil {
			t.Fatal(e)
		}
		for i, v := range values {
			if v {
				masks[i] |= 1 << k
			}
		}
	}
	// Native appends violation after the RI; Go moves/inverts it before the RI.
	if masks[0] != native.Masks[0] || masks[1] != (native.Masks[2]^255) || masks[2] != native.Masks[1] {
		t.Fatal(masks, native)
	}
}
func FuzzGiaOneHotAttach(f *testing.F) {
	f.Add(uint8(5), uint8(2), uint8(42))
	f.Fuzz(func(t *testing.T, n, regs, assignment uint8) {
		size := 1 + int(n%7)
		registers := int(regs) % (size + 1)
		m := NewGiaManager()
		inputs := make([]GiaLit, size)
		for i := range inputs {
			inputs[i], _ = m.AppendCI()
		}
		m.AppendCO(inputs[0].Not())
		for i := 0; i < registers; i++ {
			m.AppendCO(inputs[i])
		}
		m.SetRegisterCount(registers)
		g, e := m.WithOneHotAssumptions(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		values := make([]bool, size)
		for i := range values {
			values[i] = assignment&(1<<i) != 0
		}
		got, e := g.Evaluate(values)
		if e != nil {
			t.Fatal(e)
		}
		start := 0
		if registers > 0 {
			start = size - registers
		}
		count := 0
		for _, v := range values[start:] {
			if v {
				count++
			}
		}
		allowed := true
		for _, v := range got[1 : len(got)-registers] {
			allowed = allowed && v
		}
		if allowed != (count == 1) || got[0] != !values[0] {
			t.Fatal(got, values)
		}
		for i := 0; i < registers; i++ {
			if got[len(got)-registers+i] != values[i] {
				t.Fatal("RI", got, values)
			}
		}
	})
}
