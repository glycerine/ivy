package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaDecomposeNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_decompose.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Sequential, Flip, Objects int
		Drivers                   []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 4 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendAnd(a, b)
		y, _ := m.AppendAnd(x.Not(), c)
		m.AppendCO(y.NotIf(row.Flip != 0))
		if row.Sequential != 0 {
			m.AppendCO(a.Not())
		}
		m.SetRegisterCount(row.Sequential)
		r, e := m.DecomposeOutput(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		g := r.Graph
		var drivers []GiaLit
		for _, id := range g.cos {
			d, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, d)
		}
		if g.ObjectCount() != row.Objects || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(row, g.ObjectCount(), drivers)
		}
	}
}
func TestGiaDecomposeValidation(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	m.AppendCO(x)
	for _, o := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
		r, e := m.DecomposeOutput(o)
		if r.Graph != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, o := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.DecomposeOutput(o)
		if r.Graph != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 12} {
		calls := 0
		r, e := m.DecomposeOutput(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	for _, source := range []*GiaManager{nil, NewGiaManager()} {
		r, e := source.DecomposeOutput(GiaCofactorOptions{})
		if r.Graph != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	m.PatchCODriver(0, a)
	if r, e := m.DecomposeOutput(GiaCofactorOptions{}); r.Graph != nil || !errors.Is(e, ErrUnsupported) {
		t.Fatal(e)
	}
}
func FuzzGiaDecompose(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{128, 5, 6, 7, 8, 9})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 48 {
			data = data[:48]
		}
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		pool := []GiaLit{a, b, c}
		for i, v := range data {
			left := pool[int(v)%len(pool)]
			right := pool[(int(v)+i+1)%len(pool)]
			if left.ID() == right.ID() {
				continue
			}
			var x GiaLit
			var e error
			if v&1 != 0 {
				x, e = m.AppendXorReal(left, right)
			} else {
				x, e = m.AppendAnd(left.NotIf(v&2 != 0), right.NotIf(v&4 != 0))
			}
			if e != nil {
				t.Fatal(e)
			}
			pool = append(pool, x)
		}
		root := a
		for _, v := range data {
			l := pool[int(v)%len(pool)].NotIf(v&16 != 0)
			if root.ID() == l.ID() {
				continue
			}
			var e error
			root, e = m.AppendAnd(root, l)
			if e != nil {
				t.Fatal(e)
			}
		}
		if root.ID() == a.ID() {
			root, _ = m.AppendAnd(a, b)
		}
		flip := data[0]&128 != 0
		m.AppendCO(root.NotIf(flip))
		seq := data[0]&64 != 0
		if seq {
			m.AppendCO(c.Not())
			m.SetRegisterCount(1)
		}
		m.SetConstraintCount(1)
		m.Name = "decomp"
		m.Spec = "decomp.aig"
		before := slices.Clone(m.objects)
		r, e := m.DecomposeOutput(GiaCofactorOptions{})
		if errors.Is(e, ErrUnsupported) {
			return
		}
		if e != nil {
			t.Fatal(e)
		}
		if r.OR != flip || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || r.Graph.constraints != 0 || r.Graph.registers != m.registers || !slices.Equal(before, m.objects) {
			t.Fatal("metadata")
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
			value := !r.OR
			for _, v := range got[:len(r.Parts)] {
				if r.OR {
					value = value || v
				} else {
					value = value && v
				}
			}
			if value != source[0] || !slices.Equal(got[len(r.Parts):], source[1:]) {
				t.Fatal(data, bits, got, source)
			}
		}
	})
}
