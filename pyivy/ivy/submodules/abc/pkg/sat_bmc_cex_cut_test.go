package abc

// Go-only semantic, limit, and fuzz tests for src/sat/bmc/bmcCexCut.c.

import (
	"errors"
	"slices"
	"testing"
)

func cexCutTwoRegisterGraph(t testing.TB) *GiaManager {
	t.Helper()
	g := NewGiaManager()
	if _, err := g.AppendCI(); err != nil {
		t.Fatal(err)
	}
	r0, _ := g.AppendCI()
	r1, _ := g.AppendCI()
	both, err := g.AppendAnd(r0, r1)
	if err != nil {
		t.Fatal(err)
	}
	g.AppendCO(both.Not())
	g.AppendCO(r0)
	g.AppendCO(r1)
	g.SetRegisterCount(2)
	return g
}

func cexCutTrace(t testing.TB) *Counterexample {
	t.Helper()
	trace, err := NewCounterexample(2, 1, 3)
	if err != nil {
		t.Fatal(err)
	}
	trace.Property = 0
	return trace
}

func TestCounterexampleStateCutAnnotation(t *testing.T) {
	g := cexCutTwoRegisterGraph(t)
	trace := cexCutTrace(t)
	annotation, err := g.AnnotateCounterexampleStateCut(trace, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if annotation.Frames() != 3 || annotation.Objects() != g.ObjectCount() || annotation.Property() != 0 {
		t.Fatal("annotation dimensions")
	}
	for frame := 0; frame < trace.Frames(); frame++ {
		for register, ro := range g.cis[g.PrimaryInputCount():] {
			value, valueErr := annotation.Value(frame, ro)
			justified, justifyErr := annotation.Justified(frame, ro)
			wantJustified := register == 0
			if valueErr != nil || justifyErr != nil || value || justified != wantJustified {
				t.Fatal(frame, register, value, justified, valueErr, justifyErr)
			}
		}
	}
	pi := g.cis[0]
	for frame := 0; frame < trace.Frames(); frame++ {
		fixed, _ := annotation.Justified(frame, pi)
		if !fixed {
			t.Fatal("fixed PI omitted", frame)
		}
	}
	clone := annotation.Clone()
	clone.values[0] = !clone.values[0]
	value, _ := annotation.Value(0, 0)
	if value {
		t.Fatal("annotation clone aliases source")
	}
}

func TestCounterexampleTargetStateModes(t *testing.T) {
	g := cexCutTwoRegisterGraph(t)
	g.Name, g.Spec = "cut", "cut.aig"
	trace := cexCutTrace(t)
	before := slices.Clone(g.objects)
	for _, mode := range []CexTargetStateMode{CexTargetJustifiedCube, CexTargetTerminalCone, CexTargetAllFrameCones} {
		result, err := g.CounterexampleTargetStates(trace, 0, 1, CexTargetStateOptions{
			Mode: mode, CombinationalOnly: true,
		})
		if err != nil {
			t.Fatal(mode, err)
		}
		wantFrames := []int{1}
		if mode == CexTargetAllFrameCones {
			wantFrames = []int{1, 2}
		}
		if !slices.Equal(result.InitialState, []bool{false, false}) || !slices.Equal(result.Frames, wantFrames) || result.Property.InputCount() != 2 || result.Property.OutputCount() != len(wantFrames) || result.Graph != nil {
			t.Fatal("result shape", mode, result.InitialState, result.Frames)
		}
		for assignment := 0; assignment < 4; assignment++ {
			r0, r1 := assignment&1 != 0, assignment&2 != 0
			outputs, evalErr := result.Property.Evaluate([]bool{r0, r1})
			if evalErr != nil {
				t.Fatal(evalErr)
			}
			want := !r0
			if mode != CexTargetJustifiedCube {
				want = !(r0 && r1)
			}
			for _, output := range outputs {
				if output != want {
					t.Fatal(mode, assignment, outputs, want)
				}
			}
		}
	}
	if !slices.Equal(before, g.objects) {
		t.Fatal("source graph changed")
	}
}

func TestCounterexampleTargetStateSequentialAndAig(t *testing.T) {
	g := cexDepthSelfLoopGraph(t)
	g.Name, g.Spec = "loop", "loop.aig"
	trace, _ := NewCounterexample(1, 1, 3)
	trace.Property = 0
	trace.SetInitial(0, true)
	result, err := g.CounterexampleTargetStates(trace, 0, 1, CexTargetStateOptions{Mode: CexTargetTerminalCone})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(result.InitialState, []bool{true}) || result.Graph == nil || result.Graph.Name != g.Name || result.Graph.Spec != g.Spec || result.Graph.RegisterCount() != 1 || result.Graph.PrimaryOutputCount() != 1 {
		t.Fatal("sequential shape")
	}
	for assignment := 0; assignment < 4; assignment++ {
		p, state := assignment&1 != 0, assignment&2 != 0
		outputs, evalErr := result.Graph.Evaluate([]bool{p, state})
		want := []bool{!state, state}
		if evalErr != nil || !slices.Equal(outputs, want) {
			t.Fatal(assignment, outputs, want, evalErr)
		}
	}
	aigSource, _, err := g.ToAig()
	if err != nil {
		t.Fatal(err)
	}
	aigResult, initial, err := aigSource.CounterexampleTargetStates(trace, 0, 1, CexTargetStateOptions{Mode: CexTargetTerminalCone})
	if err != nil || !slices.Equal(initial, []bool{true}) || aigResult.RegisterCount() != 1 {
		t.Fatal("AIG wrapper", initial, err)
	}
	for assignment := 0; assignment < 4; assignment++ {
		p, state := assignment&1 != 0, assignment&2 != 0
		outputs, evalErr := aigResult.Evaluate([]bool{p, state})
		want := []bool{!state, state}
		if evalErr != nil || !slices.Equal(outputs, want) {
			t.Fatal(assignment, outputs, want, evalErr)
		}
	}
}

func TestCounterexampleTargetStateFullyImpliedAndFailures(t *testing.T) {
	driven := cexToolSequentialGraph(t)
	trace := cexToolSequentialTrace(t)
	if _, err := driven.AnnotateCounterexampleStateCut(trace, CounterexampleToolOptions{}); !errors.Is(err, ErrInvariant) {
		t.Fatal("cube source precondition", err)
	}
	terminal, err := driven.CounterexampleTargetStates(trace, 0, 1, CexTargetStateOptions{
		Mode: CexTargetTerminalCone, CombinationalOnly: true,
	})
	if err != nil {
		t.Fatal("terminal mode accepts implied failure", err)
	}
	for _, state := range []bool{false, true} {
		outputs, evalErr := terminal.Property.Evaluate([]bool{state})
		if evalErr != nil || outputs[0] != state {
			t.Fatal(state, outputs, evalErr)
		}
	}

	g := cexCutTwoRegisterGraph(t)
	valid := cexCutTrace(t)
	for _, test := range []struct {
		name string
		call func() error
		want error
	}{
		{"equal frames", func() error { _, e := g.CounterexampleTargetStates(valid, 1, 1, CexTargetStateOptions{}); return e }, ErrDimension},
		{"negative frame", func() error { _, e := g.CounterexampleTargetStates(valid, -1, 1, CexTargetStateOptions{}); return e }, ErrDimension},
		{"late frame", func() error { _, e := g.CounterexampleTargetStates(valid, 0, 3, CexTargetStateOptions{}); return e }, ErrDimension},
		{"bad mode", func() error {
			_, e := g.CounterexampleTargetStates(valid, 0, 1, CexTargetStateOptions{Mode: 99})
			return e
		}, ErrDimension},
		{"mapping limit", func() error {
			_, e := g.CounterexampleTargetStates(valid, 0, 1, CexTargetStateOptions{CounterexampleToolOptions: CounterexampleToolOptions{MaxMappingEntries: 1}})
			return e
		}, ErrLimit},
		{"object limit", func() error {
			_, e := g.CounterexampleTargetStates(valid, 0, 1, CexTargetStateOptions{CounterexampleToolOptions: CounterexampleToolOptions{MaxObjects: 1}})
			return e
		}, ErrLimit},
		{"cancel", func() error {
			_, e := g.CounterexampleTargetStates(valid, 0, 1, CexTargetStateOptions{CounterexampleToolOptions: CounterexampleToolOptions{Stop: func() bool { return true }}})
			return e
		}, ErrCancelled},
	} {
		t.Run(test.name, func(t *testing.T) {
			if got := test.call(); !errors.Is(got, test.want) {
				t.Fatal(got, test.want)
			}
		})
	}
	extended := NewGiaManager()
	p, _ := extended.AppendCI()
	r, _ := extended.AppendCI()
	x, _ := extended.AppendXorReal(p, r)
	extended.AppendCO(x)
	extended.AppendCO(r)
	extended.SetRegisterCount(1)
	extendedTrace, _ := NewCounterexample(1, 1, 2)
	extendedTrace.Property = 0
	extendedTrace.SetInitial(0, true)
	if _, err = extended.AnnotateCounterexampleStateCut(extendedTrace, CounterexampleToolOptions{}); !errors.Is(err, ErrUnsupported) {
		t.Fatal("extended recurrence", err)
	}
}

func TestCounterexampleStateCutDeepIterative(t *testing.T) {
	g := NewGiaManager()
	if _, err := g.AppendCI(); err != nil {
		t.Fatal(err)
	}
	r0, _ := g.AppendCI()
	r1, _ := g.AppendCI()
	current, err := g.AppendAnd(r0, r1)
	if err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 10_000; index++ {
		current, err = g.AppendAnd(current, r0)
		if err != nil {
			t.Fatal(index, err)
		}
	}
	g.AppendCO(current)
	g.AppendCO(r0)
	g.AppendCO(r1)
	g.SetRegisterCount(2)
	trace, _ := NewCounterexample(2, 1, 2)
	trace.Property = 0
	trace.SetInitial(0, true)
	trace.SetInitial(1, true)
	annotation, err := g.AnnotateCounterexampleStateCut(trace, CounterexampleToolOptions{})
	if err != nil {
		t.Fatal(err)
	}
	for _, ro := range g.cis[g.PrimaryInputCount():] {
		justified, justifyErr := annotation.Justified(0, ro)
		if justifyErr != nil || !justified {
			t.Fatal(ro, justified, justifyErr)
		}
	}
}

func FuzzCounterexampleTargetStateSoundness(f *testing.F) {
	for _, seed := range [][5]uint8{{8, 10, 1, 1, 0}, {6, 9, 0, 7, 1}, {14, 3, 1, 5, 0}, {1, 12, 0, 2, 1}} {
		f.Add(seed[0], seed[1], seed[2], seed[3], seed[4])
	}
	f.Fuzz(func(t *testing.T, outputTable, nextTable, initial, primaryBits, modeByte uint8) {
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
		mode := CexTargetTerminalCone
		if modeByte&1 == 0 {
			mode = CexTargetJustifiedCube
		}
		result, err := g.CounterexampleTargetStates(trace, 0, 1, CexTargetStateOptions{Mode: mode, CombinationalOnly: true})
		if errors.Is(err, ErrInvariant) {
			return
		}
		if err != nil {
			t.Fatal(err)
		}
		for _, candidate := range []bool{false, true} {
			accepted, evalErr := result.Property.Evaluate([]bool{candidate})
			if evalErr != nil {
				t.Fatal(evalErr)
			}
			candidateState, finalOutput := candidate, false
			for frame := 1; frame < trace.Frames(); frame++ {
				primary, _ := trace.Input(frame, 0)
				finalOutput = cexDepthTruth(outputTable, primary, candidateState)
				candidateState = cexDepthTruth(nextTable, primary, candidateState)
			}
			if accepted[0] && !finalOutput {
				t.Fatal("unsound target state", outputTable&15, nextTable&15, mode, candidate)
			}
		}
	})
}
