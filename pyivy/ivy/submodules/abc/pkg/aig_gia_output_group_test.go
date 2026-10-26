package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaOutputGroupNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_output_group.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Start, Stop, Objects int
		Inputs               []int
		CiIds, CoIds         []GiaID
		Drivers              []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 15 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		d, _ := m.AppendCI()
		x, _ := m.AppendAnd(b, c)
		y, _ := m.AppendAnd(x, a)
		m.AppendCO(x)
		m.AppendCO(GiaConst1)
		m.AppendCO(y)
		m.AppendCO(d.Not())
		m.SetRegisterCount(2)
		r, e := m.ExtractOutputGroup(row.Start, row.Stop, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		g := r.Graph
		var drivers []GiaLit
		for _, id := range g.cos {
			d, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, d)
		}
		if g.ObjectCount() != row.Objects || !slices.Equal(r.Inputs, row.Inputs) || !slices.Equal(g.cis, row.CiIds) || !slices.Equal(g.cos, row.CoIds) || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(row, g.ObjectCount(), r.Inputs, g.cis, g.cos, drivers)
		}
	}
}
func TestGiaOutputGroupValidation(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	m.AppendCO(x)
	for _, o := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
		r, e := m.ExtractOutputGroup(0, 1, o)
		if r.Graph != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, o := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.ExtractOutputGroup(0, 1, o)
		if r.Graph != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, v := range [][2]int{{-1, 0}, {1, 0}, {0, 2}} {
		r, e := m.ExtractOutputGroup(v[0], v[1], GiaCofactorOptions{})
		if r.Graph != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	if r, e := (*GiaManager)(nil).ExtractOutputGroup(0, 0, GiaCofactorOptions{}); r.Graph != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	for _, threshold := range []int{1, 2, 4, 5} {
		calls := 0
		r, e := m.ExtractOutputGroup(0, 1, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaOutputGroup(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5, 6})
	f.Add([]byte{128, 5, 7, 9})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) == 0 {
			return
		}
		if len(data) > 64 {
			data = data[:64]
		}
		m := NewGiaManager()
		pool := []GiaLit{GiaConst0}
		for i := 0; i < 3; i++ {
			v, _ := m.AppendCI()
			pool = append(pool, v)
		}
		m.HashStart()
		for i := 0; i+1 < len(data); i += 2 {
			a := pool[int(data[i])%len(pool)]
			b := pool[int(data[i+1])%len(pool)]
			var v GiaLit
			var e error
			if data[i]&4 != 0 {
				v, e = m.AppendBuffer(a.Not())
			} else if data[i]&2 != 0 && a.ID() != b.ID() {
				v, e = m.AppendXorReal(a, b)
			} else {
				v, e = m.HashAnd(a.NotIf(data[i]&8 != 0), b)
			}
			if e != nil {
				t.Fatal(e)
			}
			pool = append(pool, v)
		}
		m.HashStop()
		count := 1 + int(data[0]%6)
		for i := 0; i < count; i++ {
			v := data[i%len(data)]
			m.AppendCO(pool[int(v)%len(pool)].NotIf(v&64 != 0))
		}
		m.SetRegisterCount(int(data[0] % uint8(min(3, count)+1)))
		m.SetConstraintCount(1)
		start := int(data[len(data)-1]) % (count + 1)
		stop := start + int(data[0])%(count-start+1)
		before := slices.Clone(m.objects)
		m.Name = "group"
		m.Spec = "group.aig"
		r, e := m.ExtractOutputGroup(start, stop, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		g := r.Graph
		if g.OutputCount() != stop-start || g.registers != 0 || g.constraints != 0 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
			t.Fatal("metadata")
		}
		for bits := 0; bits < 8; bits++ {
			input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			projected := make([]bool, len(r.Inputs))
			for i, j := range r.Inputs {
				if j < 0 || j >= len(input) {
					t.Fatal(j)
				}
				projected[i] = input[j]
			}
			source, e := m.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			got, e := g.Evaluate(projected)
			if e != nil || !slices.Equal(got, source[start:stop]) {
				t.Fatal(start, stop, bits, got, source, e)
			}
		}
	})
}
