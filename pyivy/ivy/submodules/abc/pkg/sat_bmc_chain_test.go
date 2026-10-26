package abc

// Go-only model, ownership and fuzz tests for src/sat/bmc/bmcChain.c.

import (
	"errors"
	"slices"
	"testing"
)

func chainSequentialGraph(t testing.TB) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	r, err := g.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(r.Not()); err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(r); err != nil {
		t.Fatal(err)
	}
	if _, err = g.AppendCO(GiaConst1); err != nil {
		t.Fatal(err)
	}
	if err = g.SetRegisterCount(1); err != nil {
		t.Fatal(err)
	}
	if err = g.Check(); err != nil {
		t.Fatal(err)
	}
	return g
}

func TestChainMoveUsesStartOfFailingFrame(t *testing.T) {
	g := chainSequentialGraph(t)
	before := slices.Clone(g.objects)
	path, err := NewCounterexample(1, 0, 2)
	if err != nil {
		t.Fatal(err)
	}
	path.Property = 1
	moved, reached, err := g.MoveChainCounterexample(path, ChainOptions{Frames: 2})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(reached, []bool{true}) || !slices.Equal(before, g.objects) {
		t.Fatal("reached state/source ownership", reached)
	}
	outputs, next, err := moved.Step(nil, []bool{false})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(outputs, []bool{false, true}) || !slices.Equal(next, []bool{false}) {
		t.Fatal("phase-recoded transition", outputs, next)
	}
	// The state after the failing frame is also true in this fixture. Change the
	// transition to false to distinguish the required start-of-frame state.
	registerLiteral, literalErr := MakeGiaLit(g.cis[0], true)
	if literalErr != nil {
		t.Fatal(literalErr)
	}
	if err = g.PatchCODriver(2, registerLiteral); err != nil {
		t.Fatal(err)
	}
	moved, reached, err = g.MoveChainCounterexample(path, ChainOptions{Frames: 2})
	if err != nil || !slices.Equal(reached, []bool{true}) {
		t.Fatal(err, reached)
	}
	_, next, err = moved.Step(nil, []bool{false})
	if err != nil || !slices.Equal(next, []bool{true}) {
		t.Fatal("recoded next state", next, err)
	}
}

func TestChainPropagateAndFindFailedOutputs(t *testing.T) {
	g := NewGiaManager()
	p, _ := g.AppendCI()
	q, _ := g.AppendCI()
	r, _ := g.AppendCI()
	pq, _ := g.AppendAnd(p, q)
	g.AppendCO(p)
	g.AppendCO(p.Not())
	g.AppendCO(pq)
	g.AppendCO(r)
	g.AppendCO(GiaConst0)
	g.AppendCO(q.Not())
	g.SetRegisterCount(1)
	before := slices.Clone(g.objects)
	initial, err := g.PropagateChainInitialState(ChainOptions{Frames: 1})
	if err != nil {
		t.Fatal(err)
	}
	if initial.RegisterCount() != 0 || initial.PrimaryInputCount() != 2 || initial.PrimaryOutputCount() != 5 || !slices.Equal(before, g.objects) {
		t.Fatal("initial-state graph shape/source")
	}
	for assignment := 0; assignment < 4; assignment++ {
		inputs := []bool{assignment&1 != 0, assignment&2 != 0}
		got, err := initial.Evaluate(inputs)
		if err != nil {
			t.Fatal(err)
		}
		want := []bool{inputs[0], !inputs[0], inputs[0] && inputs[1], false, false}
		if !slices.Equal(got, want) {
			t.Fatal(assignment, got, want)
		}
	}
	search, err := g.FindChainFailedOutputs(ChainOptions{Frames: 1})
	if err != nil {
		t.Fatal(err)
	}
	if search.Status != SatSatisfiable || !slices.Equal(search.Outputs, []int{0, 1, 2}) || len(search.Counterexamples) != 3 {
		t.Fatal(search.Status, search.Reason, search.Outputs)
	}
	for index, trace := range search.Counterexamples {
		if trace.Property != search.Outputs[index] || trace.Registers() != 1 || trace.Inputs() != 2 || trace.Frames() != 1 {
			t.Fatal(index, trace)
		}
		inputs := make([]bool, 2)
		for input := range inputs {
			inputs[input], _ = trace.Input(0, input)
		}
		outputs, _, err := g.Step(inputs, []bool{false})
		if err != nil || !outputs[trace.Property] {
			t.Fatal("unreplayed state witness", index, err)
		}
	}
}

