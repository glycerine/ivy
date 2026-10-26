package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaBlockProjection(t *testing.T) {
	m := NewGiaManager()
	m.Name = "block"
	m.Spec = "block.aig"
	ci := make([]GiaLit, 4)
	for i := range ci {
		ci[i], _ = m.AppendCI()
	}
	x, _ := m.AppendXorReal(ci[0], ci[2])
	buf, _ := m.AppendBuffer(ci[3].Not())
	mux, _ := m.AppendMuxReal(ci[1], x, buf)
	m.AppendCO(mux)
	m.AppendCO(ci[1])
	m.AppendCO(x)
	m.AppendCO(buf)
	m.SetRegisterCount(3)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for _, stride := range []int{1, 2, 4} {
		g, e := m.ExtractInterleavedBlock(stride, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if g.InputCount() != 4/stride || g.OutputCount() != 4/stride || g.registers != 3/stride || g.constraints != 0 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source")
		}
		for assignment := 0; assignment < 1<<g.InputCount(); assignment++ {
			input := make([]bool, g.InputCount())
			source := make([]bool, 4)
			for i := range input {
				input[i] = assignment&(1<<i) != 0
				source[i*stride] = input[i]
			}
			want, e := m.Evaluate(source)
			if e != nil {
				t.Fatal(e)
			}
			got, e := g.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			for i, value := range got {
				if value != want[i*stride] {
					t.Fatal(stride, assignment, got, want)
				}
			}
		}
	}
	for _, stride := range []int{-1, 0, 3, 5} {
		if _, e := m.ExtractInterleavedBlock(stride, GiaCofactorOptions{}); e == nil {
			t.Fatal(stride)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.ExtractInterleavedBlock(2, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 10} {
		calls := 0
		g, e := m.ExtractInterleavedBlock(2, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaInterleavedBlock(f *testing.F) {
	f.Add(uint8(2), uint8(7))
	f.Fuzz(func(t *testing.T, choice, assignment uint8) {
		m := NewGiaManager()
		ci := make([]GiaLit, 4)
		for i := range ci {
			ci[i], _ = m.AppendCI()
		}
		x, _ := m.AppendXorReal(ci[0], ci[2])
		y, _ := m.AppendAnd(ci[1], ci[3])
		m.AppendCO(x.Not())
		m.AppendCO(y)
		m.AppendCO(ci[2])
		m.AppendCO(ci[3].Not())
		m.SetRegisterCount(2)
		stride := []int{1, 2, 4}[int(choice)%3]
		g, e := m.ExtractInterleavedBlock(stride, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		input := make([]bool, 4/stride)
		v := make([]bool, 4)
		for i := range input {
			input[i] = assignment&(1<<i) != 0
			v[i*stride] = input[i]
		}
		want := []bool{v[0] == v[2], v[1] && v[3], v[2], !v[3]}
		got, e := g.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		for i, value := range got {
			if value != want[i*stride] {
				t.Fatal(stride, got, want)
			}
		}
	})
}

func TestGiaBlockNative(t *testing.T) {
	m := NewGiaManager()
	ci := make([]GiaLit, 4)
	for i := range ci {
		ci[i], _ = m.AppendCI()
	}
	x, _ := m.AppendAnd(ci[0], ci[2])
	y, _ := m.AppendAnd(ci[1], ci[3])
	m.AppendCO(x.Not())
	m.AppendCO(y)
	m.AppendCO(ci[2])
	m.AppendCO(ci[3].Not())
	m.SetRegisterCount(2)
	raw, e := os.ReadFile("test_vectors/oracle/gia_block.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects, Registers int
		Drivers            []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 3 {
		t.Fatal(e)
	}
	for i, row := range rows {
		g, e := m.ExtractInterleavedBlock(1<<i, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if g.ObjectCount() != row.Objects || g.registers != row.Registers || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(i, drivers, row)
		}
	}
}
