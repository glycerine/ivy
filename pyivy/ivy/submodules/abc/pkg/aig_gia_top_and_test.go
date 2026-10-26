package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func topAndSource(mode int) *GiaManager {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, b)
	y, _ := m.AppendAnd(b, c)
	root := x
	switch mode {
	case 1:
		root, _ = m.AppendAnd(x.Not(), y.Not())
	case 2, 6:
		root, _ = m.AppendAnd(a, y.Not())
	case 3:
		root = x.Not()
	case 4:
		root, _ = m.AppendAnd(x, a.Not())
	}
	m.AppendCO(root)
	if mode == 7 {
		m.AppendCO(GiaConst0)
	}
	if mode == 5 {
		m.AppendCO(GiaConst1)
	}
	if mode == 6 {
		m.AppendCO(x.Not())
		m.SetRegisterCount(1)
	}
	return m
}
func TestGiaTopAndNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_top_and.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Mode, Full, Objects, Registers int
		Drivers                        []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 16 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m := topAndSource(row.Mode)
		var r GiaTopAndReduction
		if row.Full != 0 {
			r, e = m.ReduceTopAnd(GiaTopAndOptions{})
		} else {
			r, e = m.ReduceTopAndOnce(GiaCofactorOptions{})
		}
		if e != nil {
			t.Fatal(e)
		}
		if row.Mode == 7 {
			if !r.Contradiction || r.Graph.OutputCount() != m.OutputCount() || len(row.Drivers) != 0 {
				t.Fatal("constant-false regression", row, r.Contradiction)
			}
			continue
		}
		var drivers []GiaLit
		g := r.Graph
		for _, id := range g.cos {
			d, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, d)
		}
		if g.ObjectCount() != row.Objects || g.registers != row.Registers || !slices.Equal(drivers, row.Drivers) || r.Contradiction != (row.Mode == 4) {
			t.Fatal(row, g.ObjectCount(), g.registers, drivers, r.Contradiction)
		}
	}
}
func checkTopAndQuery(t *testing.T, m *GiaManager, full bool) {
	t.Helper()
	before := slices.Clone(m.objects)
	var r GiaTopAndReduction
	var e error
	if full {
		r, e = m.ReduceTopAnd(GiaTopAndOptions{})
	} else {
		r, e = m.ReduceTopAndOnce(GiaCofactorOptions{})
	}
	if e != nil {
		t.Fatal(e)
	}
	if len(r.Fixed) != m.InputCount() || r.Graph.InputCount() != m.InputCount() || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || !slices.Equal(before, m.objects) {
		t.Fatal("metadata")
	}
	for bits := 0; bits < 1<<m.InputCount(); bits++ {
		input := make([]bool, m.InputCount())
		for i := range input {
			input[i] = bits&(1<<i) != 0
		}
		out, e := m.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		want := true
		for _, v := range out {
			want = want && v
		}
		gotOut, e := r.Graph.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		got := !r.Contradiction
		for _, v := range gotOut {
			got = got && v
		}
		for i, l := range r.Fixed {
			if l != InvalidGiaLit {
				if l != GiaConst0 && l != GiaConst1 {
					t.Fatal(l)
				}
				got = got && (input[i] == (l == GiaConst1))
			}
		}
		if got != want {
			t.Fatal(bits, full, want, got, r.Fixed, r.Contradiction)
		}
	}
}
func TestGiaTopAndConstantsAndLimits(t *testing.T) {
	for mode := 0; mode < 6; mode++ {
		for _, full := range []bool{false, true} {
			m := topAndSource(mode)
			m.Name = "top"
			m.Spec = "top.aig"
			checkTopAndQuery(t, m, full)
		}
	}
	m := topAndSource(0)
	m.AppendCO(GiaConst0)
	for _, full := range []bool{false, true} {
		checkTopAndQuery(t, m, full)
	}
	r, e := m.ReduceTopAndOnce(GiaCofactorOptions{})
	if e != nil || !r.Contradiction || r.Graph.OutputCount() != m.OutputCount() {
		t.Fatal("false CO must survive", e)
	}
	for _, o := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
		r, e := m.ReduceTopAndOnce(o)
		if r.Graph != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, o := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		r, e := m.ReduceTopAndOnce(o)
		if r.Graph != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, threshold := range []int{1, 3, 6} {
		calls := 0
		r, e := topAndSource(1).ReduceTopAndOnce(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if r, e := (*GiaManager)(nil).ReduceTopAndOnce(GiaCofactorOptions{}); r.Graph != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if r, e := topAndSource(0).ReduceTopAnd(GiaTopAndOptions{MaxIterations: 1}); r.Graph != nil || !errors.Is(e, ErrLimit) {
		t.Fatal(e)
	}
	if r, e := m.ReduceTopAnd(GiaTopAndOptions{MaxIterations: -1}); r.Graph != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzGiaTopAnd(f *testing.F) {
	f.Add([]byte{1, 2, 3, 4, 5})
	f.Add([]byte{128, 9, 10, 11})
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
		}
		for i := 0; i < int(data[0]%7); i++ {
			v := data[i%len(data)]
			m.AppendCO(pool[int(v)%len(pool)].NotIf(v&64 != 0))
		}
		if m.OutputCount() > 0 {
			m.SetConstraintCount(1)
		}
		checkTopAndQuery(t, m, data[0]&128 != 0)
	})
}
