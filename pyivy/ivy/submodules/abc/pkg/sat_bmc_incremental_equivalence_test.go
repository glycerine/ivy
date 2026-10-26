package abc

// Go-only semantic, limit, and fuzz tests for src/sat/bmc/bmcBmci.c.

import (
	"errors"
	"testing"
)

func bmciAbsorptionGraph(t testing.TB) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	p, _ := g.AppendCI()
	r, _ := g.AppendCI()
	if err := g.HashStart(); err != nil {
		t.Fatal(err)
	}
	or, err := g.HashOr(p, r)
	if err != nil {
		t.Fatal(err)
	}
	next, err := g.HashAnd(p, or)
	if err != nil {
		t.Fatal(err)
	}
	g.HashStop()
	g.AppendCO(r)
	g.AppendCO(next)
	g.SetRegisterCount(1)
	return g
}

func TestGiaBmciUnfoldAndEquivalence(t *testing.T) {
	absorption := bmciAbsorptionGraph(t)
	result, err := absorption.CompareInitialStateEvolution(
		[]GiaBmciInitial{GiaBmciVariable}, []GiaBmciInitial{GiaBmciVariable}, GiaBmciOptions{Frames: 3},
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != GiaBmciComplete || !result.FixedPoint || len(result.Frames) != 1 || result.Frames[0].Miters != 1 || result.Frames[0].Merged != 1 || result.Frames[0].SAT != 0 || result.MergedStates != 1 || result.SATChecks != 1 {
		t.Fatal("absorption fixed point", result)
	}
	if result.Left[0] != result.Right[0] || result.Graph == nil || result.Graph.Name != absorption.Name {
		t.Fatal("merged state", result.Left, result.Right, result.Graph, result.Graph.Name, absorption.Name)
	}

	self := cexDepthSelfLoopGraph(t)
	distinct, err := self.CompareInitialStateWithZero([]GiaBmciInitial{GiaBmciVariable}, GiaBmciOptions{Frames: 3})
	if err != nil {
		t.Fatal(err)
	}
	if distinct.FixedPoint || len(distinct.Frames) != 3 || distinct.SATChecks != 3 {
		t.Fatal("persistent difference", distinct)
	}
	for _, frame := range distinct.Frames {
		if frame.SAT != 1 || frame.Miters != 1 || frame.Merged != 0 {
			t.Fatal(frame)
		}
	}

	driven := cexToolSequentialGraph(t)
	converged, err := driven.CompareInitialStateWithZero([]GiaBmciInitial{GiaBmciVariable}, GiaBmciOptions{Frames: 3})
	if err != nil || !converged.FixedPoint || len(converged.Frames) != 1 || converged.Frames[0].Miters != 0 {
		t.Fatal("shared-input convergence", converged, err)
	}
}

func TestGiaBmciErrorsLimitsAndCancellation(t *testing.T) {
	g := cexDepthSelfLoopGraph(t)
	for _, test := range []struct {
		left, right []GiaBmciInitial
		options     GiaBmciOptions
		want        error
	}{
		{[]GiaBmciInitial{GiaBmciZero}, []GiaBmciInitial{GiaBmciZero}, GiaBmciOptions{}, ErrDimension},
		{nil, []GiaBmciInitial{GiaBmciZero}, GiaBmciOptions{Frames: 1}, ErrDimension},
		{[]GiaBmciInitial{9}, []GiaBmciInitial{GiaBmciZero}, GiaBmciOptions{Frames: 1}, ErrDimension},
		{[]GiaBmciInitial{GiaBmciVariable}, []GiaBmciInitial{GiaBmciZero}, GiaBmciOptions{Frames: 1, MaxObjects: 1}, ErrLimit},
		{[]GiaBmciInitial{GiaBmciVariable}, []GiaBmciInitial{GiaBmciZero}, GiaBmciOptions{Frames: 1, MaxMappingEntries: 1}, ErrDimension},
		{[]GiaBmciInitial{GiaBmciVariable}, []GiaBmciInitial{GiaBmciZero}, GiaBmciOptions{Frames: 1, MaxVariables: 1}, ErrLimit},
		{[]GiaBmciInitial{GiaBmciVariable}, []GiaBmciInitial{GiaBmciZero}, GiaBmciOptions{Frames: 1, SAT: SatLimits{Stop: func() bool { return true }}}, ErrCancelled},
	} {
		result, err := g.CompareInitialStateEvolution(test.left, test.right, test.options)
		if result.Graph != nil || !errors.Is(err, test.want) {
			t.Fatal(test.options, err, test.want)
		}
	}
	extended := NewGiaManager()
	p, _ := extended.AppendCI()
	r, _ := extended.AppendCI()
	x, _ := extended.AppendXorReal(p, r)
	extended.AppendCO(x)
	extended.AppendCO(r)
	extended.SetRegisterCount(1)
	if _, err := extended.CompareInitialStateWithZero([]GiaBmciInitial{GiaBmciVariable}, GiaBmciOptions{Frames: 1}); !errors.Is(err, ErrUnsupported) {
		t.Fatal("extended source", err)
	}
	comb := NewGiaManager()
	p, _ = comb.AppendCI()
	comb.AppendCO(p)
	if _, err := comb.CompareInitialStateEvolution(nil, nil, GiaBmciOptions{Frames: 1}); !errors.Is(err, ErrDimension) {
		t.Fatal("combinational source", err)
	}
	absorption := bmciAbsorptionGraph(t)
	for _, options := range []GiaBmciOptions{{Frames: 1, MaxClauses: 1}, {Frames: 1, MaxLiterals: 1}} {
		if _, err := absorption.CompareInitialStateEvolution([]GiaBmciInitial{GiaBmciVariable}, []GiaBmciInitial{GiaBmciVariable}, options); !errors.Is(err, ErrLimit) {
			t.Fatal("CNF limit", options, err)
		}
	}
	unknown, err := g.CompareInitialStateWithZero([]GiaBmciInitial{GiaBmciVariable}, GiaBmciOptions{Frames: 1, SAT: SatLimits{Propagations: 1}})
	if err != nil || unknown.Status != GiaBmciUnknown || unknown.Reason != SatPropagationLimit {
		t.Fatal("SAT budget", unknown.Status, unknown.Reason, err)
	}
}

func TestGiaBmciDeepIterativeCone(t *testing.T) {
	g := NewGiaManager()
	p, _ := g.AppendCI()
	r, _ := g.AppendCI()
	current := r
	for index := 0; index < 10_000; index++ {
		var err error
		current, err = g.AppendAnd(current, p)
		if err != nil {
			t.Fatal(index, err)
		}
	}
	g.AppendCO(current)
	g.AppendCO(current)
	g.SetRegisterCount(1)
	result, err := g.CompareInitialStateWithZero([]GiaBmciInitial{GiaBmciVariable}, GiaBmciOptions{
		Frames: 1, MaxObjects: 30_000, MaxMappingEntries: 30_000, MaxVariables: 30_000, MaxClauses: 40_000, MaxLiterals: 100_000,
	})
	if err != nil || result.Status != GiaBmciComplete || len(result.Frames) != 1 || result.Frames[0].SAT != 1 {
		t.Fatal(result, err)
	}
}

func FuzzGiaBmciEquivalence(f *testing.F) {
	for _, seed := range []uint8{0, 1, 3, 5, 10, 15} {
		f.Add(seed, uint8(3))
	}
	f.Fuzz(func(t *testing.T, nextTable, framesByte uint8) {
		frames := int(framesByte%3) + 1
		g := cexDepthTruthGraph(t, 0, nextTable)
		result, err := g.CompareInitialStateEvolution(
			[]GiaBmciInitial{GiaBmciVariable}, []GiaBmciInitial{GiaBmciZero}, GiaBmciOptions{Frames: frames},
		)
		if err != nil {
			t.Fatal(err)
		}
		for frameIndex, frameResult := range result.Frames {
			mismatch := false
			inputAssignments := 1 << (frameIndex + 1)
			for initial := 0; initial < 2; initial++ {
				for inputs := 0; inputs < inputAssignments; inputs++ {
					left, right := initial != 0, false
					for frame := 0; frame <= frameIndex; frame++ {
						primary := inputs&(1<<frame) != 0
						left = cexDepthTruth(nextTable, primary, left)
						right = cexDepthTruth(nextTable, primary, right)
					}
					mismatch = mismatch || left != right
				}
			}
			if (frameResult.SAT > 0) != mismatch {
				t.Fatal(nextTable&15, frameIndex, frameResult, mismatch)
			}
			if !mismatch {
				if !result.FixedPoint || frameIndex != len(result.Frames)-1 {
					t.Fatal("fixed-point reporting", result)
				}
				break
			}
		}
	})
}
