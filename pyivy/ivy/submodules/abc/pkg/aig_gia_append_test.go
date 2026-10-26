package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func appendSource(flip, extended bool) *GiaManager {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	var x GiaLit
	if extended {
		x, _ = m.AppendXorReal(a, b)
	} else {
		x, _ = m.AppendAnd(a, b)
	}
	m.AppendCO(x.NotIf(flip))
	m.AppendCO(a.Not())
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	m.Name = "base"
	m.Spec = "base.aig"
	return m
}
func TestGiaAppendNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_append.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Shared, Objects, Registers, Constraints int
		CiIds                                   []GiaID
		Drivers                                 []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 2 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m, s := appendSource(false, false), appendSource(true, false)
		g, _, e := m.WithAppendedGraph(s, row.Shared != 0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		var drivers []GiaLit
		for _, id := range g.cos {
			d, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, d)
		}
		if g.ObjectCount() != row.Objects || g.registers != row.Registers || g.constraints != row.Constraints || !slices.Equal(g.cis, row.CiIds) || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(row, g.ObjectCount(), drivers)
		}
	}
}
func TestGiaAppendOwnershipAndRoles(t *testing.T) {
	for _, shared := range []bool{false, true} {
		for _, active := range []bool{false, true} {
			m := appendSource(false, true)
			if active {
				m.HashStart()
			}
			before := slices.Clone(m.objects)
			g, mapping, e := m.WithAppendedGraph(m, shared, GiaCofactorOptions{})
			if e != nil {
				t.Fatal(e)
			}
			regs := 0
			if shared {
				regs = 1
			}
			if g.registers != regs || g.constraints != m.constraints || g.Name != m.Name || g.Spec != m.Spec || g.hash == nil || !slices.Equal(before, m.objects) || (m.hash != nil) != active {
				t.Fatal("metadata")
			}
			for i, id := range m.cos {
				if mapping[id].ID() != g.cos[len(m.cos)+i] {
					t.Fatal("CO map")
				}
			}
			for bits := 0; bits < 1<<g.InputCount(); bits++ {
				input := make([]bool, g.InputCount())
				for i := range input {
					input[i] = bits&(1<<i) != 0
				}
				left, _ := m.Evaluate(input[:2])
				rightInputs := input[:2]
				if !shared {
					rightInputs = input[2:]
				}
				right, _ := m.Evaluate(rightInputs)
				want := append(slices.Clone(left), right...)
				got, e := g.Evaluate(input)
				if e != nil || !slices.Equal(got, want) {
					t.Fatal(bits, got, want, e)
				}
			}
			g.AppendCO(GiaConst0)
			if !slices.Equal(before, m.objects) {
				t.Fatal("alias")
			}
		}
	}
	m := appendSource(false, false)
	for _, o := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
		g, mp, e := m.WithAppendedGraph(m, true, o)
		if g != nil || mp != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, o := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, _, e := m.WithAppendedGraph(m, true, o)
		if g != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, pair := range [][2]*GiaManager{{nil, m}, {m, nil}, {m, NewGiaManager()}} {
		g, _, e := pair[0].WithAppendedGraph(pair[1], true, GiaCofactorOptions{})
		if g != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, threshold := range []int{1, 3, 6, 8} {
		calls := 0
		g, mp, e := m.WithAppendedGraph(m, false, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || mp != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaAppend(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 48 {
			data = data[:48]
		}
		build := func(reverse bool) *GiaManager {
			m := NewGiaManager()
			pool := []GiaLit{GiaConst0}
			for i := 0; i < 2; i++ {
				v, _ := m.AppendCI()
				pool = append(pool, v)
			}
			m.HashStart()
			for i, v := range data {
				a := pool[int(v)%len(pool)].NotIf(v&8 != 0)
				b := pool[(int(v)+i)%len(pool)]
				var x GiaLit
				var e error
				if v&4 != 0 {
					x, e = m.AppendBuffer(a)
				} else if v&2 != 0 && a.ID() != b.ID() {
					x, e = m.AppendXorReal(a, b)
				} else {
					x, e = m.HashAnd(a, b)
				}
				if e != nil {
					t.Fatal(e)
				}
				pool = append(pool, x)
			}
			m.HashStop()
			m.AppendCO(pool[len(pool)-1].NotIf(reverse))
			m.AppendCO(pool[1].NotIf(reverse))
			m.SetRegisterCount(1)
			return m
		}
		m, s := build(false), build(true)
		shared := data[0]&1 != 0
		g, mapping, e := m.WithAppendedGraph(s, shared, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for bits := 0; bits < 1<<g.InputCount(); bits++ {
			input := make([]bool, g.InputCount())
			for i := range input {
				input[i] = bits&(1<<i) != 0
			}
			left, _ := m.Evaluate(input[:2])
			rightInputs := input[:2]
			if !shared {
				rightInputs = input[2:]
			}
			right, _ := s.Evaluate(rightInputs)
			want := append(slices.Clone(left), right...)
			got, e := g.Evaluate(input)
			if e != nil || !slices.Equal(got, want) {
				t.Fatal(data, bits, got, want, e)
			}
		}
		for i, id := range s.cos {
			if mapping[id].ID() != g.cos[len(m.cos)+i] {
				t.Fatal("map")
			}
		}
	})
}
