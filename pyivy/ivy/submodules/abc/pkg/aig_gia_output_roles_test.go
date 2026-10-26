package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaOutputRoleConversionProjection(t *testing.T) {
	m := NewGiaManager()
	m.Name = "roles"
	m.Spec = "roles.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	buf, _ := m.AppendBuffer(b.Not())
	mux, _ := m.AppendMuxReal(a, x, buf)
	for _, l := range []GiaLit{mux, p, GiaConst1, b.Not(), x, a.Not()} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for _, roles := range [][]int{{1, 0, 2, 1}, {2, 2, 2, 2}, {1, 1, 1, 1}, {0, 0, 0, 0}, {-7, 0, 1, 9}} {
		r, e := m.ConvertOutputRoles(roles, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		count := 0
		for _, role := range roles {
			if role == 1 {
				count++
			}
		}
		if r.Graph.constraints != count || r.Graph.registers != 2 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || r.Graph.XorCount() != 1 || r.Graph.BufferCount() != 1 || r.Graph.MuxCount() != 1 || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source")
		}
		for k := 0; k < 8; k++ {
			input := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
			want, e := m.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			got, e := r.Graph.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			for i, index := range r.Sources {
				if got[i] != (want[index] != r.Complemented[i]) {
					t.Fatal(roles, k, i, index, got, want)
				}
			}
		}
	}
	if _, e := m.ConvertOutputRoles(nil, GiaCofactorOptions{}); e == nil {
		t.Fatal("dimensions")
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.ConvertOutputRoles([]int{1, 0, 2, 1}, options)
		if r.Graph != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 9, 14} {
		calls := 0
		r, e := m.ConvertOutputRoles([]int{1, 0, 2, 1}, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaOutputRoles(f *testing.F) {
	f.Add(uint8(5), uint8(7))
	f.Fuzz(func(t *testing.T, encoding, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		for _, l := range []GiaLit{x, b, p, a.Not(), p, b.Not()} {
			m.AppendCO(l)
		}
		m.SetRegisterCount(2)
		roles := make([]int, 4)
		v := int(encoding)
		for i := range roles {
			roles[i] = v % 3
			v /= 3
		}
		r, e := m.ConvertOutputRoles(roles, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		pv, av, bv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		source := []bool{pv != av, bv, pv, !av, pv, !bv}
		expected := make([]bool, 0)
		for role := 0; role < 2; role++ {
			for i, kind := range roles {
				if kind == role {
					expected = append(expected, source[i] != (role == 1))
				}
			}
		}
		expected = append(expected, pv, !bv)
		got, e := r.Graph.Evaluate([]bool{pv, av, bv})
		if e != nil || !slices.Equal(got, expected) {
			t.Fatal(roles, got, expected, e)
		}
	})
}

func TestGiaOutputRolesNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	for _, l := range []GiaLit{x, b, p, a.Not(), p, b.Not()} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(2)
	raw, e := os.ReadFile("test_vectors/oracle/gia_output_roles.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects, Constraints int
		Drivers              []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 81 {
		t.Fatal(e)
	}
	for encoding, row := range rows {
		roles := make([]int, 4)
		value := encoding
		for i := range roles {
			roles[i] = value % 3
			value /= 3
		}
		r, e := m.ConvertOutputRoles(roles, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range r.Graph.cos {
			l, _ := r.Graph.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if r.Graph.ObjectCount() != row.Objects || r.Graph.constraints != row.Constraints || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(encoding, drivers, row)
		}
	}
}
