package abc

// Go-only semantic tests for src/sat/bmc/bmcCexCare.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestCounterexamplePriorityCareRecurrences(t *testing.T) {
	for _, test := range []struct {
		name       string
		complement bool
		a, b       bool
		wantA      bool
		wantB      bool
	}{
		{"true AND", false, true, true, true, true},
		{"single controlling zero", true, true, false, false, true},
		{"double controlling zero prefers first PI", true, false, false, true, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := NewGiaManager()
			a, _ := g.AppendCI()
			b, _ := g.AppendCI()
			and, _ := g.AppendAnd(a, b)
			g.AppendCO(and.NotIf(test.complement))
			trace, _ := NewCounterexample(0, 2, 1)
			trace.Property = 0
			trace.SetInput(0, 0, test.a)
			trace.SetInput(0, 1, test.b)
			result, err := g.MinimizeCounterexampleCarePriority(trace, CexPriorityCareOptions{RealPrimaryInputs: 2, Attempts: 4})
			if err != nil || !result.Verified || len(result.Candidates) != 4 {
				t.Fatal(err, result.Verified, len(result.Candidates))
			}
			for attempt, care := range result.Candidates {
				gotA, _ := care.Input(0, 0)
				gotB, _ := care.Input(0, 1)
				if gotA != test.wantA || gotB != test.wantB {
					t.Fatal(attempt, gotA, gotB)
				}
			}
		})
	}
}

func TestCounterexamplePriorityCareInitialStateAndIntersection(t *testing.T) {
	g := cexDepthSelfLoopGraph(t)
	trace, _ := NewCounterexample(1, 1, 1)
	trace.Property = 0
	trace.SetInitial(0, true)
	result, err := g.MinimizeCounterexampleCarePriority(trace, CexPriorityCareOptions{RealPrimaryInputs: 1, Attempts: 1})
	if err != nil || result.Best.CountOnes() != 0 {
		t.Fatal(err, result.Best)
	}
	first, _ := NewCounterexample(0, 2, 1)
	second, _ := NewCounterexample(0, 2, 1)
	first.Property, second.Property = 0, 0
	first.SetInput(0, 0, true)
	first.SetInput(0, 1, true)
	second.SetInput(0, 1, true)
	intersection, err := IntersectCounterexampleCare(first, second)
	if err != nil || intersection.CountOnes() != 1 {
		t.Fatal(err, intersection)
	}
	one, _ := intersection.Input(0, 1)
	if !one || first.CountOnes() != 2 {
		t.Fatal("intersection ownership")
	}
}

func TestCounterexampleCareObjectExtensionAndAigAdapters(t *testing.T) {
	g := NewGiaManager()
	a, _ := g.AppendCI()
	b, _ := g.AppendCI()
	and, _ := g.AppendAnd(a, b)
	g.AppendCO(and)
	trace, _ := NewCounterexample(0, 2, 1)
	trace.Property = 0
	trace.SetInput(0, 0, true)
	trace.SetInput(0, 1, true)
	care, _ := NewCounterexample(0, 2, 1)
	care.Property = 0
	care.SetInput(0, 0, true)
	if extended, err := g.ExtendCounterexampleCareToObjects(trace, care, CounterexampleToolOptions{}); extended != nil || !errors.Is(err, ErrInvariant) {
		t.Fatal("insufficient care", err)
	}
	care.SetInput(0, 1, true)
	extended, err := g.ExtendCounterexampleCareToObjects(trace, care, CounterexampleToolOptions{})
	if err != nil || extended.Inputs() != g.ObjectCount() || extended.CountOnes() != g.ObjectCount() {
		t.Fatal(err, extended)
	}
	aig, _, err := g.ToAig()
	if err != nil {
		t.Fatal(err)
	}
	priority, err := aig.MinimizeCounterexampleCarePriority(trace, CexPriorityCareOptions{RealPrimaryInputs: 2, Attempts: 4})
	if err != nil || !priority.Verified {
		t.Fatal(err)
	}
	verified, err := aig.VerifyCounterexampleCare(trace, priority.Best, CounterexampleToolOptions{})
	if err != nil || !verified {
		t.Fatal(verified, err)
	}
	for _, highEffort := range []bool{false, true} {
		minimized, minimizeErr := aig.MinimizeCounterexampleCareSAT(trace, 2, highEffort, CounterexampleToolOptions{})
		if minimizeErr != nil {
			t.Fatal(highEffort, minimizeErr)
		}
		verified, minimizeErr = aig.VerifyCounterexampleCare(trace, minimized, CounterexampleToolOptions{})
		if minimizeErr != nil || !verified {
			t.Fatal(highEffort, verified, minimizeErr)
		}
	}
}

