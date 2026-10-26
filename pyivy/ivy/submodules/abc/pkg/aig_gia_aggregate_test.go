package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaAggregateNative(t *testing.T) {
	data, e := os.ReadFile("test_vectors/oracle/gia_aggregate.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Empty, UseOr, Flip, Outputs, Objects int
		Drivers                              []GiaLit
	}
	if e = json.Unmarshal(data, &rows); e != nil {
		t.Fatal(e)
	}
	if len(rows) != 24 {
		t.Fatal(len(rows))
	}
	for _, r := range rows {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendAnd(a, b)
		if r.Empty == 0 {
			m.AppendCO(x)
			m.AppendCO(a.Not())
			m.SetConstraintCount(1)
		}
		m.Name = "aggregate"
		m.Spec = "discard.aig"
		before := slices.Clone(m.objects)
		g, e := m.AggregateOutputs(r.Outputs, r.UseOr != 0, r.Flip != 0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		var drivers []GiaLit
		for _, id := range g.cos {
			d, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, d)
		}
		if g.ObjectCount() != r.Objects || !slices.Equal(drivers, r.Drivers) || g.InputCount() != 2 || g.Name != m.Name || g.Spec != "" || g.constraints != 0 || !slices.Equal(before, m.objects) {
			t.Fatal(r, g.ObjectCount(), drivers)
		}
		for bits := 0; bits < 4; bits++ {
			input := []bool{bits&1 != 0, bits&2 != 0}
			want := r.UseOr == 0
			if r.Empty == 0 {
				if r.UseOr != 0 {
					want = (input[0] && input[1]) || !input[0]
				} else {
					want = (input[0] && input[1]) && !input[0]
				}
			}
			want = want != (r.Flip != 0)
			got, e := g.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			for _, value := range got {
				if value != want {
					t.Fatal(r, bits, got, want)
				}
			}
		}
	}
}

func TestGiaAggregateValidation(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	m.AppendCO(x)
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}} {
		g, e := m.AggregateOutputs(2, false, false, options)
		if g != nil || !errors.Is(e, ErrLimit) {
			t.Fatal(e)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.AggregateOutputs(1, false, false, options)
		if g != nil || !errors.Is(e, ErrDimension) {
			t.Fatal(e)
		}
	}
	for _, threshold := range []int{1, 3, 5, 7} {
		calls := 0
		g, e := m.AggregateOutputs(3, true, false, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if g, e := m.AggregateOutputs(-1, false, false, GiaCofactorOptions{}); g != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	if g, e := (*GiaManager)(nil).AggregateOutputs(1, false, false, GiaCofactorOptions{}); g != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
	m.SetRegisterCount(1)
	if g, e := m.AggregateOutputs(1, false, false, GiaCofactorOptions{}); g != nil || !errors.Is(e, ErrDimension) {
		t.Fatal(e)
	}
}

func FuzzGiaAggregate(f *testing.F) {
	f.Add(uint8(0), uint8(0))
	f.Add(uint8(15), uint8(7))
	f.Fuzz(func(t *testing.T, flags, bits uint8) {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(a, b)
		buf, _ := m.AppendBuffer(b.Not())
		mux, _ := m.AppendMuxReal(c, x, buf)
		roots := []GiaLit{x.NotIf(flags&16 != 0), mux, buf.Not()}
		count := int((flags >> 5) % 4)
		for _, r := range roots[:count] {
			m.AppendCO(r)
		}
		useOr, flip := flags&1 != 0, flags&2 != 0
		n := int((flags >> 2) % 4)
		g, e := m.AggregateOutputs(n, useOr, flip, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		av, bv, cv := bits&1 != 0, bits&2 != 0, bits&4 != 0
		xv := av != bv
		mv := !bv
		if cv {
			mv = xv
		}
		values := []bool{xv != (flags&16 != 0), mv, bv}
		want := !useOr
		for _, v := range values[:count] {
			if useOr {
				want = want || v
			} else {
				want = want && v
			}
		}
		want = want != flip
		got, e := g.Evaluate([]bool{av, bv, cv})
		if e != nil || len(got) != n {
			t.Fatal(got, e)
		}
		for _, v := range got {
			if v != want {
				t.Fatal(flags, bits, got, want)
			}
		}
	})
}
