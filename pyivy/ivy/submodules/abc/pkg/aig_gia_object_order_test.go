package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func objectOrderSource(extended bool) *GiaManager {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	var x GiaLit
	if extended {
		x, _ = m.AppendXorReal(a, b)
	} else {
		x, _ = m.AppendAnd(a, b)
	}
	m.AppendCO(x)
	c, _ := m.AppendCI()
	var y GiaLit
	if extended {
		y, _ = m.AppendMuxReal(c, x, a.Not())
	} else {
		y, _ = m.AppendAnd(x.Not(), c)
	}
	m.AppendAnd(a, c)
	m.AppendCO(y.Not())
	m.AppendCO(b)
	m.SetRegisterCount(1)
	m.SetConstraintCount(1)
	m.Name = "order"
	m.Spec = "order.aig"
	return m
}
func TestGiaObjectOrderNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_object_order.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Zero, Objects, Constraints int
		CiIds, CoIds               []GiaID
		Drivers                    []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 2 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m := objectOrderSource(false)
		var g *GiaManager
		if row.Zero != 0 {
			g, e = m.ZeroOutputCopy(GiaCofactorOptions{})
		} else {
			g, _, e = m.NormalizeObjectOrder(GiaCofactorOptions{})
		}
		if e != nil {
			t.Fatal(e)
		}
		var drivers []GiaLit
		for _, id := range g.cos {
			d, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, d)
		}
		if g.ObjectCount() != row.Objects || g.constraints != row.Constraints || !slices.Equal(g.cis, row.CiIds) || !slices.Equal(g.cos, row.CoIds) || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(row, g.ObjectCount(), g.cis, g.cos, drivers)
		}
	}
}
func TestGiaObjectOrderSemantics(t *testing.T) {
	m := objectOrderSource(true)
	before := slices.Clone(m.objects)
	g, mapping, e := m.NormalizeObjectOrder(GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if g.ObjectCount() != m.ObjectCount() || g.registers != m.registers || g.constraints != 0 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
		t.Fatal("metadata")
	}
	for i, id := range m.cis {
		if mapping[id].ID() != g.cis[i] {
			t.Fatal("CI map")
		}
	}
	for i, id := range m.cos {
		if mapping[id].ID() != g.cos[i] {
			t.Fatal("CO map")
		}
	}
	for i, o := range m.objects {
		if o.IsAnd() {
			target := g.objects[mapping[i].ID()]
			if o.IsXor() != target.IsXor() || o.IsBuffer() != target.IsBuffer() {
				t.Fatal("gate kind")
			}
			_, mux := m.MuxControl(GiaID(i))
			_, targetMux := g.MuxControl(mapping[i].ID())
			if mux != targetMux {
				t.Fatal("mux kind")
			}
		}
	}
	for bits := 0; bits < 8; bits++ {
		input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
		want, _ := m.Evaluate(input)
		got, e := g.Evaluate(input)
		if e != nil || !slices.Equal(want, got) {
			t.Fatal(want, got, e)
		}
	}
	h, map2, e := g.NormalizeObjectOrder(GiaCofactorOptions{})
	if e != nil || !slices.Equal(h.objects, g.objects) {
		t.Fatal("idempotent", e)
	}
	for i, l := range map2 {
		if l != GiaLit(i<<1) {
			t.Fatal(i, l)
		}
	}
	zero, e := m.ZeroOutputCopy(GiaCofactorOptions{})
	if e != nil || zero.AndCount() != 0 || zero.Spec != "" || zero.Name != m.Name || zero.registers != m.registers || zero.constraints != 0 {
		t.Fatal(e)
	}
	for bits := 0; bits < 8; bits++ {
		values, e := zero.Evaluate([]bool{bits&1 != 0, bits&2 != 0, bits&4 != 0})
		if e != nil {
			t.Fatal(e)
		}
		for _, v := range values {
			if v {
				t.Fatal("zero")
			}
		}
	}
}
func TestGiaObjectOrderValidation(t *testing.T) {
	m := objectOrderSource(false)
	for _, o := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
		g, _, e := m.NormalizeObjectOrder(o)
		if g != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
		g, e = m.ZeroOutputCopy(o)
		if g != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, o := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, _, e := m.NormalizeObjectOrder(o)
		if g != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
		g, e = m.ZeroOutputCopy(o)
		if g != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, threshold := range []int{1, 3, 5} {
		calls := 0
		o := GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }}
		g, _, e := m.NormalizeObjectOrder(o)
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(e)
		}
		calls = 0
		g, e = m.ZeroOutputCopy(o)
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(e)
		}
	}
	if g, _, e := (*GiaManager)(nil).NormalizeObjectOrder(GiaCofactorOptions{}); g != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if g, e := (*GiaManager)(nil).ZeroOutputCopy(GiaCofactorOptions{}); g != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzGiaObjectOrder(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 48 {
			data = data[:48]
		}
		m := NewGiaManager()
		pool := []GiaLit{GiaConst0}
		for i := 0; i < 2; i++ {
			v, _ := m.AppendCI()
			pool = append(pool, v)
		}
		for i, v := range data {
			a := pool[int(v)%len(pool)]
			b := pool[(int(v)+i+1)%len(pool)]
			if a.ID() != b.ID() {
				x, e := m.AppendAnd(a.NotIf(v&1 != 0), b)
				if e != nil {
					t.Fatal(e)
				}
				pool = append(pool, x)
			}
			if v&2 != 0 {
				m.AppendCO(pool[len(pool)-1])
			}
			if i == len(data)/2 {
				c, _ := m.AppendCI()
				pool = append(pool, c)
			}
		}
		m.AppendCO(pool[len(pool)-1])
		m.SetRegisterCount(int(data[0] % uint8(min(m.InputCount(), m.OutputCount())+1)))
		g, _, e := m.NormalizeObjectOrder(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		zero, e := m.ZeroOutputCopy(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for bits := 0; bits < 1<<m.InputCount(); bits++ {
			input := make([]bool, m.InputCount())
			for i := range input {
				input[i] = bits&(1<<i) != 0
			}
			want, _ := m.Evaluate(input)
			got, e := g.Evaluate(input)
			if e != nil || !slices.Equal(want, got) {
				t.Fatal(e)
			}
			values, e := zero.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			for _, v := range values {
				if v {
					t.Fatal("zero")
				}
			}
		}
	})
}
