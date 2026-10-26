package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaRegisterAbstractionProjection(t *testing.T) {
	m := NewGiaManager()
	m.Name = "abstract"
	m.Spec = "abstract.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	buf, _ := m.AppendBuffer(b.Not())
	mux, _ := m.AppendMuxReal(c, x, buf)
	m.AppendAnd(a, c)
	m.AppendCO(mux)
	m.AppendCO(p)
	m.AppendCO(x)
	m.AppendCO(buf)
	m.AppendCO(c.Not())
	m.SetRegisterCount(3)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for _, external := range [][]int{nil, {2, 0}, {0, 1, 2}, {1}} {
		r, e := m.AbstractRegisters(external, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if r.Graph.registers != 3-len(external) || r.Graph.constraints != 1 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || r.Graph.ObjectCount() >= m.ObjectCount() || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source", r)
		}
		if r.Graph.PrimaryInputCount() != 1+len(external) {
			t.Fatal("pseudo inputs")
		}
		for assignment := 0; assignment < 16; assignment++ {
			source := make([]bool, 4)
			for i := range source {
				source[i] = assignment&(1<<i) != 0
			}
			input := make([]bool, 4)
			for i, index := range r.Inputs {
				input[i] = source[index]
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
					t.Fatal(external, assignment, i, index)
				}
			}
		}
	}
	for _, external := range [][]int{{-1}, {3}, {0, 0}, {0, 1, 2, 3}} {
		if _, e := m.AbstractRegisters(external, GiaCofactorOptions{}); e == nil {
			t.Fatal(external)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.AbstractRegisters([]int{0}, options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 9, 12} {
		calls := 0
		r, e := m.AbstractRegisters([]int{0}, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaRegisterAbstraction(f *testing.F) {
	f.Add(uint8(2), uint8(13))
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
		external := [][]int{nil, {0}, {1}, {1, 0}}[int(choice)%4]
		r, e := m.AbstractRegisters(external, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		source := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		input := make([]bool, 3)
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

func TestGiaRegisterAbstractionNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	y, _ := m.AppendAnd(b, c)
	m.AppendAnd(p, c)
	m.AppendCO(x)
	m.AppendCO(y)
	m.AppendCO(a.Not())
	m.AppendCO(b)
	m.SetRegisterCount(3)
	raw, e := os.ReadFile("test_vectors/oracle/gia_abstract_registers.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects, Registers int
		Drivers            []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 2 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		var external []int
		if mode != 0 {
			external = []int{2, 0}
		}
		r, e := m.AbstractRegisters(external, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range r.Graph.cos {
			l, _ := r.Graph.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if r.Graph.ObjectCount() != row.Objects || r.Graph.registers != row.Registers || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(mode, drivers, row)
		}
	}
}
