package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaDualSideOrderingAndSharing(t *testing.T) {
	for mode := 0; mode < 3; mode++ {
		m := NewGiaManager()
		m.Name = "sides"
		m.Spec = "sides.aig"
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendAnd(a, b)
		y, _ := m.AppendAnd(x, c)
		m.AppendAnd(a, c)
		roots := []GiaLit{y, x}
		if mode == 1 {
			roots = []GiaLit{x, y}
		} else if mode == 2 {
			roots = []GiaLit{y, y.Not()}
		}
		for _, l := range roots {
			m.AppendCO(l)
		}
		m.SetConstraintCount(1)
		before := slices.Clone(m.objects)
		even, e := m.CollectOutputSide(0, 0)
		if e != nil {
			t.Fatal(e)
		}
		odd, e := m.CollectOutputSide(1, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := m.OrderDualOutputSides(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if r.Swapped != (mode == 0) || r.EvenNodes != len(even) || r.OddNodes != len(odd) || r.Graph.constraints != 0 || r.Graph.Name != m.Name || r.Graph.Spec != m.Spec || r.Graph.AndCount() != 2 || !slices.Equal(before, m.objects) {
			t.Fatal(r, mode)
		}
		for k := 0; k < 8; k++ {
			input := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
			want, e := m.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			if r.Swapped {
				want[0], want[1] = want[1], want[0]
			}
			got, e := r.Graph.Evaluate(input)
			if e != nil || !slices.Equal(got, want) {
				t.Fatal(mode, k, got, want, e)
			}
		}
		for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
			r, e := m.OrderDualOutputSides(options)
			if r.Graph != nil || e == nil {
				t.Fatal(options, e)
			}
		}
		for _, threshold := range []int{1, 3, 6, 9} {
			calls := 0
			r, e := m.OrderDualOutputSides(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
			if r.Graph != nil || !errors.Is(e, ErrCancelled) {
				t.Fatal(threshold, calls, e)
			}
		}
	}
}
func TestGiaOutputSideExtendedTraversal(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	buf, _ := m.AppendBuffer(c.Not())
	x, _ := m.AppendXorReal(a, b)
	mux, _ := m.AppendMuxReal(buf, x, a)
	m.AppendCO(mux)
	m.AppendCO(a)
	nodes, e := m.CollectOutputSide(0, 0)
	if e != nil || !slices.Contains(nodes, buf.ID()) || !slices.Contains(nodes, x.ID()) || !slices.Contains(nodes, mux.ID()) {
		t.Fatal(nodes, e)
	}
	if _, e = m.CollectOutputSide(2, 0); e == nil {
		t.Fatal("side")
	}
	r, e := m.OrderDualOutputSides(GiaCofactorOptions{})
	if e != nil || !r.Swapped {
		t.Fatal(r, e)
	}
	for k := 0; k < 8; k++ {
		a, b, c := k&1 != 0, k&2 != 0, k&4 != 0
		v := a
		if !c {
			v = a != b
		}
		got, e := r.Graph.Evaluate([]bool{a, b, c})
		if e != nil || !slices.Equal(got, []bool{a, v}) {
			t.Fatal(got, e)
		}
	}
}
func FuzzGiaDualSideOrdering(f *testing.F) {
	f.Add(uint8(3), uint8(5))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(a, b)
		y, _ := m.AppendMuxReal(c, x, a)
		left, right := y.NotIf(phase&1 != 0), a
		if phase&2 != 0 {
			left, right = right, left
		}
		m.AppendCO(left)
		m.AppendCO(right)
		r, e := m.OrderDualOutputSides(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		av, bv, cv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		v := av
		if cv {
			v = av != bv
		}
		v = v != (phase&1 != 0)
		want := []bool{av, v}
		got, e := r.Graph.Evaluate([]bool{av, bv, cv})
		if e != nil || !slices.Equal(got, want) || r.Swapped != (phase&2 == 0) {
			t.Fatal(r.Swapped, got, want, e)
		}
	})
}

func TestGiaDualSideNative(t *testing.T) {
	raw, e := os.ReadFile("test_vectors/oracle/gia_dual_side.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects   int
		Even, Odd []GiaID
		Drivers   []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 3 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendAnd(a, b)
		y, _ := m.AppendAnd(x, c)
		m.AppendAnd(a, c)
		left, right := y, x
		if mode == 1 {
			left, right = x, y
		} else if mode == 2 {
			left, right = y, y.Not()
		}
		m.AppendCO(left)
		m.AppendCO(right)
		even, e := m.CollectOutputSide(0, 0)
		if e != nil {
			t.Fatal(e)
		}
		odd, e := m.CollectOutputSide(1, 0)
		if e != nil {
			t.Fatal(e)
		}
		r, e := m.OrderDualOutputSides(GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range r.Graph.cos {
			l, _ := r.Graph.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if r.Graph.ObjectCount() != row.Objects || !slices.Equal(even, row.Even) || !slices.Equal(odd, row.Odd) || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(mode, even, odd, drivers, row)
		}
	}
}
