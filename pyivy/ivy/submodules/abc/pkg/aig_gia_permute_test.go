package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaPermutationsProjection(t *testing.T) {
	m := NewGiaManager()
	m.Name = "permute"
	m.Spec = "permute.aig"
	ci := make([]GiaLit, 6)
	for i := range ci {
		ci[i], _ = m.AppendCI()
	}
	x, _ := m.AppendXorReal(ci[0], ci[3])
	buf, _ := m.AppendBuffer(ci[4].Not())
	mux, _ := m.AppendMuxReal(ci[2], x, buf)
	m.AppendCO(mux)
	m.AppendCO(ci[1])
	m.AppendCO(x)
	m.AppendCO(ci[5])
	m.AppendCO(ci[3].Not())
	m.SetRegisterCount(3)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	orders := [][]int{{0, 1, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}, {1, 0, 2}, {0, 2, 1}}
	for _, registers := range []bool{false, true} {
		for _, order := range orders {
			g, e := m.permuteTerminals(order, registers, GiaCofactorOptions{})
			if e != nil {
				t.Fatal(e)
			}
			if g.constraints != 1 || g.registers != 3 || g.Name != m.Name || g.Spec != m.Spec || g.MuxCount() != 1 || g.XorCount() != 1 || g.BufferCount() != 1 || !slices.Equal(before, m.objects) {
				t.Fatal("metadata/source")
			}
			for assignment := 0; assignment < 64; assignment++ {
				source := make([]bool, 6)
				for i := range source {
					source[i] = assignment&(1<<i) != 0
				}
				target := slices.Clone(source)
				offset := 0
				if registers {
					offset = 3
				}
				for i, index := range order {
					target[offset+i] = source[offset+index]
				}
				want, e := m.Evaluate(source)
				if e != nil {
					t.Fatal(e)
				}
				got, e := g.Evaluate(target)
				if e != nil {
					t.Fatal(e)
				}
				for i, v := range got {
					sourceIndex := i
					if registers && i >= 2 {
						sourceIndex = 2 + order[i-2]
					}
					if v != want[sourceIndex] {
						t.Fatal(registers, order, assignment, i)
					}
				}
			}
		}
	}
	for _, order := range [][]int{{}, {0, 0, 2}, {-1, 1, 2}, {0, 1, 3}} {
		if _, e := m.PermuteRegisters(order, GiaCofactorOptions{}); e == nil {
			t.Fatal(order)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.PermutePrimaryInputs(orders[0], options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 9, 15} {
		calls := 0
		g, e := m.PermuteRegisters(orders[1], GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaPermutations(f *testing.F) {
	f.Add(uint8(2), uint8(41))
	f.Fuzz(func(t *testing.T, choice, assignment uint8) {
		m := NewGiaManager()
		ci := make([]GiaLit, 6)
		for i := range ci {
			ci[i], _ = m.AppendCI()
		}
		x, _ := m.AppendXorReal(ci[0], ci[3])
		m.AppendCO(ci[2])
		m.AppendCO(x)
		m.AppendCO(ci[4].Not())
		m.AppendCO(ci[5])
		m.SetRegisterCount(3)
		order := [][]int{{0, 1, 2}, {1, 2, 0}, {2, 0, 1}, {2, 1, 0}, {1, 0, 2}, {0, 2, 1}}[int(choice)%6]
		registers := choice&8 != 0
		g, e := m.permuteTerminals(order, registers, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		source := make([]bool, 6)
		for i := range source {
			source[i] = assignment&(1<<i) != 0
		}
		input := slices.Clone(source)
		offset := 0
		if registers {
			offset = 3
		}
		for i, index := range order {
			input[offset+i] = source[offset+index]
		}
		want := []bool{source[2], source[0] != source[3], !source[4], source[5]}
		got, e := g.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		for i, v := range got {
			index := i
			if registers && i > 0 {
				index = 1 + order[i-1]
			}
			if v != want[index] {
				t.Fatal(got, want, order)
			}
		}
	})
}

func TestGiaPermutationsNative(t *testing.T) {
	m := NewGiaManager()
	ci := make([]GiaLit, 6)
	for i := range ci {
		ci[i], _ = m.AppendCI()
	}
	x, _ := m.AppendAnd(ci[0], ci[3])
	m.AppendCO(ci[2])
	m.AppendCO(x)
	m.AppendCO(ci[4].Not())
	m.AppendCO(ci[5])
	m.SetRegisterCount(3)
	raw, e := os.ReadFile("test_vectors/oracle/gia_permute.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows [][]GiaLit
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 2 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		g, e := m.permuteTerminals([]int{1, 2, 0}, mode != 0, GiaCofactorOptions{})
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
