package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaFlopClassProjection(t *testing.T) {
	m := NewGiaManager()
	m.Name = "class"
	m.Spec = "class.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	mux, _ := m.AppendMuxReal(b, x, c)
	buf, _ := m.AppendBuffer(mux.Not())
	m.AppendCO(buf)
	m.AppendCO(p)
	m.AppendCO(b)
	m.AppendCO(c.Not())
	m.AppendCO(x)
	m.SetRegisterCount(3)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	classes := []int{1, 2, 1}
	for _, selected := range []int{1, 2, 9} {
		r, e := m.ExtractFlopClass(classes, selected, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		expected := []int{}
		for i, class := range classes {
			if class == selected {
				expected = append(expected, i)
			}
		}
		if !slices.Equal(r.Registers, expected) || r.Graph.registers != len(expected) || r.Graph.constraints != 1 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || r.Graph.MuxCount() != 1 || r.Graph.XorCount() != 1 || r.Graph.BufferCount() != 1 || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/projection", r)
		}
		for assignment := 0; assignment < 16; assignment++ {
			source := make([]bool, 4)
			for i := range source {
				source[i] = assignment&(1<<i) != 0
			}
			want, e := m.Evaluate(source)
			if e != nil {
				t.Fatal(e)
			}
			input := make([]bool, len(r.Inputs))
			for i, index := range r.Inputs {
				input[i] = source[index]
			}
			got, e := r.Graph.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			for i, index := range r.Outputs {
				if got[i] != want[index] {
					t.Fatal(selected, assignment, i, index)
				}
			}
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.ExtractFlopClass(classes, 1, options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 9, 13} {
		calls := 0
		r, e := m.ExtractFlopClass(classes, 1, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if _, e := m.ExtractFlopClass(nil, 1, GiaCofactorOptions{}); e == nil {
		t.Fatal("classes")
	}
}
func FuzzGiaFlopClass(f *testing.F) {
	f.Add(uint8(5), uint8(11))
	f.Fuzz(func(t *testing.T, classes, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		m.AppendCO(b.Not())
		m.AppendCO(x)
		m.AppendCO(a.Not())
		m.SetRegisterCount(2)
		labels := []int{int(classes & 1), int((classes >> 1) & 1)}
		selected := int((classes >> 2) & 1)
		r, e := m.ExtractFlopClass(labels, selected, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		source := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		input := make([]bool, len(r.Inputs))
		for i, index := range r.Inputs {
			input[i] = source[index]
		}
		got, e := r.Graph.Evaluate(input)
		want := []bool{!source[2], source[0] != source[1], !source[1]}
		if e != nil {
			t.Fatal(e)
		}
		for i, index := range r.Outputs {
			if got[i] != want[index] {
				t.Fatal(got, want)
			}
		}
	})
}

func TestGiaFlopClassNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(p, c)
	m.AppendCO(x)
	m.AppendCO(y)
	m.AppendCO(a.Not())
	m.AppendCO(b)
	m.SetRegisterCount(3)
	raw, e := os.ReadFile("test_vectors/oracle/gia_flop_class.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Registers int
		Drivers   []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 3 {
		t.Fatal(e)
	}
	for i, row := range rows {
		r, e := m.ExtractFlopClass([]int{1, 2, 1}, i+1, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range r.Graph.cos {
			l, _ := r.Graph.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if r.Graph.registers != row.Registers || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(i, drivers, row)
		}
	}
}
