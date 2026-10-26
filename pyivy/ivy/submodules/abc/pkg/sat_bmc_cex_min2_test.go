package abc

// Go-only exhaustive tests for src/sat/bmc/bmcCexMin2.c.

import (
	"errors"
	"slices"
	"testing"
)

func cexMin2SequentialGraph(t testing.TB) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	p, _ := g.AppendCI()
	q, _ := g.AppendCI()
	r, _ := g.AppendCI()
	and, err := g.AppendAnd(p, q)
	if err != nil {
		t.Fatal(err)
	}
	g.AppendCO(r)
	g.AppendCO(and)
	if err = g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return g
}

func cexMin2SequentialTrace(t testing.TB) *Counterexample {
	t.Helper()
	trace, err := NewCounterexample(1, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	trace.Property = 0
	trace.SetInput(0, 0, true)
	trace.SetInput(0, 1, true)
	return trace
}

func TestCexMin2AnnotationSequentialValuesAndPath(t *testing.T) {
	g := cexMin2SequentialGraph(t)
	trace := cexMin2SequentialTrace(t)
	before := slices.Clone(g.objects)
	annotation, err := g.AnnotateCounterexampleUnrolling(trace, false, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	changed := g.Clone()
	if err = changed.PatchCODriver(0, GiaConst0); err != nil {
		t.Fatal(err)
	}
	if _, err = changed.CreateCounterexampleUnate(trace, annotation, 0, -1, false, CounterexampleToolOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal("stale annotation", err)
	}
	if annotation.Frames() != 2 || annotation.Objects() != g.ObjectCount() || annotation.Property() != 0 || !slices.Equal(before, g.objects) {
		t.Fatal("annotation metadata/source")
	}
	p, q, r := g.cis[0], g.cis[1], g.cis[2]
	ri, po := g.cos[1], g.cos[0]
	and := GiaID(4)
	checks := []struct {
		frame int
		id    GiaID
		value bool
		path  bool
	}{
		{0, p, true, true}, {0, q, true, true}, {0, r, false, false},
		{0, and, true, true}, {0, ri, true, true}, {0, po, false, false},
		{1, p, false, false}, {1, q, false, false}, {1, r, true, true}, {1, po, true, true},
	}
	for _, check := range checks {
		value, valueErr := annotation.Value(check.frame, check.id)
		path, pathErr := annotation.Justified(check.frame, check.id)
		if valueErr != nil || pathErr != nil || value != check.value || path != check.path {
			t.Fatal(check, value, path, valueErr, pathErr)
		}
	}
	clone := annotation.Clone()
	clone.values[0] = !clone.values[0]
	if clone.values[0] == annotation.values[0] {
		t.Fatal("annotation clone aliases")
	}
}

func TestCexMin2StateAndRefinementNetworks(t *testing.T) {
	g := cexMin2SequentialGraph(t)
	trace := cexMin2SequentialTrace(t)
	stateResult, err := g.CounterexampleUnateNetworks(trace, CexUnateOptions{StartFrame: 0, RealPrimaryInputs: -1})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(stateResult.Frames, []int{1, 0}) || len(stateResult.Networks) != 2 {
		t.Fatal(stateResult.Frames, len(stateResult.Networks))
	}
	last, first := stateResult.Networks[0], stateResult.Networks[1]
	if last.PrimaryInputCount() != 1 || first.PrimaryInputCount() != 0 {
		t.Fatal("state candidate counts", last.PrimaryInputCount(), first.PrimaryInputCount())
	}
	for value := 0; value < 2; value++ {
		outputs, err := last.Evaluate([]bool{value != 0})
		if err != nil || outputs[0] != (value != 0) {
			t.Fatal("last-frame state unate", value, outputs, err)
		}
	}
	outputs, err := first.Evaluate(nil)
	if err != nil || !outputs[0] {
		t.Fatal("first-frame fixed PI justification", outputs, err)
	}
	refinement, err := g.CounterexampleUnateNetworks(trace, CexUnateOptions{StartFrame: 0, RealPrimaryInputs: 1})
	if err != nil || len(refinement.Networks) != 1 || refinement.Networks[0].PrimaryInputCount() != 1 {
		t.Fatal(err)
	}
	for value := 0; value < 2; value++ {
		outputs, err = refinement.Networks[0].Evaluate([]bool{value != 0})
		if err != nil || outputs[0] != (value != 0) {
			t.Fatal("pseudo-PI refinement", value, outputs, err)
		}
	}
	all, err := g.CounterexampleUnateNetworks(trace, CexUnateOptions{StartFrame: 0, RealPrimaryInputs: 1, UseAllObjects: true})
	if err != nil || all.Networks[0].PrimaryInputCount() != 2 {
		t.Fatal("all pseudo candidates", all.Networks[0].PrimaryInputCount(), err)
	}
	for mask := 0; mask < 4; mask++ {
		outputs, err = all.Networks[0].Evaluate([]bool{mask&1 != 0, mask&2 != 0})
		if err != nil || outputs[0] != (mask&1 != 0) {
			t.Fatal("unused candidate changed formula", mask, outputs, err)
		}
	}
}

func TestCexMin2JustificationMaximum(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	and, _ := g.AppendAnd(a, b)
	g.AppendCO(and.Not())
	trace, _ := NewCounterexample(0, 2, 1)
	trace.Property = 0
	for _, maximum := range []bool{false, true} {
		result, err := g.CounterexampleUnateNetworks(trace, CexUnateOptions{RealPrimaryInputs: 0, JustifyMaximum: maximum})
		if err != nil {
			t.Fatal(maximum, err)
		}
		network := result.Networks[0]
		wantInputs := 1
		if maximum {
			wantInputs = 2
		}
		if network.PrimaryInputCount() != wantInputs {
			t.Fatal(maximum, network.PrimaryInputCount())
		}
		for mask := 0; mask < 1<<wantInputs; mask++ {
			inputs := make([]bool, wantInputs)
			for input := range inputs {
				inputs[input] = mask>>input&1 != 0
			}
			outputs, err := network.Evaluate(inputs)
			want := inputs[0]
			if maximum {
				want = inputs[0] || inputs[1]
			}
			if err != nil || outputs[0] != want {
				t.Fatal(maximum, mask, outputs, want, err)
			}
		}
	}
}

func TestCexMin2ValidationLimitsCancellationAndDeepCone(t *testing.T) {
	g := cexMin2SequentialGraph(t)
	trace := cexMin2SequentialTrace(t)
	before := slices.Clone(g.objects)
	if _, err := g.AnnotateCounterexampleUnrolling(trace, false, CounterexampleToolOptions{MaxMappingEntries: g.ObjectCount()*trace.Frames() - 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("annotation entry limit", err)
	}
	if _, err := g.AnnotateCounterexampleUnrolling(trace, false, CounterexampleToolOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal("annotation cancellation", err)
	}
	annotation, err := g.AnnotateCounterexampleUnrolling(trace, false, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.CreateCounterexampleUnate(trace, annotation, 0, -1, false, CounterexampleToolOptions{MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("network object limit", err)
	}
	if _, err = g.CreateCounterexampleUnate(trace, annotation, 0, -1, false, CounterexampleToolOptions{Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal("network cancellation", err)
	}
	if _, err = g.CounterexampleUnateNetworks(trace, CexUnateOptions{StartFrame: -1, RealPrimaryInputs: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal("start frame", err)
	}
	bad := trace.Clone()
	bad.Property = 1
	if _, err = g.AnnotateCounterexampleUnrolling(bad, false, CounterexampleToolOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal("property dimension", err)
	}
	extended := NewGiaManager()
	x, _ := extended.AppendCI()
	y, _ := extended.AppendCI()
	z, _ := extended.AppendXorReal(x, y)
	extended.AppendCO(z)
	extendedTrace, _ := NewCounterexample(0, 2, 1)
	extendedTrace.Property = 0
	extendedTrace.SetInput(0, 0, true)
	if _, err = extended.AnnotateCounterexampleUnrolling(extendedTrace, false, CounterexampleToolOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal("extended graph", err)
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("source mutation")
	}
	deep := NewGiaManager()
	a, _ := deep.AppendCI()
	b, _ := deep.AppendCI()
	node, _ := deep.AppendAnd(a, b)
	for index := 0; index < 10000; index++ {
		fanin := a
		if index&1 != 0 {
			fanin = b
		}
		node, err = deep.AppendAnd(node, fanin)
		if err != nil {
			t.Fatal(index, err)
		}
	}
	deep.AppendCO(node)
	deepTrace, _ := NewCounterexample(0, 2, 1)
	deepTrace.Property = 0
	deepTrace.SetInput(0, 0, true)
	deepTrace.SetInput(0, 1, true)
	if _, err = deep.CounterexampleUnateNetworks(deepTrace, CexUnateOptions{RealPrimaryInputs: 0}); err != nil {
		t.Fatal("deep unate", err)
	}
}

func FuzzCexMin2UnateSufficiency(f *testing.F) {
	f.Add(uint8(8), uint8(3), false)
	f.Add(uint8(0x96), uint8(2), true)
	f.Fuzz(func(t *testing.T, truth, rawAssignment uint8, maximum bool) {
		assignment := int(rawAssignment & 3)
		if truth>>uint(assignment)&1 == 0 {
			return
		}
		g := ecoTruthGraph(t, truth)
		trace, err := NewCounterexample(0, 2, 1)
		if err != nil {
			t.Fatal(err)
		}
		trace.Property = 0
		trace.SetInput(0, 0, assignment&1 != 0)
		trace.SetInput(0, 1, assignment&2 != 0)
		result, err := g.CounterexampleUnateNetworks(trace, CexUnateOptions{RealPrimaryInputs: 0, JustifyMaximum: maximum, UseAllObjects: true})
		if err != nil {
			t.Fatal(truth, assignment, maximum, err)
		}
		network := result.Networks[0]
		if network.PrimaryInputCount() != 2 {
			t.Fatal("candidate shape", network.PrimaryInputCount())
		}
		for fixedMask := 0; fixedMask < 4; fixedMask++ {
			outputs, evalErr := network.Evaluate([]bool{fixedMask&1 != 0, fixedMask&2 != 0})
			if evalErr != nil {
				t.Fatal(evalErr)
			}
			if !outputs[0] {
				continue
			}
			for candidate := 0; candidate < 4; candidate++ {
				compatible := true
				for input := 0; input < 2; input++ {
					if fixedMask>>input&1 != 0 && candidate>>input&1 != assignment>>input&1 {
						compatible = false
					}
				}
				if compatible && truth>>uint(candidate)&1 == 0 {
					t.Fatal("unate formula is insufficient", truth, assignment, maximum, fixedMask, candidate)
				}
			}
		}
	})
}
