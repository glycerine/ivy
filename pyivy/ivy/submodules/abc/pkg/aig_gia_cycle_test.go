package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func cycleSource() *GiaManager {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	s, _ := m.AppendCI()
	r, _ := m.AppendCI()
	x, _ := m.AppendAnd(a, s)
	y, _ := m.AppendAnd(b.Not(), r)
	z, _ := m.AppendAnd(x.Not(), y.Not())
	m.AppendCO(z.Not())
	m.AppendCO(s.Not())
	m.AppendCO(y.Not())
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	m.Name = "cycle"
	m.Spec = "cycle.aig"
	return m
}
func TestGiaCycleNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_cycle.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Trace, Frames int
		Reached       []int
		Drivers       []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 12 {
		t.Fatal(len(rows))
	}
	for _, row := range rows {
		m := cycleSource()
		before := slices.Clone(m.objects)
		var trace *Counterexample
		if row.Trace != 0 {
			trace, _ = NewCounterexample(2, 2, 6)
			trace.SetInitial(0, true)
			trace.SetInitial(1, true)
			for i := 0; i < 6; i++ {
				for k := 0; k < 2; k++ {
					trace.SetInput(i, k, (i*3+k+1)%4 < 2)
				}
			}
		}
		r, e := m.CycleFromZero(trace, row.Frames, GiaCycleOptions{})
		if e != nil {
			t.Fatal(e)
		}
		g := r.Graph
		var drivers []GiaLit
		for _, id := range g.cos {
			d, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, d)
		}
		if !slices.Equal(drivers, row.Drivers) || g.ObjectCount() != m.ObjectCount() || g.constraints != 0 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
			t.Fatal(row, drivers)
		}
		for i, v := range r.Reached {
			if v != (row.Reached[i] != 0) {
				t.Fatal(row, r.Reached)
			}
		}
		for bits := 0; bits < 16; bits++ {
			pi := []bool{bits&1 != 0, bits&2 != 0}
			state := []bool{bits&4 != 0, bits&8 != 0}
			sourceState := []bool{state[0] != r.Reached[0], state[1] != r.Reached[1]}
			want, next, e := m.Step(pi, sourceState)
			if e != nil {
				t.Fatal(e)
			}
			got, targetNext, e := g.Step(pi, state)
			if e != nil || !slices.Equal(got, want) {
				t.Fatal(e)
			}
			for i, v := range next {
				if targetNext[i] != (v != r.Reached[i]) {
					t.Fatal("transition recoding")
				}
			}
		}
	}
}
func TestGiaCycleValidation(t *testing.T) {
	m := cycleSource()
	c, _ := NewCounterexample(2, 2, 2)
	for _, o := range []GiaCycleOptions{{MaxFrames: 1}, {GiaCofactorOptions: GiaCofactorOptions{MaxObjects: 1}}, {GiaCofactorOptions: GiaCofactorOptions{MaxMappingEntries: 1}}} {
		r, e := m.CycleFromZero(nil, 2, o)
		if r.Graph != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, o := range []GiaCycleOptions{{MaxFrames: -1}, {GiaCofactorOptions: GiaCofactorOptions{MaxObjects: -1}}, {GiaCofactorOptions: GiaCofactorOptions{MaxMappingEntries: -1}}} {
		r, e := m.CycleFromZero(nil, 0, o)
		if r.Graph != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	if r, e := m.CycleFromZero(c, 2, GiaCycleOptions{}); r.Graph != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	wrong, _ := NewCounterexample(1, 2, 4)
	if r, e := m.CycleFromZero(wrong, 0, GiaCycleOptions{}); r.Graph != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	for _, threshold := range []int{1, 4, 8, 12} {
		calls := 0
		r, e := m.CycleFromZero(nil, 5, GiaCycleOptions{GiaCofactorOptions: GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }}})
		if r.Graph != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if r, e := (*GiaManager)(nil).CycleFromZero(nil, 0, GiaCycleOptions{}); r.Graph != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if r, e := m.CycleFromZero(nil, -1, GiaCycleOptions{}); r.Graph != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}
func FuzzGiaCycle(f *testing.F) {
	f.Add(uint8(5), uint16(137))
	f.Fuzz(func(t *testing.T, flags uint8, pattern uint16) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		q, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		y, _ := m.AppendMuxReal(q, x, b)
		m.AppendCO(y.NotIf(flags&1 != 0))
		m.AppendCO(x.Not())
		m.AppendCO(y)
		m.SetRegisterCount(2)
		n := int(flags % 9)
		c, _ := NewCounterexample(2, 2, n+1)
		for i := 0; i < n; i++ {
			for k := 0; k < 2; k++ {
				c.SetInput(i, k, pattern&(1<<((i*2+k)%16)) != 0)
			}
		}
		words := c.PackedWords()
		r, e := m.CycleFromZero(c, n, GiaCycleOptions{})
		if e != nil {
			t.Fatal(e)
		}
		state := []bool{false, false}
		for i := 0; i < n; i++ {
			pv := pattern&(1<<((i*2)%16)) != 0
			qv := pattern&(1<<((i*2+1)%16)) != 0
			xv := pv != state[0]
			yv := state[1]
			if qv {
				yv = xv
			}
			state = []bool{!xv, yv}
		}
		if !slices.Equal(state, r.Reached) || !slices.Equal(words, c.PackedWords()) {
			t.Fatal(flags, pattern, state, r.Reached)
		}
		for bits := 0; bits < 16; bits++ {
			pv, qv, av, bv := bits&1 != 0, bits&2 != 0, bits&4 != 0, bits&8 != 0
			oldA, oldB := av != state[0], bv != state[1]
			xv := pv != oldA
			yv := oldB
			if qv {
				yv = xv
			}
			po, next, e := r.Graph.Step([]bool{pv, qv}, []bool{av, bv})
			if e != nil || po[0] != (yv != (flags&1 != 0)) || next[0] != (!xv != state[0]) || next[1] != (yv != state[1]) {
				t.Fatal(bits, po, next, e)
			}
		}
	})
}
