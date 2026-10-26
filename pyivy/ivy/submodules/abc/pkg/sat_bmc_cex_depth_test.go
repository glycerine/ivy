package abc

// Go-only semantic, limit, and fuzz tests for src/sat/bmc/bmcCexDepth.c.

import (
	"errors"
	"slices"
	"testing"
)

func cexDepthSelfLoopGraph(t testing.TB) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	if _, err := g.AppendCI(); err != nil {
		t.Fatal(err)
	}
	r, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(r); err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(r); err != nil {
		t.Fatal(err)
	}
	if err = g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestEnlargeCounterexampleTargetReversedFrames(t *testing.T) {
	g := cexToolSequentialGraph(t)
	g.Name, g.Spec = "depth", "depth.aig"
	target, err := g.EnlargeCounterexampleTarget(2, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if target.Name != g.Name || target.Spec != g.Spec || target.RegisterCount() != 0 || target.InputCount() != 3 || target.PrimaryOutputCount() != 1 {
		t.Fatal("target shape or metadata")
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		outputs, evalErr := target.Evaluate(inputs)
		if evalErr != nil || outputs[0] != inputs[1] {
			t.Fatal(assignment, inputs, outputs, evalErr)
		}
	}
	if outputs, stepErr := g.Evaluate([]bool{false, true}); stepErr != nil || !outputs[0] || outputs[1] {
		t.Fatal("source graph changed", outputs, stepErr)
	}
}

func TestCounterexampleReachabilityTarget(t *testing.T) {
	g := cexDepthSelfLoopGraph(t)
	reachability, err := g.CounterexampleReachabilityTarget(3, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if reachability.PrimaryInputCount() != 1 || reachability.RegisterCount() != 1 || reachability.PrimaryOutputCount() != 2 {
		t.Fatal("appended target shape")
	}
	for assignment := 0; assignment < 4; assignment++ {
		p, r := assignment&1 != 0, assignment&2 != 0
		got, evalErr := reachability.Evaluate([]bool{p, r})
		want := []bool{r, r, r}
		if evalErr != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, evalErr)
		}
	}

	driven := cexToolSequentialGraph(t)
	quantified, err := driven.CounterexampleReachabilityTarget(2, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		p, r := assignment&1 != 0, assignment&2 != 0
		got, evalErr := quantified.Evaluate([]bool{p, r})
		want := []bool{r, true, p}
		if evalErr != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, evalErr)
		}
	}
}

