package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func supportClassSource(empty bool) *GiaManager {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	d, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(c, d)
	z, _ := m.AppendAnd(a, b.Not())
	w, _ := m.AppendAnd(x, c)
	if !empty {
		for _, l := range []GiaLit{x, y, z, x.Not(), a, c.Not(), GiaConst0, GiaConst1, GiaConst0, w} {
			m.AppendCO(l)
		}
	}
	return m
}
func TestGiaSupportClassesNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_support_classes.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Empty, Objects int
		Classes        [][]int
		Drivers        []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 2 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m := supportClassSource(row.Empty != 0)
		r, e := m.ClassifyOutputsBySupport(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		g := r.Graph
		var drivers []GiaLit
		for _, id := range g.cos {
			d, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, d)
		}
		if g.ObjectCount() != row.Objects || !slices.EqualFunc(r.Classes, row.Classes, slices.Equal[[]int]) || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(row, g.ObjectCount(), r.Classes, drivers)
		}
	}
}
func TestGiaSupportClassesValidation(t *testing.T) {
	m := supportClassSource(false)
	for _, o := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
		r, e := m.ClassifyOutputsBySupport(o)
		if r.Graph != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, o := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.ClassifyOutputsBySupport(o)
		if r.Graph != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, threshold := range []int{1, 5, 15, 30} {
		calls := 0
		r, e := m.ClassifyOutputsBySupport(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if r, e := (*GiaManager)(nil).ClassifyOutputsBySupport(GiaCofactorOptions{}); r.Graph != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	m.SetRegisterCount(1)
	if r, e := m.ClassifyOutputsBySupport(GiaCofactorOptions{}); r.Graph != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzGiaSupportClasses(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add([]byte{128, 9, 7, 6})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 48 {
			data = data[:48]
		}
		m := NewGiaManager()
		pool := []GiaLit{GiaConst0}
		for i := 0; i < 3; i++ {
			v, _ := m.AppendCI()
			pool = append(pool, v)
		}
		m.HashStart()
		for i, v := range data {
			a := pool[int(v)%len(pool)].NotIf(v&16 != 0)
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
		n := int(data[0] % 9)
		for i := 0; i < n; i++ {
			v := data[i%len(data)]
			m.AppendCO(pool[int(v)%len(pool)].NotIf(v&64 != 0))
		}
		if n > 0 {
			m.SetConstraintCount(1)
		}
		m.Name = "classes"
		m.Spec = "classes.aig"
		before := slices.Clone(m.objects)
		r, e := m.ClassifyOutputsBySupport(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if r.Graph.InputCount() != m.InputCount() || r.Graph.OutputCount() != len(r.Classes) || r.Graph.constraints != 0 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || !slices.Equal(before, m.objects) || len(r.Outputs) != n {
			t.Fatal("metadata")
		}
		covered := make([]bool, n)
		last := -1
		for group, items := range r.Classes {
			if len(items) == 0 || items[0] <= last {
				t.Fatal("class order")
			}
			last = items[0]
			for i, co := range items {
				if co < 0 || co >= n || covered[co] || r.Outputs[co] != group || (i > 0 && co <= items[i-1]) {
					t.Fatal("membership")
				}
				covered[co] = true
			}
		}
		for _, v := range covered {
			if !v {
				t.Fatal("missing")
			}
		}
		for bits := 0; bits < 8; bits++ {
			input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			source, e := m.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			got, e := r.Graph.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			for group, items := range r.Classes {
				if got[group] != source[items[0]] {
					t.Fatal("representative")
				}
				var value bool
				for k, co := range items {
					projected := make([]bool, 3)
					for i, ci := range r.Supports[co] {
						projected[ci] = input[i]
					}
					out, e := m.Evaluate(projected)
					if e != nil {
						t.Fatal(e)
					}
					if k == 0 {
						value = out[co]
					} else if value != out[co] {
						t.Fatal("invalid normalized class", data, items, bits)
					}
				}
			}
		}
	})
}
