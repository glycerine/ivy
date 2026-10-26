package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaQuantificationExhaustive(t *testing.T) {
	m := NewGiaManager()
	m.Name = "quantify"
	m.Spec = "quantify.aig"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, b)
	buf, _ := m.AppendBuffer(c.Not())
	mux, _ := m.AppendMuxReal(buf, x, a)
	m.AppendCO(mux.Not())
	before := slices.Clone(m.objects)
	for index := 0; index < 3; index++ {
		for _, exists := range []bool{false, true} {
			g, e := m.quantifyInput(index, exists, GiaCofactorOptions{})
			if e != nil {
				t.Fatal(e)
			}
			if g.InputCount() != 3 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
				t.Fatal("metadata/source")
			}
			for k := 0; k < 8; k++ {
				input := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
				low := slices.Clone(input)
				high := slices.Clone(input)
				low[index] = false
				high[index] = true
				v0, e := m.Evaluate(low)
				if e != nil {
					t.Fatal(e)
				}
				v1, e := m.Evaluate(high)
				if e != nil {
					t.Fatal(e)
				}
				want := v0[0] && v1[0]
				if exists {
					want = v0[0] || v1[0]
				}
				got, e := g.Evaluate(input)
				if e != nil || got[0] != want {
					t.Fatal(index, exists, k, got, want, e)
				}
			}
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.ExistentialInput(0, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 12} {
		calls := 0
		g, e := m.ExistentialInput(0, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	m.AppendCO(a.Not())
	g, e := m.UniversalFirstOutput(0, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	for k := 0; k < 8; k++ {
		v, e := g.Evaluate([]bool{k&1 != 0, k&2 != 0, k&4 != 0})
		if e != nil || v[1] {
			t.Fatal("later outputs must be positive cofactors", v, e)
		}
	}
	if _, e := m.ExistentialInput(0, GiaCofactorOptions{}); e == nil {
		t.Fatal("multiple exists outputs")
	}
	if _, e := m.UniversalFirstOutput(-1, GiaCofactorOptions{}); e == nil {
		t.Fatal("index")
	}
	m.SetConstraintCount(1)
	if _, e := m.UniversalFirstOutput(0, GiaCofactorOptions{}); e == nil {
		t.Fatal("constraints")
	}
}
func FuzzGiaQuantification(f *testing.F) {
	f.Add(uint8(5), uint8(2), uint8(7))
	f.Fuzz(func(t *testing.T, phase, index, assignment uint8) {
		m := NewGiaManager()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		c, _ := m.AppendCI()
		x, _ := m.AppendXorReal(a, b)
		y, _ := m.AppendMuxReal(c, x, a.Not())
		m.AppendCO(y.NotIf(phase&1 != 0))
		pivot := int(index % 3)
		for _, exists := range []bool{false, true} {
			g, e := m.quantifyInput(pivot, exists, GiaCofactorOptions{})
			if e != nil {
				t.Fatal(e)
			}
			evaluate := func(bits uint8) bool {
				a, b, c := bits&1 != 0, bits&2 != 0, bits&4 != 0
				v := !a
				if c {
					v = a != b
				}
				return v != (phase&1 != 0)
			}
			lo := evaluate(assignment &^ (1 << pivot))
			hi := evaluate(assignment | (1 << pivot))
			want := lo && hi
			if exists {
				want = lo || hi
			}
			got, e := g.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0})
			if e != nil || got[0] != want {
				t.Fatal(got, want, e)
			}
		}
	})
}

func TestGiaQuantificationNative(t *testing.T) {
	raw, e := os.ReadFile("test_vectors/oracle/gia_quantify.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects int
		Masks   []uint
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 6 {
		t.Fatal(e)
	}
	for mode := 0; mode < 2; mode++ {
		for pivot := 0; pivot < 3; pivot++ {
			m := NewGiaManager()
			a, _ := m.AppendCI()
			b, _ := m.AppendCI()
			c, _ := m.AppendCI()
			x, _ := m.AppendAnd(a, b)
			y, _ := m.AppendAnd(x.Not(), c)
			m.AppendCO(y)
			if mode != 0 {
				m.AppendCO(a.Not())
			}
			g, e := m.quantifyInput(pivot, mode == 0, GiaCofactorOptions{})
			if e != nil {
				t.Fatal(e)
			}
			masks := make([]uint, g.OutputCount())
			for k := 0; k < 8; k++ {
				v, e := g.Evaluate([]bool{k&1 != 0, k&2 != 0, k&4 != 0})
				if e != nil {
					t.Fatal(e)
				}
				for i, value := range v {
					if value {
						masks[i] |= 1 << k
					}
				}
			}
			row := rows[mode*3+pivot]
			if g.ObjectCount() != row.Objects || !slices.Equal(masks, row.Masks) {
				t.Fatal(mode, pivot, masks, row, g.ObjectCount())
			}
		}
	}
}
