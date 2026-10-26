package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaChoiceMiterCollectionAndAliasOwnership(t *testing.T) {
	build := func(flip bool) *GiaManager {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendXorReal(p, a)
		buf, _ := m.AppendBuffer(b.Not())
		mux, _ := m.AppendMuxReal(a, x, buf)
		m.AppendCO(mux.NotIf(flip))
		m.AppendCO(p)
		m.AppendCO(x)
		m.AppendCO(a.Not())
		m.SetRegisterCount(2)
		m.SetConstraintCount(1)
		return m
	}
	a, b := build(false), build(true)
	a.Name = "choice"
	a.Spec = "choice.aig"
	before := slices.Clone(a.objects)
	sources := []*GiaManager{a, a, b}
	g, e := BuildGiaChoiceMiter(sources, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if g.OutputCount() != 12 || g.registers != 2 || g.constraints != 0 || g.Name != a.Name || g.Spec != a.Spec || !slices.Equal(before, a.objects) {
		t.Fatal("metadata/source")
	}
	for k := 0; k < 8; k++ {
		input := []bool{k&1 != 0, k&2 != 0, k&4 != 0}
		got, e := g.Evaluate(input)
		if e != nil {
			t.Fatal(e)
		}
		for i, source := range sources {
			want, e := source.Evaluate(input)
			if e != nil {
				t.Fatal(e)
			}
			for output, value := range want {
				if got[output*3+i] != value {
					t.Fatal(k, i, output, got, want)
				}
			}
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := BuildGiaChoiceMiter(sources, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 9, 18} {
		calls := 0
		g, e := BuildGiaChoiceMiter(sources, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	for _, sources := range [][]*GiaManager{nil, {nil}, {a, NewGiaManager()}} {
		if _, e := BuildGiaChoiceMiter(sources, GiaCofactorOptions{}); e == nil {
			t.Fatal("invalid sources")
		}
	}
}
func FuzzGiaChoiceMiter(f *testing.F) {
	f.Add(uint8(5), uint8(7))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		sources := make([]*GiaManager, 3)
		for i := range sources {
			m := NewGiaManager()
			p, _ := m.AppendCI()
			a, _ := m.AppendCI()
			x, _ := m.AppendXorReal(p, a)
			m.AppendCO(x.NotIf(phase&(1<<i) != 0))
			m.AppendCO(p.Not())
			m.SetRegisterCount(1)
			sources[i] = m
		}
		g, e := BuildGiaChoiceMiter(sources, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		p, a := assignment&1 != 0, assignment&2 != 0
		got, e := g.Evaluate([]bool{p, a})
		if e != nil {
			t.Fatal(e)
		}
		for i := range sources {
			if got[i] != ((p != a) != (phase&(1<<i) != 0)) || got[3+i] != !p {
				t.Fatal(got)
			}
		}
	})
}

func TestGiaChoiceMiterNative(t *testing.T) {
	build := func(flip bool) *GiaManager {
		m := NewGiaManager()
		p, _ := m.AppendCI()
		a, _ := m.AppendCI()
		b, _ := m.AppendCI()
		x, _ := m.AppendAnd(p, a)
		for _, l := range []GiaLit{x.NotIf(flip), p, b, a.Not()} {
			m.AppendCO(l)
		}
		m.SetRegisterCount(2)
		return m
	}
	g, e := BuildGiaChoiceMiter([]*GiaManager{build(false), build(false), build(true)}, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_choice_miter.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct {
		Objects int
		Drivers []GiaLit
	}
	if e = json.Unmarshal(raw, &native); e != nil {
		t.Fatal(e)
	}
	drivers := make([]GiaLit, 0)
	for _, id := range g.cos {
		l, _ := g.objects[id].Fanin0(id)
		drivers = append(drivers, l)
	}
	if g.ObjectCount() != native.Objects || !slices.Equal(drivers, native.Drivers) {
		t.Fatal(drivers, native)
	}
}
