package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaDualOutputLayoutsAndRoundtrip(t *testing.T) {
	m := NewGiaManager()
	m.Name = "dual"
	m.Spec = "dual.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	buf, _ := m.AppendBuffer(b.Not())
	mux, _ := m.AppendMuxReal(a, x, buf)
	for _, l := range []GiaLit{mux, x.Not(), p, b, a.Not(), x} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for _, interleave := range []bool{false, true} {
		g, e := m.dualOutputLayout(interleave, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if g.registers != 2 || g.constraints != 0 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source")
		}
		miter, e := g.TransformOutputPairs(false, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		for k := 0; k < 8; k++ {
			input := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
			source, e := m.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			want := make([]bool, 0)
			if interleave {
				want = []bool{source[0], source[2], source[1], source[3]}
			} else {
				for _, v := range source[:4] {
					want = append(want, v, false)
				}
			}
			want = append(want, source[4:]...)
			got, e := g.Evaluate(input)
			if e != nil || !slices.Equal(got, want) {
				t.Fatal(interleave, k, got, want, e)
			}
			compared, e := miter.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			if interleave {
				if !slices.Equal(compared, []bool{source[0] != source[2], source[1] != source[3], source[4], source[5]}) {
					t.Fatal(compared)
				}
			} else if !slices.Equal(compared, source) {
				t.Fatal("dual XOR roundtrip", compared, source)
			}
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.ToDualOutputs(options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 12} {
		calls := 0
		g, e := m.ToDualOutputs(GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	odd := NewGiaManager()
	odd.AppendCO(GiaConst1)
	if _, e := odd.InterleaveOutputHalves(GiaCofactorOptions{}); e == nil {
		t.Fatal("odd")
	}
}
func FuzzGiaDualOutputLayouts(f *testing.F) {
	f.Add(uint8(3), uint8(5))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		for _, l := range []GiaLit{x.NotIf(phase&1 != 0), b, p.Not(), a, p, a.Not()} {
			m.AppendCO(l)
		}
		m.SetRegisterCount(2)
		interleave := phase&2 != 0
		g, e := m.dualOutputLayout(interleave, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		pv, av, bv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		source := []bool{(pv != av) != (phase&1 != 0), bv, !pv, av}
		want := []bool{}
		if interleave {
			want = []bool{source[0], source[2], source[1], source[3]}
		} else {
			for _, v := range source {
				want = append(want, v, false)
			}
		}
		want = append(want, pv, !av)
		got, e := g.Evaluate([]bool{pv, av, bv})
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaDualOutputLayoutsNative(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	m.AppendAnd(p, b)
	for _, l := range []GiaLit{x, b, p.Not(), a, p, a.Not()} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(2)
	raw, e := os.ReadFile("test_vectors/oracle/gia_dual_outputs.json")
	if e != nil {
		t.Fatal(e)
	}
	var rows []struct {
		Objects int
		Drivers []GiaLit
	}
	if e = json.Unmarshal(raw, &rows); e != nil || len(rows) != 2 {
		t.Fatal(e)
	}
	for mode, row := range rows {
		g, e := m.dualOutputLayout(mode != 0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if g.ObjectCount() != row.Objects || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(mode, drivers, row, g.ObjectCount())
		}
	}
}