func TestCounterexampleDepthCones(t *testing.T) {
	g := cexToolSequentialGraph(t)
	trace := cexToolSequentialTrace(t)
	states, _, err := g.CounterexampleInnerStates(trace, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, prune := range []bool{false, true} {
		last, buildErr := g.BuildCounterexampleDepthCone(states, 1, prune, CounterexampleToolOptions{})
		if buildErr != nil {
			t.Fatal(prune, buildErr)
		}
		first, buildErr := g.BuildCounterexampleDepthCone(states, 0, prune, CounterexampleToolOptions{})
		if buildErr != nil {
			t.Fatal(prune, buildErr)
		}
		if last.Name != "unate" || last.InputCount() != 1 || last.PrimaryOutputCount() != 1 || first.InputCount() != 1 {
			t.Fatal("cone shape")
		}
		for _, state := range []bool{false, true} {
			lastOut, evalErr := last.Evaluate([]bool{state})
			if evalErr != nil || lastOut[0] != state {
				t.Fatal("last cone", prune, state, lastOut, evalErr)
			}
			firstOut, evalErr := first.Evaluate([]bool{state})
			if evalErr != nil || !firstOut[0] {
				t.Fatal("first cone", prune, state, firstOut, evalErr)
			}
		}
	}
}

func TestCounterexampleDepthTargetsAndAnalysis(t *testing.T) {
	g := cexToolSequentialGraph(t)
	trace := cexToolSequentialTrace(t)
	states, _, err := g.CounterexampleInnerStates(trace, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	targets, err := g.BuildCounterexampleDepthTargets(states, 20, CexDepthOptions{PruneImplied: true})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(targets.Starts, []int{1}) || len(targets.Cones) != 1 || targets.Graph.PrimaryOutputCount() != 2 {
		t.Fatal("target selection", targets.Starts)
	}
	for assignment := 0; assignment < 4; assignment++ {
		p, r := assignment&1 != 0, assignment&2 != 0
		got, evalErr := targets.Graph.Evaluate([]bool{p, r})
		want := []bool{r, r, p}
		if evalErr != nil || !slices.Equal(got, want) {
			t.Fatal(assignment, got, want, evalErr)
		}
	}
	empty, err := g.BuildCounterexampleDepthTargets(states, 0, CexDepthOptions{})
	if err != nil || len(empty.Starts) != 0 || len(empty.Cones) != 0 || empty.Graph.PrimaryOutputCount() != 1 {
		t.Fatal("empty targets", empty.Starts, err)
	}
	result, err := g.AnalyzeCounterexampleDepth(trace, 1, CexDepthOptions{PruneImplied: true})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Analysis.CareVerified || !result.Analysis.MinimizedVerified || !slices.Equal(result.Starts, []int{1}) || result.Graph == nil {
		t.Fatal("analysis result")
	}
}

func TestCounterexampleDepthExtendedAndFailures(t *testing.T) {
	extended := NewGiaManager()
	p, _ := extended.AppendCI()
	r, _ := extended.AppendCI()
	x, _ := extended.AppendXorReal(p, r)
	extended.AppendCO(x)
	extended.AppendCO(r)
	extended.SetRegisterCount(1)
	enlarged, err := extended.EnlargeCounterexampleTarget(2, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal("enlargement supports semantic extended gates", err)
	}
	for assignment := 0; assignment < 8; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0, assignment&4 != 0}
		got, evalErr := enlarged.Evaluate(inputs)
		want := inputs[0] != inputs[2]
		if evalErr != nil || got[0] != want {
			t.Fatal(assignment, got, want, evalErr)
		}
	}
	states, _ := NewCounterexample(0, 2, 1)
	states.Property = 0
	states.SetInput(0, 0, true)
	if cone, buildErr := extended.BuildCounterexampleDepthCone(states, 0, false, CounterexampleToolOptions{}); cone != nil || !errors.Is(buildErr, ErrUnsupported) {
		t.Fatal("extended depth recurrence", buildErr)
	}

	g := cexToolSequentialGraph(t)
	trace := cexToolSequentialTrace(t)
	validStates, _, err := g.CounterexampleInnerStates(trace, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	badStates, _ := NewCounterexample(0, 1, 1)
	badStates.Property = 0
	for _, test := range []struct {
		name string
		call func() error
		want error
	}{
		{"zero frames", func() error { _, e := g.EnlargeCounterexampleTarget(0, CounterexampleToolOptions{}); return e }, ErrDimension},
		{"frame limit", func() error {
			_, e := g.EnlargeCounterexampleTarget(2, CounterexampleToolOptions{MaxFrames: 1})
			return e
		}, ErrLimit},
		{"mapping limit", func() error {
			_, e := g.EnlargeCounterexampleTarget(1, CounterexampleToolOptions{MaxMappingEntries: 1})
			return e
		}, ErrLimit},
		{"object limit", func() error {
			_, e := g.EnlargeCounterexampleTarget(1, CounterexampleToolOptions{MaxObjects: 1})
			return e
		}, ErrLimit},
		{"cancel enlarge", func() error {
			_, e := g.EnlargeCounterexampleTarget(1, CounterexampleToolOptions{Stop: func() bool { return true }})
			return e
		}, ErrCancelled},
		{"bad expanded dimensions", func() error {
			_, e := g.BuildCounterexampleDepthCone(badStates, 0, false, CounterexampleToolOptions{})
			return e
		}, ErrDimension},
		{"bad start", func() error {
			_, e := g.BuildCounterexampleDepthCone(validStates, 2, false, CounterexampleToolOptions{})
			return e
		}, ErrDimension},
		{"bit limit", func() error {
			_, e := g.BuildCounterexampleDepthCone(validStates, 0, false, CounterexampleToolOptions{MaxBits: 1})
			return e
		}, ErrLimit},
		{"cancel cone", func() error {
			_, e := g.BuildCounterexampleDepthCone(validStates, 0, false, CounterexampleToolOptions{Stop: func() bool { return true }})
			return e
		}, ErrCancelled},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := test.call()
			if !errors.Is(got, test.want) {
				t.Fatal(got, test.want)
			}
		})
	}
	if _, err = g.BuildCounterexampleDepthTargets(validStates, 1, CexDepthOptions{MaxCones: -1}); !errors.Is(err, ErrDimension) {
		t.Fatal("negative cone limit", err)
	}
	longTrace, _ := NewCounterexample(1, 1, 3)
	longTrace.Property = 0
	longTrace.SetInput(0, 0, true)
	longTrace.SetInput(1, 0, true)
	longStates, _, err := g.CounterexampleInnerStates(longTrace, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.BuildCounterexampleDepthTargets(longStates, 2, CexDepthOptions{MaxCones: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("cone limit", err)
	}
}

func TestCounterexampleDepthDeepCone(t *testing.T) {
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
	g.AppendCO(r)
	g.SetRegisterCount(1)
	trace, _ := NewCounterexample(1, 1, 2)
	trace.Property = 0
	trace.SetInitial(0, true)
	trace.SetInput(0, 0, true)
	trace.SetInput(1, 0, true)
	states, _, err := g.CounterexampleInnerStates(trace, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	cone, err := g.BuildCounterexampleDepthCone(states, 1, true, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, state := range []bool{false, true} {
		outputs, evalErr := cone.Evaluate([]bool{state})
		if evalErr != nil || outputs[0] != state {
			t.Fatal(state, outputs, evalErr)
		}
	}
}

func cexDepthTruthGraph(t testing.TB, outputTable, nextTable uint8) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	p, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	if err = g.HashStart(); err != nil {
		t.Fatal(err)
	}
	truth := func(table uint8) GiaLit {
		result := GiaConst0
		for assignment := 0; assignment < 4; assignment++ {
			if table&(1<<assignment) == 0 {
				continue
			}
			term, hashErr := g.HashAnd(p.NotIf(assignment&1 == 0), r.NotIf(assignment&2 == 0))
			if hashErr != nil {
				t.Fatal(hashErr)
			}
			result, hashErr = g.HashOr(result, term)
			if hashErr != nil {
				t.Fatal(hashErr)
			}
		}
		return result
	}
	output, next := truth(outputTable&15), truth(nextTable&15)
	g.HashStop()
	if _, err = g.AppendCO(output); err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(next); err != nil {
		t.Fatal(err)
	}
	if err = g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	return g
}

func cexDepthTruth(table uint8, primary, state bool) bool {
	index := 0
	if primary {
		index |= 1
	}
	if state {
		index |= 2
	}
	return table&(1<<index) != 0
}

func FuzzCounterexampleDepthConeSoundness(f *testing.F) {
	for _, seed := range [][5]uint8{{8, 10, 0, 1, 1}, {6, 9, 1, 5, 0}, {14, 3, 1, 7, 1}, {1, 12, 0, 2, 0}} {
		f.Add(seed[0], seed[1], seed[2], seed[3], seed[4])
	}
	f.Fuzz(func(t *testing.T, outputTable, nextTable, initial, primaryBits, pruneByte uint8) {
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
		states, _, err := g.CounterexampleInnerStates(trace, CounterexampleToolOptions{})
		if err != nil {
			t.Fatal(err)
		}
		for start := 0; start < trace.Frames(); start++ {
			cone, buildErr := g.BuildCounterexampleDepthCone(states, start, pruneByte&1 != 0, CounterexampleToolOptions{})
			if buildErr != nil {
				t.Fatal(start, buildErr)
			}
			observed, _ := states.Input(start, 1)
			observedOutput, evalErr := cone.Evaluate([]bool{observed})
			if evalErr != nil || !observedOutput[0] {
				t.Fatal("observed state rejected", start, observed, observedOutput, evalErr)
			}
			for _, candidate := range []bool{false, true} {
				coneOutput, evalErr := cone.Evaluate([]bool{candidate})
				if evalErr != nil {
					t.Fatal(evalErr)
				}
				candidateState, finalOutput := candidate, false
				for frame := start; frame < trace.Frames(); frame++ {
					primary, _ := trace.Input(frame, 0)
					finalOutput = cexDepthTruth(outputTable, primary, candidateState)
					candidateState = cexDepthTruth(nextTable, primary, candidateState)
				}
				if coneOutput[0] && !finalOutput {
					t.Fatal("unsound cone witness", outputTable&15, nextTable&15, start, candidate)
				}
			}
		}
	})
}
