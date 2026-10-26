package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaDesignMiterModesAndInputMaps(t *testing.T) {
	build := func(second, sequential bool) *GiaManager {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		if sequential && second {
			m.AppendCO(a.Not())
			m.AppendCO(p)
			m.AppendCO(p.Not())
			m.SetRegisterCount(1)
		} else {
			b, _ := m.AppendCI()
			x, _ := m.AppendXorReal(p, a)
			if sequential {
				m.AppendCO(b)
				m.AppendCO(p)
				m.AppendCO(x)
				m.AppendCO(a.Not())
				m.SetRegisterCount(2)
			} else {
				m.AppendCO(x.NotIf(second))
				m.AppendCO(b)
				m.AppendCO(p.Not())
				m.SetRegisterCount(2)
			}
		}
		m.SetConstraintCount(1)
		return m
	}
	for _, sequential := range []bool{false, true} {
		first, second := build(false, sequential), build(true, sequential)
		before := slices.Clone(first.objects)
		inputCount := first.InputCount()
		if sequential {
			inputCount = first.PrimaryInputCount()
		}
		for independent := 0; independent <= inputCount; independent++ {
			for mode := GiaMiterXor; mode <= GiaMiterDual; mode++ {
				r, e := BuildGiaDesignMiter(first, second, GiaMiterOptions{Sequential: sequential, IndependentSuffix: independent, Mode: mode})
				if e != nil {
					t.Fatal(e)
				}
				g := r.Graph
				if g.Name != "miter" || g.Spec != "" || g.constraints != 0 || !slices.Equal(before, first.objects) {
					t.Fatal("metadata/source")
				}
				for assignment := 0; assignment < 1<<g.InputCount(); assignment++ {
					input := make([]bool, g.InputCount())
					for i := range input {
						input[i] = assignment&(1<<i) != 0
					}
					a, b := make([]bool, first.InputCount()), make([]bool, second.InputCount())
					for i, index := range r.FirstInputs {
						a[i] = input[index]
					}
					for i, index := range r.SecondInputs {
						b[i] = input[index]
					}
					left, e := first.Evaluate(a)
					if e != nil {
						t.Fatal(e)
					}
					right, e := second.Evaluate(b)
					if e != nil {
						t.Fatal(e)
					}
					count := first.OutputCount()
					if sequential {
						count = first.PrimaryOutputCount()
					}
					expected := make([]bool, 0)
					for i := 0; i < count; i++ {
						switch mode {
						case GiaMiterXor:
							expected = append(expected, left[i] != right[i])
						case GiaMiterImplication:
							expected = append(expected, left[i] && !right[i])
						case GiaMiterDual:
							expected = append(expected, left[i], right[i])
						}
					}
					if sequential {
						expected = append(expected, left[first.PrimaryOutputCount():]...)
						expected = append(expected, right[second.PrimaryOutputCount():]...)
					}
					got, e := g.Evaluate(input)
					if e != nil || !slices.Equal(got, expected) {
						t.Fatal(sequential, independent, mode, assignment, got, expected, e)
					}
				}
			}
		}
	}
	first, second := build(false, true), build(true, true)
	for _, options := range []GiaMiterOptions{{IndependentSuffix: -1}, {IndependentSuffix: 2, Sequential: true}, {Mode: 3}, {Sequential: true, GiaCofactorOptions: GiaCofactorOptions{MaxObjects: 1}}, {Sequential: true, GiaCofactorOptions: GiaCofactorOptions{MaxMappingEntries: 1}}} {
		r, e := BuildGiaDesignMiter(first, second, options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}

	for _, options := range []GiaMiterOptions{
		{Sequential: true, GiaCofactorOptions: GiaCofactorOptions{MaxObjects: 1}},
		{Sequential: true, GiaCofactorOptions: GiaCofactorOptions{MaxMappingEntries: 1}},
	} {
		r, e := BuildGiaDesignMiter(first, second, options)
		if r.Graph != nil || !errors.Is(e, ErrLimit) {
			t.Fatal("resource limit", e)
		}
	}
	for _, pair := range [][2]*GiaManager{{nil, first}, {first, nil}, {first, NewGiaManager()}} {
		r, e := BuildGiaDesignMiter(pair[0], pair[1], GiaMiterOptions{Sequential: true})
		if r.Graph != nil || !errors.Is(e, ErrDimension) {
			t.Fatal("invalid sources", e)
		}
	}
	alias, e := BuildGiaDesignMiter(first, first, GiaMiterOptions{})
	if e != nil {
		t.Fatal(e)
	}
	for _, id := range alias.Graph.cos {
		d, e := alias.Graph.objects[id].Fanin0(id)
		if e != nil || d != GiaConst0 {
			t.Fatal("alias comparison", d, e)
		}
	}
	for _, threshold := range []int{1, 5, 10, 18} {
		calls := 0
		r, e := BuildGiaDesignMiter(first, second, GiaMiterOptions{Sequential: true, GiaCofactorOptions: GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }}})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaDesignMiter(f *testing.F) {
	f.Add(uint8(5), uint8(13))
	f.Fuzz(func(t *testing.T, flags, assignment uint8) {
		build := func(flip bool) *GiaManager {
			m := NewGiaManager()
			p, _ := m.AppendCI()
			r, _ := m.AppendCI()
			x, _ := m.AppendXorReal(p, r)
			m.AppendCO(x.NotIf(flip))
			m.AppendCO(p.Not())
			m.SetRegisterCount(1)
			return m
		}
		seq := flags&1 != 0
		independent := int((flags >> 1) & 1)
		mode := GiaMiterMode((flags >> 2) % 3)
		r, e := BuildGiaDesignMiter(build(false), build(true), GiaMiterOptions{Sequential: seq, IndependentSuffix: independent, Mode: mode})
		if e != nil {
			t.Fatal(e)
		}
		input := make([]bool, r.Graph.InputCount())
		for i := range input {
			input[i] = assignment&(1<<i) != 0
		}
		p0, s0 := input[r.FirstInputs[0]], input[r.FirstInputs[1]]
		p1, s1 := input[r.SecondInputs[0]], input[r.SecondInputs[1]]
		left, right := []bool{p0 != s0, !p0}, []bool{p1 == s1, !p1}
		count := 2
		if seq {
			count = 1
		}
		expected := make([]bool, 0)
		for i := 0; i < count; i++ {
			if mode == GiaMiterDual {
				expected = append(expected, left[i], right[i])
			} else if mode == GiaMiterImplication {
				expected = append(expected, left[i] && !right[i])
			} else {
				expected = append(expected, left[i] != right[i])
			}
		}
		if seq {
			expected = append(expected, !p0, !p1)
		}
		got, e := r.Graph.Evaluate(input)
		if e != nil || !slices.Equal(got, expected) {
			t.Fatal(flags, got, expected, e)
		}
	})
}

func TestGiaDesignMiterNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_design_miter.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Sequential, Independent, Mode, Objects, Inputs, Registers int
		Drivers                                                   []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 12 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		build := func(flip bool) *GiaManager {
			m := NewGiaManager()
			a, _ := m.AppendCI()
			b, _ := m.AppendCI()
			x, _ := m.AppendAnd(a, b)
			m.AppendCO(x.NotIf(flip))
			m.AppendCO(a.Not())
			m.SetRegisterCount(row.Sequential)
			return m
		}
		r, e := BuildGiaDesignMiter(build(false), build(true), GiaMiterOptions{Sequential: row.Sequential != 0, IndependentSuffix: row.Independent, Mode: GiaMiterMode(row.Mode)})
		if e != nil {
			t.Fatal(e)
		}
		g := r.Graph
		var drivers []GiaLit
		for _, id := range g.cos {
			d, e := g.objects[id].Fanin0(id)
			if e != nil {
				t.Fatal(e)
			}
			drivers = append(drivers, d)
		}
		if g.ObjectCount() != row.Objects || g.InputCount() != row.Inputs || g.registers != row.Registers || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(row, g.ObjectCount(), g.InputCount(), g.registers, drivers)
		}
	}
}