func TestCounterexamplePriorityCareLimitsCancellationAndDepth(t *testing.T) {
	g := cexToolSequentialGraph(t)
	trace := cexToolSequentialTrace(t)
	for _, test := range []struct {
		options CexPriorityCareOptions
		want    error
	}{
		{CexPriorityCareOptions{RealPrimaryInputs: -1, Attempts: 1}, ErrDimension},
		{CexPriorityCareOptions{RealPrimaryInputs: 2, Attempts: 1}, ErrDimension},
		{CexPriorityCareOptions{RealPrimaryInputs: 1, Attempts: 0}, ErrDimension},
		{CexPriorityCareOptions{RealPrimaryInputs: 1, Attempts: 5}, ErrDimension},
		{CexPriorityCareOptions{CounterexampleToolOptions: CounterexampleToolOptions{MaxObjects: 1}, RealPrimaryInputs: 1, Attempts: 1}, ErrLimit},
		{CexPriorityCareOptions{CounterexampleToolOptions: CounterexampleToolOptions{MaxMappingEntries: 1}, RealPrimaryInputs: 1, Attempts: 1}, ErrLimit},
		{CexPriorityCareOptions{CounterexampleToolOptions: CounterexampleToolOptions{Stop: func() bool { return true }}, RealPrimaryInputs: 1, Attempts: 1}, ErrCancelled},
	} {
		if result, err := g.MinimizeCounterexampleCarePriority(trace, test.options); result.Best != nil || !errors.Is(err, test.want) {
			t.Fatal(test.options, err, test.want)
		}
	}
	extended := NewGiaManager()
	a, _ := extended.AppendCI()
	b, _ := extended.AppendCI()
	x, _ := extended.AppendXorReal(a, b)
	extended.AppendCO(x)
	extendedTrace, _ := NewCounterexample(0, 2, 1)
	extendedTrace.Property = 0
	extendedTrace.SetInput(0, 0, true)
	if _, err := extended.MinimizeCounterexampleCarePriority(extendedTrace, CexPriorityCareOptions{RealPrimaryInputs: 2, Attempts: 1}); !errors.Is(err, ErrUnsupported) {
		t.Fatal("extended recurrence", err)
	}
	badCare, _ := NewCounterexample(0, 2, 2)
	badCare.Property = 0
	if _, err := g.ExtendCounterexampleCareToObjects(trace, badCare, CounterexampleToolOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal("object care dimensions", err)
	}
	first, _ := NewCounterexample(0, 1, 1)
	second, _ := NewCounterexample(0, 2, 1)
	if _, err := IntersectCounterexampleCare(first, second); !errors.Is(err, ErrDimension) {
		t.Fatal("intersection dimensions", err)
	}

	deep := NewGiaManager()
	deep.AppendCI()
	r0, _ := deep.AppendCI()
	r1, _ := deep.AppendCI()
	current, _ := deep.AppendAnd(r0, r1)
	for index := 0; index < 10_000; index++ {
		current, _ = deep.AppendAnd(current, r0)
	}
	deep.AppendCO(current)
	deep.AppendCO(r0)
	deep.AppendCO(r1)
	deep.SetRegisterCount(2)
	deepTrace, _ := NewCounterexample(2, 1, 2)
	deepTrace.Property = 0
	deepTrace.SetInitial(0, true)
	deepTrace.SetInitial(1, true)
	before := slices.Clone(deep.objects)
	result, err := deep.MinimizeCounterexampleCarePriority(deepTrace, CexPriorityCareOptions{RealPrimaryInputs: 1, Attempts: 4})
	if err != nil || !result.Verified || result.Best.CountOnes() != 0 || !slices.Equal(before, deep.objects) {
		t.Fatal("deep iterative priority", err, result.Verified)
	}
}

func FuzzCounterexamplePriorityCareSoundness(f *testing.F) {
	for _, seed := range [][5]uint8{{8, 10, 1, 1, 4}, {6, 9, 0, 7, 3}, {14, 3, 1, 5, 2}, {1, 12, 0, 2, 1}} {
		f.Add(seed[0], seed[1], seed[2], seed[3], seed[4])
	}
	f.Fuzz(func(t *testing.T, outputTable, nextTable, initial, primaryBits, attemptsByte uint8) {
		g := cexDepthTruthGraph(t, outputTable, nextTable)
		trace, err := NewCounterexample(1, 1, 3)
		if err != nil {
			t.Fatal(err)
		}
		trace.Property = 0
		state := initial&1 != 0
		trace.SetInitial(0, state)
		for frame := 0; frame < trace.Frames(); frame++ {
			primary := primaryBits&(1<<frame) != 0
			trace.SetInput(frame, 0, primary)
			if frame == trace.LastFrame() && !cexDepthTruth(outputTable, primary, state) {
				return
			}
			state = cexDepthTruth(nextTable, primary, state)
		}
		attempts := int(attemptsByte%4) + 1
		result, err := g.MinimizeCounterexampleCarePriority(trace, CexPriorityCareOptions{RealPrimaryInputs: int(attemptsByte & 1), Attempts: attempts})
		if err != nil {
			t.Fatal(err)
		}
		care := result.Best
		free := make([][2]int, 0, trace.Frames())
		for frame := 0; frame < trace.Frames(); frame++ {
			known, _ := care.Input(frame, 0)
			if !known {
				free = append(free, [2]int{frame, 0})
			}
		}
		for assignment := 0; assignment < 1<<len(free); assignment++ {
			candidateState := initial&1 != 0
			final := false
			for frame := 0; frame < trace.Frames(); frame++ {
				primary, _ := trace.Input(frame, 0)
				for index, location := range free {
					if location[0] == frame {
						primary = assignment&(1<<index) != 0
					}
				}
				final = cexDepthTruth(outputTable, primary, candidateState)
				candidateState = cexDepthTruth(nextTable, primary, candidateState)
			}
			if !final {
				t.Fatal("care mask admits non-failure", outputTable&15, nextTable&15, care.PackedWords(), assignment)
			}
		}
	})
}
