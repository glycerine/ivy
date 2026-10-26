package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaRegisterGapsProjection(t *testing.T) {
	m := NewGiaManager()
	m.Name = "gaps"
	m.Spec = "gaps.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	buf, _ := m.AppendBuffer(b.Not())
	mux, _ := m.AppendMuxReal(a, x, buf)
	m.AppendCO(mux)
	m.AppendCO(p)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	mask := []int{1, -1, 0, -1}
	for _, permute := range []bool{false, true} {
		g, e := m.spreadRegisters(mask, permute, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if g.registers != 4 || g.constraints != 1 || g.Name != m.Name || g.Spec != m.Spec || g.XorCount() != 1 || g.MuxCount() != 1 || g.BufferCount() != 1 || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source")
		}
		for k := 0; k < 32; k++ {
			input := make([]bool, 5)
			for i := range input {
				input[i] = k&(1<<i) != 0
			}
			a, b := input[1], input[3]
			if permute {
				a, b = b, a
			}
			want, e := m.Evaluate([]bool{input[0], a, b})
			if e != nil {
				t.Fatal(e)
			}
			got, e := g.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			first, second := want[2], want[3]
			if permute {
				first, second = second, first
			}
			expected := []bool{want[0], want[1], first, false, second, false}
			if !slices.Equal(got, expected) {
				t.Fatal(permute, k, got, expected)
			}
		}
	}
	for _, mask := range [][]int{{}, {0}, {0, 0}, {0, 2}, {0, -1, -1}} {
		if _, e := m.PermuteRegistersWithGaps(mask, GiaCofactorOptions{}); e == nil {
			t.Fatal(mask)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.SpreadRegisters(mask, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 9, 14} {
		calls := 0
		g, e := m.SpreadRegisters(mask, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaRegisterGaps(f *testing.F) {
	f.Add(uint8(5), uint8(42))
	f.Fuzz(func(t *testing.T, choice, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		m.AppendCO(b.Not())
		m.AppendCO(x)
		m.AppendCO(a.Not())
		m.SetRegisterCount(2)
		masks := [][]int{{0, 1}, {1, -1, 0}, {-1, 0, 1, -1}, {-1, 1, -1, 0}}
		mask := masks[int(choice)%len(masks)]
		g, e := m.PermuteRegistersWithGaps(mask, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		input := make([]bool, 1+len(mask))
		for i := range input {
			input[i] = assignment&(1<<i) != 0
		}
		source := []bool{input[0], false, false}
		for i, index := range mask {
			if index >= 0 {
				source[1+index] = input[1+i]
			}
		}
		expected := []bool{!source[2]}
		next := []bool{source[0] != source[1], !source[1]}
		for _, index := range mask {
			v := false
			if index >= 0 {
				v = next[index]
			}
			expected = append(expected, v)
		}
		got, e := g.Evaluate(input)
		if e != nil || !slices.Equal(got, expected) {
			t.Fatal(got, expected, e)
		}
	})
}

func TestGiaRegisterGapsNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	m.AppendCO(b)
	m.AppendCO(x)
	m.AppendCO(a.Not())
	m.SetRegisterCount(2)
	raw, e := os.ReadFile("test_vectors/oracle/gia_spread_registers.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows [][]GiaLit
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 2 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		g, e := m.spreadRegisters([]int{1, -1, 0, -1}, mode != 0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if !slices.Equal(drivers, row) {
			t.Fatal(mode, drivers, row)
		}
	}
}
