package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaSliceNative(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(x, c)
	z, _ := m.AppendAnd(b, c)
	m.AppendAnd(a, c)
	m.AppendCO(y)
	m.AppendCO(z.Not())
	m.AppendCO(a)
	raw, e := os.ReadFile("test_vectors/oracle/gia_slice.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects int
		Drivers []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 5 {
		t.Fatal(e)
	}
	for limit, row := range rows {
		r, e := m.SliceBySupport(limit, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range r.Graph.cos {
			l, _ := r.Graph.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if r.Graph.ObjectCount() != row.Objects || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(limit, drivers, row)
		}
		expected := []GiaID{}
		if limit == 2 {
			expected = []GiaID{x.ID(), z.ID()}
		} else if limit >= 3 {
			expected = []GiaID{y.ID(), z.ID()}
		}
		if !slices.Equal(r.Roots, expected) {
			t.Fatal(limit, r.Roots, expected)
		}
		for assignment := 0; assignment < 8; assignment++ {
			av, bv, cv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
			truth := map[GiaID]bool{x.ID(): av && bv, y.ID(): av && bv && cv, z.ID(): bv && cv}
			got, e := r.Graph.Evaluate([]bool{av, bv, cv})
			if e != nil {
				t.Fatal(e)
			}
			for i, id := range r.Roots {
				if got[i] != truth[id] {
					t.Fatal(limit, assignment, id)
				}
			}
		}
	}
}
func TestGiaSliceExtendedAndLimits(t *testing.T) {
	m := NewGiaManager()
	m.Name = "slice"
	m.Spec = "slice.aig"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	buf, _ := m.AppendBuffer(c.Not())
	mux, _ := m.AppendMuxReal(buf, x, a)
	m.AppendCO(mux)
	m.AppendCO(a)
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	r, e := m.SliceBySupport(2, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if !slices.Equal(r.Roots, []GiaID{x.ID(), buf.ID()}) || r.Graph.registers != 0 || r.Graph.constraints != 0 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || !slices.Equal(before, m.objects) {
		t.Fatal(r, e)
	}
	for assignment := 0; assignment < 8; assignment++ {
		a, b, c := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		v, e := r.Graph.Evaluate([]bool{a, b, c})
		if e != nil || !slices.Equal(v, []bool{a != b, !c}) {
			t.Fatal(v, e)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.SliceBySupport(2, options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 3, 6, 9} {
		calls := 0
		r, e := m.SliceBySupport(2, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if _, e := m.SliceBySupport(-1, GiaCofactorOptions{}); e == nil {
		t.Fatal("negative")
	}
}
