package abc

import (
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"
)

func TestGiaExplicitSupportAndInsertion(t *testing.T) {
	m := NewGiaManager()
	m.Name = "support"
	m.Spec = "support.aig"
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	x, _ := m.AppendXorReal(a, c)
	buf, _ := m.AppendBuffer(b.Not())
	mux, _ := m.AppendMuxReal(buf, x, a)
	m.AppendCO(mux)
	before := slices.Clone(m.objects)
	cone, e := m.ExtractSignalWithInputs(mux.Not(), []int{2, 0, 1}, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if cone.registers != 0 || cone.constraints != 0 || cone.Name != m.Name || cone.Spec != m.Spec || cone.XorCount() != 1 || cone.MuxCount() != 1 || cone.BufferCount() != 1 {
		t.Fatal("cone metadata")
	}
	g, inserted, e := m.WithInsertedCone(cone, []GiaLit{c, a, b}, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	if g.OutputCount() != m.OutputCount() || !slices.Equal(before, m.objects) || g.Name != m.Name {
		t.Fatal("source/target metadata")
	}
	g.AppendCO(inserted)
	for k := 0; k < 8; k++ {
		a, b, c := k&1 != 0, k&2 != 0, k&4 != 0
		want := a
		if !b {
			want = a != c
		}
		extracted, e := cone.Evaluate([]bool{c, a, b})
		if e != nil || extracted[0] == want {
			t.Fatal(k, extracted, want, e)
		}
		got, e := g.Evaluate([]bool{a, b, c})
		if e != nil || !slices.Equal(got, []bool{want, !want}) {
			t.Fatal(k, got, want, e)
		}
	}
	if _, e := m.ExtractSignalWithInputs(mux, []int{0, 2}, GiaCofactorOptions{}); e == nil {
		t.Fatal("incomplete support")
	}
	for _, indices := range [][]int{{0, 0}, {-1}, {3}} {
		if _, e := m.ExtractSignalWithInputs(a, indices, GiaCofactorOptions{}); e == nil {
			t.Fatal(indices)
		}
	}
	constants, e := m.ExtractSignalWithInputs(GiaConst1, nil, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	_, lit, e := m.WithInsertedCone(constants, nil, GiaCofactorOptions{})
	if e != nil || lit != GiaConst1 {
		t.Fatal(lit, e)
	}
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		if g, l, e := m.WithInsertedCone(cone, []GiaLit{c, a, b}, options); g != nil || l != InvalidGiaLit || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 2, 4} {
		calls := 0
		g, l, e := m.WithInsertedCone(cone, []GiaLit{c, a, b}, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || l != InvalidGiaLit || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
}
func FuzzGiaConeSupportInsertion(f *testing.F) {
	f.Add(uint8(5), uint8(7))
	f.Fuzz(func(t *testing.T, phase, assignment uint8) {
		source := NewGiaManager()
		a, _ := source.AppendCI()
		b, _ := source.AppendCI()
		x, _ := source.AppendXorReal(a.NotIf(phase&1 != 0), b)
		source.AppendCO(x)
		cone, e := source.ExtractSignalWithInputs(x.NotIf(phase&2 != 0), []int{1, 0}, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		target := NewGiaManager()
		p, _ := target.AppendCI()
		q, _ := target.AppendCI()
		g, l, e := target.WithInsertedCone(cone, []GiaLit{p.NotIf(phase&4 != 0), q}, GiaCofactorOptions{})
		if e != nil {
			t.Fatal(e)
		}
		g.AppendCO(l)
		pv, qv := assignment&1 != 0, assignment&2 != 0
		want := ((qv != (phase&1 != 0)) != (pv != (phase&4 != 0))) != (phase&2 != 0)
		got, e := g.Evaluate([]bool{pv, qv})
		if e != nil || got[0] != want {
			t.Fatal(got, want, e)
		}
	})
}

func TestGiaConeSupportNative(t *testing.T) {
	m := NewGiaManager()
	a, _ := m.AppendCI()
	b, _ := m.AppendCI()
	c, _ := m.AppendCI()
	root, _ := m.AppendAnd(a, c)
	m.AppendAnd(a, b)
	m.AppendCO(root.Not())
	cone, e := m.ExtractSignalWithInputs(root.Not(), []int{2, 0, 1}, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	g, back, e := m.WithInsertedCone(cone, []GiaLit{c, a, b}, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	g, backObjects, e := g.WithInsertedCone(cone, []GiaLit{root, a, b}, GiaCofactorOptions{})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := os.ReadFile("test_vectors/oracle/gia_cone_support.json")
	if e != nil {
		t.Fatal(e)
	}
	var native struct {
		ConeObjects, TargetObjects    int
		ConeDriver, Back, BackObjects GiaLit
	}
	if e = json.Unmarshal(raw, &native); e != nil {
		t.Fatal(e)
	}
	co := cone.cos[0]
	driver, _ := cone.objects[co].Fanin0(co)
	if cone.ObjectCount() != native.ConeObjects || g.ObjectCount() != native.TargetObjects || driver != native.ConeDriver || back != native.Back || backObjects != native.BackObjects {
		t.Fatal(driver, back, backObjects, native)
	}
}

func TestGiaSupportCopyLimitsAndActiveHash(t *testing.T) {
	source := NewGiaManager()
	a, _ := source.AppendCI()
	b, _ := source.AppendCI()
	x, _ := source.AppendXorReal(a, b)
	source.AppendCO(x)
	for _, options := range []GiaCofactorOptions{{MaxObjects: 1}, {MaxMappingEntries: 1}, {MaxObjects: -1}, {MaxMappingEntries: -1}} {
		g, e := source.ExtractSignalWithInputs(x, []int{0, 1}, options)
		if g != nil || e == nil {
			t.Fatal(options, e)
		}
	}
	for _, threshold := range []int{1, 2, 4} {
		calls := 0
		g, e := source.ExtractSignalWithInputs(x, []int{0, 1}, GiaCofactorOptions{Stop: func() bool { calls++; return calls >= threshold }})
		if g != nil || !errors.Is(e, ErrCancelled) {
			t.Fatal(threshold, calls, e)
		}
	}
	target := NewGiaManager()
	p, _ := target.AppendCI()
	target.AppendCO(p)
	if e := target.HashStart(); e != nil {
		t.Fatal(e)
	}
	before := slices.Clone(target.objects)
	g, l, e := target.WithInsertedCone(source, []GiaLit{p, p}, GiaCofactorOptions{})
	if e != nil || l != GiaConst0 || g.hash == nil || target.hash == nil || !slices.Equal(before, target.objects) {
		t.Fatal(l, e)
	}
	beforeCount := target.ObjectCount()
	g, l, e = target.WithInsertedCone(source, []GiaLit{p, InvalidGiaLit}, GiaCofactorOptions{})
	if g != nil || l != InvalidGiaLit || e == nil || target.ObjectCount() != beforeCount {
		t.Fatal("invalid binding", e)
	}
}
