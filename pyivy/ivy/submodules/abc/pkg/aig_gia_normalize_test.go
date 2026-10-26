package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaNormalizeNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_normalize.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Hash, Objects, Constraints int
		BufferIds                  []GiaID
		Drivers                    []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 2 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendAnd(a, b)
		y, _ := m.AppendAnd(a, b)
		m.AppendCO(x)
		c, _ := m.AppendCI()
		buf, _ := m.AppendBuffer(c)
		m.AppendAnd(x.Not(), y)
		m.AppendCO(buf)
		m.AppendCO(y.Not())
		m.SetRegisterCount(1)
		m.SetConstraintCount(1)
		g, _, e := m.NormalizeGraph(row.Hash != 0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		var drivers []GiaLit
		var buffers []GiaID
		for i, o := range g.objects {
			if o.IsBuffer() {
				buffers = append(buffers, GiaID(i))
			}
		}
		for _, id := range g.cos {
			d, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, d)
		}
		if g.ObjectCount() != row.Objects || g.constraints != row.Constraints || !slices.Equal(buffers, row.BufferIds) || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(row, g.ObjectCount(), buffers, drivers)
		}
	}
}
func TestGiaNormalizeValidation(t *testing.T) {
	m := objectOrderSource(true)
	m.HashStart()
	before := slices.Clone(m.objects)
	for _, hash := range []bool{false, true} {
		g, _, e := m.NormalizeGraph(hash, GiaCofactorOptions{})
		if e != nil || g.constraints != m.constraints || g.registers != m.registers || g.Name != m.Name || g.Spec != m.Spec || g.hash != nil || m.hash == nil || !slices.Equal(before, m.objects) {
			t.Fatal(e)
		}
		for _, o := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
			g, mp, e := m.NormalizeGraph(hash, o)
			if g != nil || mp != nil || !errors.Is(e, ErrLimit) {
				t.Fatal(e)
			}
		}
		for _, o := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
			g, _, e := m.NormalizeGraph(hash, o)
			if g != nil || !errors.Is(e, ErrDimension) {
				t.Fatal(e)
			}
		}
		for _, threshold := range []int{1, 3, 6} {
			calls := 0
			g, _, e := m.NormalizeGraph(hash, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
			if g != nil || !errors.Is(e, ErrCancelled) {
				t.Fatal(threshold, calls, e)
			}
		}
		if g, _, e := (*GiaManager)(nil).NormalizeGraph(hash, GiaCofactorOptions{}); g != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
}
func FuzzGiaNormalize(f *testing.F) {
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
			a := pool[int(v)%len(pool)].NotIf(v&8 != 0)
			b := pool[(int(v)+i)%len(pool)]
			var x GiaLit
			var e error
			if v&4 != 0 {
				x, e = m.AppendBuffer(a)
			} else if a.ID() != b.ID() {
				if v&2 != 0 {
					x, e = m.AppendXorReal(a, b)
				} else {
					x, e = m.AppendAnd(a, b)
				}
			} else {
				x = a
			}
			if e != nil {
				t.Fatal(e)
			}
			pool = append(pool, x)
			if v&16 != 0 {
				m.AppendCO(x)
			}
			if i == len(data)/2 {
				c, _ := m.AppendCI()
				pool = append(pool, c)
			}
		}
		m.AppendCO(pool[len(pool)-1])
		m.SetRegisterCount(int(data[0] % uint8(min(m.InputCount(), m.OutputCount())+1)))
		if m.PrimaryOutputCount() > 0 {
			m.SetConstraintCount(1)
		}
		before := slices.Clone(m.objects)
		hash := data[0]&1 != 0
		g, mapping, e := m.NormalizeGraph(hash, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if !slices.Equal(before, m.objects) || g.constraints != m.constraints || g.registers != m.registers || g.hash != nil {
			t.Fatal("metadata")
		}
		sourceBuffers, targetBuffers := 0, 0
		for _, o := range m.objects {
			if o.IsBuffer() {
				sourceBuffers++
			}
		}
		for _, o := range g.objects {
			if o.IsBuffer() {
				targetBuffers++
			}
		}
		if sourceBuffers != targetBuffers {
			t.Fatal("buffers")
		}
		for bits := 0; bits < 1<<m.InputCount(); bits++ {
			input := make([]bool, m.InputCount())
			for i := range input {
				input[i] = bits&(1<<i) != 0
			}
			source, e := m.evaluateValues(input)
			if e != nil {
				t.Fatal(e)
			}
			target, e := g.evaluateValues(input)
			if e != nil {
				t.Fatal(e)
			}
			for i, l := range mapping {
				if !l.Valid() || source[i] != (target[l.ID()] != l.Complemented()) {
					t.Fatal(hash, i, l, bits)
				}
			}
		}
	})
}
