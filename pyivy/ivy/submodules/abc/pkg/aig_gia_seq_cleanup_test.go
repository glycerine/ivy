package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaSequentialCleanupClosure(t *testing.T) {
	m := NewGiaManager()
	m.Name = "closure"
	m.Spec = "closure.aig"
	p, _ := m.AppendCI()
	m.AppendCI()
	a, _ := m.AppendCI()
	unused, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, c)
	buf, _ := m.AppendBuffer(a.Not())
	m.AppendAnd(p, unused)
	m.AppendCO(x)
	m.AppendCO(c)
	m.AppendCO(p)
	m.AppendCO(unused)
	m.AppendCO(buf)
	m.SetRegisterCount(3)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	r, e := m.SequentialCleanup(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if !slices.Equal(r.Registers, []int{0, 2}) || !slices.Equal(r.PrimaryInputs, []int{0, 1}) || r.Graph.registers != 2 || r.Graph.constraints != 1 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || r.Graph.XorCount() != 1 || r.Graph.BufferCount() != 1 || !slices.Equal(before, m.objects) {
		t.Fatal("metadata/closure", r)
	}
	for k := 0; k < 32; k++ {
		source := make([]bool, 5)
		for i := range source {
			source[i] = k&(1<<i) != 0
		}
		want, e := m.Evaluate(source)
		if e != nil {
			t.Fatal(e)
		}
		got, e := r.Graph.Evaluate([]bool{source[0], source[1], source[2], source[4]})
		if e != nil || !slices.Equal(got, []bool{want[0], want[1], want[2], want[4]}) {
			t.Fatal(k, got, want, e)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.SequentialCleanup(options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 3, 6, 10} {
		calls := 0
		r, e := m.SequentialCleanup(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	noPO := NewGiaManager()
	p, _ = noPO.AppendCI()
	a, _ = noPO.AppendCI()
	noPO.AppendCO(a)
	noPO.SetRegisterCount(1)
	r, e = noPO.SequentialCleanup(GiaCofactorOptions{})
	if e != nil || r.Graph.registers != 0 || r.Graph.OutputCount() != 0 || r.Graph.InputCount() != 1 {
		t.Fatal("unobservable cycle", r, e)
	}
}
func FuzzGiaSequentialCleanup(f *testing.F) {
	f.Add(uint8(5), uint8(2), uint8(13))
	f.Fuzz(func(t *testing.T, links, selected, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		state := make([]GiaLit, 3)
		for i := range state {
			state[i], _ = m.AppendCI()
		}
		root := int(selected % 3)
		m.AppendCO(state[root])
		for i := range state {
			driver := p
			if links&(1<<i) != 0 {
				driver = state[(i+1)%3].Not()
			}
			m.AppendCO(driver)
		}
		m.SetRegisterCount(3)
		r, e := m.SequentialCleanup(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		used := uint8(1 << root)
		for round := 0; round < 3; round++ {
			for i := 0; i < 3; i++ {
				if used&(1<<i) != 0 && links&(1<<i) != 0 {
					used |= 1 << ((i + 1) % 3)
				}
			}
		}
		expected := make([]int, 0)
		for i := 0; i < 3; i++ {
			if used&(1<<i) != 0 {
				expected = append(expected, i)
			}
		}
		if !slices.Equal(r.Registers, expected) {
			t.Fatal(r.Registers, expected)
		}
		values := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0, assignment&8 != 0}
		input := []bool{values[0]}
		for _, index := range expected {
			input = append(input, values[1+index])
		}
		got, e := r.Graph.Evaluate(input)
		if e != nil || got[0] != values[1+root] {
			t.Fatal(got, e)
		}
		for i, index := range expected {
			want := values[0]
			if links&(1<<index) != 0 {
				want = !values[1+(index+1)%3]
			}
			if got[i+1] != want {
				t.Fatal(got, index, want)
			}
		}
	})
}

func TestGiaSequentialCleanupNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	m.AppendCI()
	a, _ := m.AppendCI()
	unused, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, c)
	m.AppendAnd(a, unused)
	m.AppendCO(x)
	m.AppendCO(p)
	m.AppendCO(unused)
	m.AppendCO(a.Not())
	m.SetRegisterCount(3)
	r, e := m.SequentialCleanup(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_seq_cleanup.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct {
		Objects, Registers int
		Drivers            []GiaLit
	}
	if e = json.Unmarshal(raw, &native); e != nil {
		t.Fatal(e)
	}
	drivers := make([]GiaLit, 0)
	for _, id := range r.Graph.cos {
		l, _ := r.Graph.objects[id].Fanin0(id)
		drivers = append(drivers, l)
	}
	if r.Graph.ObjectCount() != native.Objects || r.Graph.registers != native.Registers || !slices.Equal(drivers, native.Drivers) {
		t.Fatal(drivers, native)
	}
}