func TestChainCleanupAtomicAndExact(t *testing.T) {
	g := chainSequentialGraph(t)
	before := slices.Clone(g.objects)
	clean, err := g.CleanupChainOutputs([]int{0}, ChainOptions{Frames: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(before, g.objects) || clean.RegisterCount() != 1 || clean.PrimaryOutputCount() != 2 {
		t.Fatal("cleanup ownership/shape")
	}
	count, err := clean.CountChainNonzeroOutputs()
	if err != nil || count != 1 {
		t.Fatal(count, err)
	}
	for state := 0; state < 2; state++ {
		outputs, next, err := clean.Step(nil, []bool{state != 0})
		if err != nil || outputs[0] || outputs[1] != (state != 0) || !next[0] {
			t.Fatal(state, outputs, next, err)
		}
	}
	for _, outputs := range [][]int{{0, 0}, {-1}, {2}} {
		if graph, err := g.CleanupChainOutputs(outputs, ChainOptions{Frames: 1}); graph != nil || !errors.Is(err, ErrDimension) {
			t.Fatal(outputs, err)
		}
	}
	if graph, err := clean.CleanupChainOutputs([]int{0}, ChainOptions{Frames: 1}); graph != nil || !errors.Is(err, ErrDimension) {
		t.Fatal("already false", err)
	}
}

func TestChainBMCCompleteTwoStateSequence(t *testing.T) {
	g := chainSequentialGraph(t)
	before := slices.Clone(g.objects)
	result, err := g.ChainBMC(ChainOptions{Frames: 3, MaxIterations: 8})
	if err != nil {
		t.Fatal(err)
	}
	if result.Reason != ChainAllFailed || result.Depth != 1 || result.Failed != 2 || result.Undecided != 0 || len(result.Segments) != 2 {
		t.Fatal("chain summary", result.Reason, result.Depth, result.Failed, result.Undecided, len(result.Segments))
	}
	if !slices.Equal(result.Segments[0].FailedOutputs, []int{0}) || result.Segments[0].Path.LastFrame() != 0 || !slices.Equal(result.Segments[0].Reached, []bool{false}) {
		t.Fatal("first segment", result.Segments[0])
	}
	if !slices.Equal(result.Segments[1].FailedOutputs, []int{1}) || result.Segments[1].Path.LastFrame() != 1 || !slices.Equal(result.Segments[1].Reached, []bool{true}) {
		t.Fatal("second segment", result.Segments[1])
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("chain mutated source")
	}
	for state := 0; state < 2; state++ {
		outputs, _, err := result.FinalGraph.Step(nil, []bool{state != 0})
		if err != nil || outputs[0] || outputs[1] {
			t.Fatal("final properties", state, outputs, err)
		}
	}
}

func TestChainBMCBoundsCancellationAndValidation(t *testing.T) {
	g := chainSequentialGraph(t)
	result, err := g.ChainBMC(ChainOptions{Frames: 1, MaxIterations: 4})
	if err != nil || result.Reason != ChainBoundReached || result.Failed != 1 || result.Undecided != 1 || len(result.Segments) != 1 {
		t.Fatal("finite bound", result.Reason, result.Failed, result.Undecided, len(result.Segments), err)
	}
	result, err = g.ChainBMC(ChainOptions{Frames: 3, MaxIterations: 1})
	if err != nil || result.Reason != ChainIterationLimit || len(result.Segments) != 1 {
		t.Fatal("iteration limit", result.Reason, len(result.Segments), err)
	}
	result, err = g.ChainBMC(ChainOptions{Frames: 3, SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || result.Reason != ChainCancelled || result.FinalGraph == nil {
		t.Fatal("cancel", result.Reason, err)
	}
	for _, options := range []ChainOptions{
		{Frames: 0}, {Frames: 1, MaxIterations: -1}, {Frames: 1, MaxObjects: -1},
		{Frames: 1, MaxMappingEntries: -1}, {Frames: 1, MaxVariables: -1},
	} {
		if _, err := g.ChainBMC(options); !errors.Is(err, ErrDimension) {
			t.Fatal(options, err)
		}
	}
	if _, _, err = g.MoveChainCounterexample(resultCounterexample(t, 1, 0, 1, 0), ChainOptions{Frames: 1, SAT: SatLimits{Stop: func() bool { return true }}}); !errors.Is(err, ErrCancelled) {
		t.Fatal("move cancellation", err)
	}
	if _, err = g.PropagateChainInitialState(ChainOptions{Frames: 1, MaxMappingEntries: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("map limit", err)
	}
	if _, err = g.PropagateChainInitialState(ChainOptions{Frames: 1, MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("propagation object limit", err)
	}
	if _, _, err = g.MoveChainCounterexample(resultCounterexample(t, 1, 0, 1, 0), ChainOptions{Frames: 1, MaxObjects: g.ObjectCount() - 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("move object limit", err)
	}
	if _, _, err = g.MoveChainCounterexample(resultCounterexample(t, 1, 0, 1, 0), ChainOptions{Frames: 1, MaxMappingEntries: g.ObjectCount() - 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("move map limit", err)
	}
	if _, err = g.CleanupChainOutputs([]int{0}, ChainOptions{Frames: 1, MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("object limit", err)
	}
	if _, err = g.CleanupChainOutputs([]int{0}, ChainOptions{Frames: 1, MaxMappingEntries: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("cleanup map limit", err)
	}
	withInput := NewGiaManager()
	input, _ := withInput.AppendCI()
	withInput.AppendCO(input)
	if _, err = withInput.FindChainFailedOutputs(ChainOptions{Frames: 1, MaxVariables: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal("SAT variable limit", err)
	}
	constraints := g.Clone()
	if err = constraints.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	if _, err = constraints.ChainBMC(ChainOptions{Frames: 1}); !errors.Is(err, ErrUnsupported) {
		t.Fatal("constraints", err)
	}
}

func TestChainOutputSearchPreservesSATUnknown(t *testing.T) {
	g := NewGiaManager()
	p, _ := g.AppendCI()
	q, _ := g.AppendCI()
	first, _ := g.AppendAnd(p, q)
	second, _ := g.AppendAnd(p, q)
	impossible, _ := g.AppendXorReal(first, second)
	g.AppendCO(impossible)
	calls := 0
	if _, err := g.FindChainFailedOutputs(ChainOptions{Frames: 1, SAT: SatLimits{Stop: func() bool { calls++; return false }}}); err != nil {
		t.Fatal(err)
	}
	if calls == 0 {
		t.Fatal("cancellation callback was not polled")
	}
	limit := calls
	calls = 0
	search, err := g.FindChainFailedOutputs(ChainOptions{Frames: 1, SAT: SatLimits{Stop: func() bool { calls++; return calls >= limit }}})
	if err != nil {
		t.Fatal(err)
	}
	if search.Status != SatUnknown || search.Reason != SatCancelled || len(search.Outputs) != 0 || len(search.Counterexamples) != 0 {
		t.Fatal(search.Status, search.Reason, search.Outputs)
	}
}

func resultCounterexample(t testing.TB, registers, inputs, frames, property int) *Counterexample {
	t.Helper()
	c, err := NewCounterexample(registers, inputs, frames)
	if err != nil {
		t.Fatal(err)
	}
	c.Property = property
	return c
}

func FuzzChainFailedOutputsAgainstEnumeration(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4, 5})
	f.Add([]byte{255, 17, 91, 3, 8, 2, 1})
	f.Fuzz(func(t *testing.T, data []byte) {
		g := NewGiaManager()
		if err := g.HashStart(); err != nil {
			t.Fatal(err)
		}
		a, _ := g.AppendCI()
		b, _ := g.AppendCI()
		pool := []GiaLit{GiaConst0, GiaConst1, a, a.Not(), b, b.Not()}
		for i, value := range data {
			if i == 12 {
				break
			}
			left := pool[int(value)%len(pool)]
			right := pool[(int(value)/7+i+1)%len(pool)]
			var node GiaLit
			var err error
			if value&1 == 0 {
				node, err = g.HashAnd(left, right)
			} else {
				node, err = g.HashXor(left, right)
			}
			if err != nil {
				t.Fatal(err)
			}
			pool = append(pool, node)
		}
		for output := 0; output < 5; output++ {
			index := output
			if len(data) > 0 {
				index = int(data[output%len(data)]) + output
			}
			if _, err := g.AppendCO(pool[index%len(pool)]); err != nil {
				t.Fatal(err)
			}
		}
		g.HashStop()
		search, err := g.FindChainFailedOutputs(ChainOptions{Frames: 1, MaxObjects: 1024, MaxMappingEntries: 1024, MaxVariables: 1024})
		if err != nil {
			t.Fatal(err)
		}
		want := make([]int, 0, 5)
		for output := 0; output < 5; output++ {
			possible := false
			for assignment := 0; assignment < 4; assignment++ {
				values, evalErr := g.Evaluate([]bool{assignment&1 != 0, assignment&2 != 0})
				if evalErr != nil {
					t.Fatal(evalErr)
				}
				possible = possible || values[output]
			}
			if possible {
				want = append(want, output)
			}
		}
		if search.Status != SatSatisfiable || !slices.Equal(search.Outputs, want) {
			t.Fatal(search.Status, search.Outputs, want)
		}
	})
}
