package abc

// Go-only semantic, exhaustive and fuzz tests for src/sat/bmc/bmcBmcAnd.c.

import (
	"errors"
	"slices"
	"testing"
)

func giaAndDelayGraph(t *testing.T, delay int) *GiaManager {
	t.Helper()
	graph := NewGiaManager()
	input, err := graph.AppendCI()
	if err != nil {
		t.Fatal(err)
	}
	registers := make([]GiaLit, delay)
	for index := range registers {
		registers[index], err = graph.AppendCI()
		if err != nil {
			t.Fatal(err)
		}
	}
	output := input
	if delay != 0 {
		output = registers[delay-1]
	}
	if _, err = graph.AppendCO(output); err != nil {
		t.Fatal(err)
	}
	for index := range registers {
		next := input
		if index != 0 {
			next = registers[index-1]
		}
		if _, err = graph.AppendCO(next); err != nil {
			t.Fatal(err)
		}
	}
	if err = graph.SetRegisterCount(delay); err != nil {
		t.Fatal(err)
	}
	return graph
}

func TestGiaAndTernaryUnrollSemantics(t *testing.T) {
	graph := giaAndDelayGraph(t, 1)
	before := graph.Clone()
	unrolling, err := graph.AndTernaryUnroll(GiaAndUnrollOptions{Frames: 3})
	if err != nil {
		t.Fatal(err)
	}
	if unrolling.Frames() != 3 || unrolling.FirstUnknownFrame() != 1 {
		t.Fatal(unrolling.Frames(), unrolling.FirstUnknownFrame())
	}
	inputs := unrolling.Inputs()
	if len(inputs) != 2 || inputs[0].Frame != 0 || inputs[1].Frame != 1 || inputs[0].Input != 0 || inputs[1].Input != 0 {
		t.Fatal(inputs)
	}
	for assignment := 0; assignment < 1<<len(inputs); assignment++ {
		sparseInputs := make([]bool, len(inputs))
		originalInputs := make([]bool, 3)
		for index, input := range inputs {
			sparseInputs[index] = assignment>>index&1 != 0
			originalInputs[input.Frame] = sparseInputs[index]
		}
		got, evalErr := unrolling.Graph().Evaluate(sparseInputs)
		if evalErr != nil {
			t.Fatal(evalErr)
		}
		state := []bool{false}
		for frame := 0; frame < 3; frame++ {
			outputs, next, stepErr := graph.Step([]bool{originalInputs[frame]}, state)
			if stepErr != nil {
				t.Fatal(stepErr)
			}
			active, activeErr := unrolling.FrameActive(frame)
			if activeErr != nil {
				t.Fatal(activeErr)
			}
			want := active && outputs[0]
			if got[frame] != want {
				t.Fatalf("assignment=%d frame=%d got=%v want=%v active=%v", assignment, frame, got[frame], want, active)
			}
			state = next
		}
	}
	values, err := unrolling.FrameValues(0)
	if err != nil || values[graph.Outputs()[0]] != GiaTernaryZero {
		t.Fatal(values, err)
	}
	values[0] = GiaTernaryOne
	again, _ := unrolling.FrameValues(0)
	if again[0] != GiaTernaryZero {
		t.Fatal("frame values alias caller storage")
	}
	inputs[0].Frame = 99
	if unrolling.Inputs()[0].Frame == 99 {
		t.Fatal("input provenance aliases caller storage")
	}
	if !slices.Equal(graph.objects, before.objects) || !slices.Equal(graph.cis, before.cis) || !slices.Equal(graph.cos, before.cos) {
		t.Fatal("sparse unroll mutated source graph")
	}
	if _, err = unrolling.FrameValues(3); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = unrolling.FrameActive(-1); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestGiaAndTernaryAdaptiveAndBackwardCut(t *testing.T) {
	shift := giaAndDelayGraph(t, 2)
	unrolling, err := shift.AndTernaryUnroll(GiaAndUnrollOptions{AdditionalFrames: 2, MaxFrames: 8})
	if err != nil {
		t.Fatal(err)
	}
	if unrolling.FirstUnknownFrame() != 2 || unrolling.Frames() != 4 {
		t.Fatal(unrolling.FirstUnknownFrame(), unrolling.Frames())
	}
	combinational := giaAndDelayGraph(t, 0)
	cut, err := combinational.AndTernaryUnroll(GiaAndUnrollOptions{Frames: 4})
	if err != nil {
		t.Fatal(err)
	}
	for frame := 0; frame < 3; frame++ {
		active, _ := cut.FrameActive(frame)
		if active {
			t.Fatal("backward traversal should stop before frame", frame)
		}
	}
	active, _ := cut.FrameActive(3)
	if !active || len(cut.Inputs()) != 1 || cut.Inputs()[0].Frame != 3 {
		t.Fatal(active, cut.Inputs())
	}
	safe := cexDepthTruthGraph(t, 0, 0xa)
	capped, err := safe.AndTernaryUnroll(GiaAndUnrollOptions{AdditionalFrames: 2, MaxFrames: 3})
	if err != nil || capped.FirstUnknownFrame() != -1 || capped.Frames() != 3 {
		t.Fatal(capped, err)
	}
}

func TestGiaAndTernaryKnownControllingAndErrors(t *testing.T) {
	graph := NewGiaManager()
	input, _ := graph.AppendCI()
	changing, _ := graph.AppendCI()
	fixed, _ := graph.AppendCI()
	root, _ := graph.AppendAnd(changing, fixed.Not())
	graph.AppendCO(root)
	graph.AppendCO(input)
	graph.AppendCO(GiaConst0)
	graph.SetRegisterCount(2)
	unrolling, err := graph.AndTernaryUnroll(GiaAndUnrollOptions{Frames: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(unrolling.Inputs()) != 1 || unrolling.Inputs()[0].Frame != 0 {
		t.Fatal(unrolling.Inputs())
	}
	constant := NewGiaManager()
	constant.AppendCO(GiaConst1)
	if _, err = constant.AndTernaryUnroll(GiaAndUnrollOptions{Frames: 1}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	if _, err = graph.AndTernaryUnroll(GiaAndUnrollOptions{}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
	if _, err = graph.AndTernaryUnroll(GiaAndUnrollOptions{Frames: 2, MaxFrames: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = graph.AndTernaryUnroll(GiaAndUnrollOptions{Frames: 2, MaxMappingEntries: graph.ObjectCount()}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = graph.AndTernaryUnroll(GiaAndUnrollOptions{Frames: 2, MaxObjects: 1}); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if _, err = graph.AndTernaryUnroll(GiaAndUnrollOptions{Frames: 2, Stop: func() bool { return true }}); !errors.Is(err, ErrCancelled) {
		t.Fatal(err)
	}
	extended := NewGiaManager()
	a, _ := extended.AppendCI()
	b, _ := extended.AppendCI()
	x, _ := extended.AppendXorReal(a, b)
	extended.AppendCO(x)
	if _, err = extended.AndTernaryUnroll(GiaAndUnrollOptions{Frames: 1}); !errors.Is(err, ErrUnsupported) {
		t.Fatal(err)
	}
	constrained := giaAndDelayGraph(t, 1)
	if err = constrained.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	if _, err = constrained.AndTernaryUnroll(GiaAndUnrollOptions{Frames: 1}); !errors.Is(err, ErrDimension) {
		t.Fatal(err)
	}
}

func TestGiaAndBoundedCheckWitnessAndExclusion(t *testing.T) {
	graph := giaAndDelayGraph(t, 1)
	result, err := graph.AndBoundedCheck(GiaAndBMCOptions{Unroll: GiaAndUnrollOptions{Frames: 3}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != BmcCounterexample || result.Frame != 1 || result.Property != 0 || result.Counterexample == nil {
		t.Fatal(result.Status, result.Frame, result.Property, result.Counterexample)
	}
	if value, _ := result.Counterexample.Input(0, 0); !value {
		t.Fatal("witness does not drive delayed property")
	}
	safe := cexDepthTruthGraph(t, 0, 0xa)
	result, err = safe.AndBoundedCheck(GiaAndBMCOptions{Unroll: GiaAndUnrollOptions{Frames: 3}})
	if err != nil || result.Status != BmcNoCounterexample || result.CheckedFrames != 3 {
		t.Fatal(result, err)
	}
	result, err = graph.AndBoundedCheck(GiaAndBMCOptions{Unroll: GiaAndUnrollOptions{Frames: 3}, SAT: SatLimits{Conflicts: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != BmcCounterexample && result.Status != BmcUnknown {
		t.Fatal(result.Status)
	}
	result, err = graph.AndBoundedCheck(GiaAndBMCOptions{Unroll: GiaAndUnrollOptions{Frames: 3}, MaxVariables: 1})
	if err != nil || result.Reason != BmcResourceLimit {
		t.Fatal(result.Reason, err)
	}
	result, err = graph.AndBoundedCheck(GiaAndBMCOptions{Unroll: GiaAndUnrollOptions{Frames: 3}, SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || result.Reason != BmcCancelled || result.SatReason != SatCancelled {
		t.Fatal(result.Reason, result.SatReason, err)
	}
}

func TestGiaAndBoundedCheckOutputOrder(t *testing.T) {
	graph := NewGiaManager()
	input, _ := graph.AppendCI()
	graph.AppendCO(GiaConst0)
	graph.AppendCO(input)
	result, err := graph.AndBoundedCheck(GiaAndBMCOptions{Unroll: GiaAndUnrollOptions{Frames: 2}})
	if err != nil || result.Status != BmcCounterexample || result.Frame != 1 || result.Property != 1 {
		t.Fatal(result.Status, result.Frame, result.Property, err)
	}
	first := NewGiaManager()
	input, _ = first.AppendCI()
	first.AppendCO(input)
	first.AppendCO(input.Not())
	result, err = first.AndBoundedCheck(GiaAndBMCOptions{Unroll: GiaAndUnrollOptions{Frames: 2}})
	if err != nil || result.Status != BmcCounterexample || result.Frame != 1 || result.Property != 0 {
		t.Fatal(result.Status, result.Frame, result.Property, err)
	}
}

func TestGiaAndTernaryDeepCone(t *testing.T) {
	graph := NewGiaManager()
	a, _ := graph.AppendCI()
	b, _ := graph.AppendCI()
	root := a
	for index := 0; index < 10000; index++ {
		var err error
		root, err = graph.AppendAnd(root, b.NotIf(index&1 != 0))
		if err != nil {
			t.Fatal(index, err)
		}
	}
	graph.AppendCO(root)
	unrolling, err := graph.AndTernaryUnroll(GiaAndUnrollOptions{Frames: 1, MaxObjects: 20050, MaxMappingEntries: 20050})
	if err != nil {
		t.Fatal(err)
	}
	if err = unrolling.Graph().Check(); err != nil {
		t.Fatal(err)
	}
}

func FuzzGiaAndBoundedCheck(f *testing.F) {
	f.Add(uint8(0xa), uint8(0xa), uint8(3))
	f.Add(uint8(0x8), uint8(0xa), uint8(4))
	f.Add(uint8(0x6), uint8(0xc), uint8(2))
	f.Fuzz(func(t *testing.T, outputTable, nextTable, frameByte uint8) {
		frames := int(frameByte%4) + 1
		graph := cexDepthTruthGraph(t, outputTable&15, nextTable&15)
		if _, unrollErr := graph.AndTernaryUnroll(GiaAndUnrollOptions{Frames: frames}); unrollErr != nil {
			if errors.Is(unrollErr, ErrUnsupported) {
				return
			}
			t.Fatalf("unroll: %v", unrollErr)
		}
		result, err := graph.AndBoundedCheck(GiaAndBMCOptions{Unroll: GiaAndUnrollOptions{Frames: frames}, MaxVariables: 10000})
		if errors.Is(err, ErrUnsupported) {
			return
		}
		if err != nil {
			t.Fatalf("bounded check: %v", err)
		}
		reachable := false
		for sequence := 0; sequence < 1<<frames; sequence++ {
			state := []bool{false}
			for frame := 0; frame < frames; frame++ {
				outputs, next, stepErr := graph.Step([]bool{sequence>>frame&1 != 0}, state)
				if stepErr != nil {
					t.Fatal(stepErr)
				}
				if outputs[0] {
					reachable = true
				}
				state = next
			}
		}
		if (result.Status == BmcCounterexample) != reachable {
			t.Fatalf("tables=%x/%x frames=%d status=%v reachable=%v", outputTable&15, nextTable&15, frames, result.Status, reachable)
		}
		if result.Status == BmcCounterexample && result.Counterexample == nil {
			t.Fatal("missing witness")
		}
	})
}
