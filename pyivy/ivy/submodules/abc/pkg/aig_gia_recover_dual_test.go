package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaRecoverDualNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_recover_dual.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Mode, Objects int
		Drivers       []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 9 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		d, _ := m.AppendCI()
		m.HashStart()
		x, _ := m.HashXor(a, b)
		y, _ := m.HashXor(c.Not(), d)
		z, _ := m.HashOr(x, y)
		switch row.Mode {
		case 0:
			m.AppendCO(z)
		case 1:
			m.AppendCO(z.Not())
		case 2:
			m.AppendCO(x)
			m.AppendCO(y)
			m.AppendCO(a)
		case 3:
			m.AppendCO(GiaConst0)
		case 4:
			m.AppendCO(GiaConst1)
		case 5:
			m.AppendCO(GiaConst1)
			m.AppendCO(x)
		case 6:
			m.AppendCO(y)
			m.AppendCO(x)
			m.AppendCO(x)
			m.AppendCO(GiaConst0)
		}

		if row.Mode >= 7 {
			h1, _ := m.HashAnd(a, c)
			h2, _ := m.HashAnd(b, d)
			h3, _ := m.HashAnd(a, d)
			k, _ := m.HashXor(h1, h2)
			w, _ := m.HashXor(h1, h3)
			m.AppendCO(k)
			m.AppendCO(w)
			if row.Mode == 7 {
				m.AppendCO(y)
			} else {
				m.AppendCO(h2)
			}
		}
		m.HashStop()
		r, e := m.RecoverDualOutputs(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		g := r.Graph
		var drivers []GiaLit
		for _, id := range g.cos {
			v, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, v)
		}
		if g.ObjectCount() != row.Objects || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(row, g.ObjectCount(), drivers, r.Candidates, r.Signals)
		}
	}
}
func TestGiaRecoverDualValidation(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	m.AppendCO(x.Not())
	for _, o := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
		r, e := m.RecoverDualOutputs(o)
		if r.Graph != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, o := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.RecoverDualOutputs(o)
		if r.Graph != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 12} {
		calls := 0
		r, e := m.RecoverDualOutputs(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	for _, source := range []*GiaManager{nil, NewGiaManager()} {
		r, e := source.RecoverDualOutputs(GiaCofactorOptions{})
		if r.Graph != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
}
func FuzzGiaRecoverDual(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{128, 6, 7, 8, 9})
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
		for i := 0; i+1 < len(data); i += 2 {
			a := pool[int(data[i])%len(pool)].NotIf(data[i]&128 != 0)
			b := pool[int(data[i+1])%len(pool)].NotIf(data[i+1]&128 != 0)
			var v GiaLit
			var e error
			if data[i]&4 != 0 {
				v, e = m.AppendBuffer(a)
			} else if data[i]&2 != 0 && a.ID() != b.ID() {
				v, e = m.AppendXorReal(a, b)
			} else if data[i]&1 != 0 {
				v, e = m.HashXor(a, b)
			} else {
				v, e = m.HashAnd(a, b)
			}
			if e != nil {
				t.Fatal(e)
			}
			pool = append(pool, v)
		}
		m.HashStop()
		for i := 0; i < 1+int(data[0]%5); i++ {
			v := data[i%len(data)]
			m.AppendCO(pool[int(v)%len(pool)].NotIf(v&64 != 0))
		}
		if data[0]&32 != 0 {
			m.SetRegisterCount(1)
		}
		m.SetConstraintCount(1)
		m.Name = "recovery"
		m.Spec = "recovery.aig"
		before := slices.Clone(m.objects)
		r, e := m.RecoverDualOutputs(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if r.Graph.registers != 0 || r.Graph.constraints != 0 || r.Graph.InputCount() != m.InputCount() || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || !slices.Equal(before, m.objects) || len(r.Signals)%2 != 0 {
			t.Fatal("metadata")
		}
		for bits := 0; bits < 8; bits++ {
			input := []bool{bits&1 != 0, bits&2 != 0, bits&4 != 0}
			values, e := m.evaluateValues(input)
			if e != nil {
				t.Fatal(e)
			}
			got, e := r.Graph.Evaluate(input)
			if e != nil || len(got) != len(r.Signals) {
				t.Fatal(e)
			}
			for i, l := range r.Signals {
				want := values[l.ID()] != l.Complemented()
				if got[i] != want {
					t.Fatal(data, bits, got, r.Signals)
				}
			}
		}
	})
}
