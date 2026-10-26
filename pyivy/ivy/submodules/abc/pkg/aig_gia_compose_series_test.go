package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaSeriesCompositionTruth(t *testing.T) {
	bottom := NewGiaManager()
	bottom.Name = "series"
	bottom.Spec = "series.aig"
	a, _ := bottom.AppendCI()
	b, _ := bottom.AppendCI()
	c, _ := bottom.AppendCI()
	x, _ := bottom.AppendXorReal(a, b)
	buf, _ := bottom.AppendBuffer(c.Not())
	bottom.AppendCO(x)
	bottom.AppendCO(buf)
	bottom.AppendCO(a.Not())
	bottom.AppendCO(GiaConst1)
	top := NewGiaManager()
	ci := make([]GiaLit, 4)
	for i := range ci {
		ci[i], _ = top.AppendCI()
	}
	mux, _ := top.AppendMuxReal(ci[1], ci[0], ci[2])
	top.AppendCO(mux.Not())
	top.AppendCO(ci[3])
	top.SetConstraintCount(1)
	before := slices.Clone(bottom.objects)
	beforeTop := slices.Clone(top.objects)
	g, e := ComposeGiaSeries(bottom, top, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if g.Name != bottom.Name || g.Spec != bottom.Spec || g.constraints != 1 || g.InputCount() != 3 || !slices.Equal(before, bottom.objects) || !slices.Equal(beforeTop, top.objects) {
		t.Fatal("metadata/source")
	}
	for k := 0; k < 8; k++ {
		a, b, c := k&1 != 0, k&2 != 0, k&4 != 0
		v := !a
		if !c {
			v = a != b
		}
		got, e := g.Evaluate([]bool{a, b, c})
		if e != nil || !slices.Equal(got, []bool{!v, true}) {
			t.Fatal(k, got, e)
		}
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := ComposeGiaSeries(bottom, top, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 4, 7, 10} {
		calls := 0
		g, e := ComposeGiaSeries(bottom, top, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	if _, e := ComposeGiaSeries(nil, top, GiaCofactorOptions{}); e == nil {
		t.Fatal("nil")
	}
	bottom.SetConstraintCount(1)
	if _, e := ComposeGiaSeries(bottom, top, GiaCofactorOptions{}); e == nil {
		t.Fatal("bottom constraints")
	}
}
func FuzzGiaSeriesComposition(f *testing.F) {
	f.Add(uint8(3), uint8(7))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		bottom := NewGiaManager()
		a, _ := bottom.AppendCI()
		b, _ := bottom.AppendCI()
		x, _ := bottom.AppendXorReal(a, b)
		bottom.AppendCO(x.NotIf(phase&1 != 0))
		bottom.AppendCO(a.Not())
		top := NewGiaManager()
		u, _ := top.AppendCI()
		v, _ := top.AppendCI()
		y, _ := top.AppendAnd(u, v)
		top.AppendCO(y.NotIf(phase&2 != 0))
		g, e := ComposeGiaSeries(bottom, top, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		av, bv := assignment&1 != 0, assignment&2 != 0
		want := ((((av != bv) != (phase&1 != 0)) && !av) != (phase&2 != 0))
		got, e := g.Evaluate([]bool{av, bv})
		if e != nil || got[0] != want {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaSeriesCompositionNative(t *testing.T) {
	bottom := NewGiaManager()
	a, _ := bottom.AppendCI()
	b, _ := bottom.AppendCI()
	x, _ := bottom.AppendAnd(a, b)
	bottom.AppendCO(x)
	bottom.AppendCO(a.Not())
	top := NewGiaManager()
	u, _ := top.AppendCI()
	v, _ := top.AppendCI()
	y, _ := top.AppendAnd(u, v)
	top.AppendCO(y.Not())
	g, e := ComposeGiaSeries(bottom, top, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_compose_series.json")
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
