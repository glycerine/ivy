package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaOutputPairMiterTruth(t *testing.T) {
	m := NewGiaManager()
	m.Name = "pairs"
	m.Spec = "pairs.aig"
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendXorReal(p, a)
	buf, _ := m.AppendBuffer(b.Not())
	mux, _ := m.AppendMuxReal(a, x, buf)
	m.AppendCO(mux)
	m.AppendCO(x.Not())
	m.AppendCO(p)
	m.AppendCO(b)
	m.AppendCO(a.Not())
	m.AppendCO(x)
	m.SetRegisterCount(2)
	m.SetConstraintCount(1)
	before := slices.Clone(m.objects)
	for _, split := range []bool{false, true} {
		g, e := m.TransformOutputPairs(split, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		if g.registers != 2 || g.constraints != 0 || g.OutputCount() != 4 || g.Name != m.Name || g.Spec != m.Spec || !slices.Equal(before, m.objects) {
			t.Fatal("metadata/source")
		}
		for k := 0; k < 8; k++ {
			input := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
			want, e := m.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			got, e := g.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			expected := []bool{want[0] != want[1], want[2] != want[3], want[4], want[5]}
			if split {
				expected[0], expected[1] = want[0] != want[2], want[1] != want[3]
			}
			if !slices.Equal(got, expected) {
				t.Fatal(split, k, got, expected)
			}
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := m.TransformOutputPairs(false, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 8, 11} {
		calls := 0
		g, e := m.TransformOutputPairs(false, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func TestGiaNonconstantPairSelectionDropsBothConstants(t *testing.T) {
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	for _, l := range []GiaLit{GiaConst0, GiaConst0, GiaConst1, GiaConst0, p, p, p, p.Not(), a, GiaConst0, p, b} {
		m.AppendCO(l)
	}
	m.AppendCO(a)
	m.AppendCO(p)
	m.SetRegisterCount(2)
	indices, e := m.NonconstantOutputPairIndices(GiaStructuralSweepOptions{})
	if e != nil || !slices.Equal(indices, []int{5}) {
		t.Fatal(indices, e)
	}
	odd := NewGiaManager()
	odd.AppendCO(GiaConst0)
	if _, e := odd.TransformOutputPairs(false, GiaCofactorOptions{}); e == nil {
		t.Fatal("odd POs")
	}
}
func FuzzGiaOutputPairMiter(f *testing.F) {
	f.Add(uint8(3), uint8(7))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		for _, l := range []GiaLit{x.NotIf(phase&1 != 0), b, p.Not(), a} {
			m.AppendCO(l)
		}
		m.AppendCO(p)
		m.AppendCO(a.Not())
		m.SetRegisterCount(2)
		split := phase&2 != 0
		g, e := m.TransformOutputPairs(split, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		pv, av, bv := assignment&1 != 0, assignment&2 != 0, assignment&4 != 0
		v := []bool{(pv != av) != (phase&1 != 0), bv, !pv, av}
		want := []bool{v[0] != v[1], v[2] != v[3], pv, !av}
		if split {
			want[0], want[1] = v[0] != v[2], v[1] != v[3]
		}
		got, e := g.Evaluate([]bool{pv, av, bv})
		if e != nil || !slices.Equal(got, want) {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaPairMiterNative(t *testing.T) {
	raw, e := os.ReadFile("test_vectors/oracle/gia_pair_miter.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct {
		Transforms []struct {
			Objects int
			Drivers []GiaLit
		}
		Selected []int
	}
	if e = json.Unmarshal(raw, &native); e != nil || len(native.Transforms) != 2 {
		t.Fatal(e)
	}
	m := NewGiaManager()
	p, _ := m.AppendCI()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	x, _ := m.AppendAnd(p, a)
	for _, l := range []GiaLit{x, b, p.Not(), a, p, a.Not()} {
		m.AppendCO(l)
	}
	m.SetRegisterCount(2)
	for mode, row := range native.Transforms {
		g, e := m.TransformOutputPairs(mode != 0, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		drivers := make([]GiaLit, 0)
		for _, id := range g.cos {
			l, _ := g.objects[id].Fanin0(id)
			drivers = append(drivers, l)
		}
		if g.ObjectCount() != row.Objects || !slices.Equal(drivers, row.Drivers) {
			t.Fatal(mode, drivers, row)
		}
	}
	s := NewGiaManager()
	p, _ = s.AppendCI()
	a, _ = s.AppendCI()
	b, _ = s.AppendCI()
	for _, l := range []GiaLit{GiaConst0, GiaConst0, GiaConst1, GiaConst0, p, p, p, p.Not(), a, GiaConst0, p, b, a, p} {
		s.AppendCO(l)
	}
	s.SetRegisterCount(2)
	selected, e := s.NonconstantOutputPairIndices(GiaStructuralSweepOptions{})
	if e != nil || !slices.Equal(selected, native.Selected) {
		t.Fatal(selected, native.Selected, e)
	}
}
