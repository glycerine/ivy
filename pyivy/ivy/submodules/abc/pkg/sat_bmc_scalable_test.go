package abc

// Go-only solve-all and fuzz tests for src/sat/bmc/bmcBmc3.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestAigScalableBoundedCheckOrderAndWitness(t *testing.T) {
	graph := aigIntervalTruthGraph(t, 0xc, 0xa, 0xa)
	result, err := graph.ScalableBoundedCheck(AigScalableBMCOptions{Frames: 3, StoreWitness: true})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != BmcCounterexample || result.Frame != 0 || result.Property != 1 || result.Outputs[1].Witness == nil || result.CheckedTargets != 2 {
		t.Fatal(result.Status, result.Frame, result.Property, result.CheckedTargets, result.Outputs)
	}
	if result.Outputs[0].Status != BmcUnknown || result.Outputs[1].Frame != 0 {
		t.Fatal(result.Outputs)
	}
}

func TestAigScalableSolveAllSameModelAndCallbacks(t *testing.T) {
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	graph.CreateCO(input)
	graph.CreateCO(input)
	var failed []int
	result, err := graph.ScalableBoundedCheck(AigScalableBMCOptions{
		Frames: 2, SolveAll: true, StoreWitness: true,
		OnFailure: func(property int, witness *Counterexample) bool {
			failed = append(failed, property)
			if witness == nil || witness.Property != property {
				t.Fatal(property, witness)
			}
			return false
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != BmcCounterexample || result.Failed != 2 || result.CheckedTargets != 1 || !slices.Equal(failed, []int{0}) {
		t.Fatal(result.Status, result.Failed, result.CheckedTargets, failed)
	}
	for property, output := range result.Outputs {
		if output.Status != BmcCounterexample || output.Frame != 0 || output.Witness == nil || output.Witness.Property != property {
			t.Fatal(property, output)
		}
	}
}

func TestAigScalableSolveAllOppositeOutputs(t *testing.T) {
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	graph.CreateCO(input)
	graph.CreateCO(input.Not())
	var failed []int
	result, err := graph.ScalableBoundedCheck(AigScalableBMCOptions{Frames: 1, SolveAll: true, OnFailure: func(property int, witness *Counterexample) bool {
		failed = append(failed, property)
		if witness != nil {
			t.Fatal("witness supplied when StoreWitness is false")
		}
		return false
	}})
	if err != nil || result.Failed != 2 || result.CheckedTargets != 2 || !slices.Equal(failed, []int{0, 1}) {
		t.Fatal(result, failed, err)
	}
}

func TestAigScalableProgressStartAndCompletion(t *testing.T) {
	safe := aigIntervalTruthGraph(t, 0, 0, 0xa)
	var progress [][2]int
	result, err := safe.ScalableBoundedCheck(AigScalableBMCOptions{Frames: 4, SolveAll: true, OnProgress: func(property, depth int) {
		progress = append(progress, [2]int{property, depth})
	}})
	if err != nil || result.Status != BmcNoCounterexample || !result.Complete || result.CheckedFrames != 2 || len(progress) != 4 {
		t.Fatal(result.Status, result.Complete, result.CheckedFrames, progress, err)
	}
	for property, output := range result.Outputs {
		if output.Status != BmcNoCounterexample || output.Frame != 1 {
			t.Fatal(property, output)
		}
	}
	combinational := NewAigManager()
	input, _ := combinational.CreateCI()
	combinational.CreateCO(input)
	result, err = combinational.ScalableBoundedCheck(AigScalableBMCOptions{StartFrame: 2, Frames: 3, FramesJump: 1})
	if err != nil || result.Status != BmcCounterexample || result.Frame != 2 || result.SkippedFrames != 2 {
		t.Fatal(result.Status, result.Frame, result.SkippedFrames, err)
	}
}

func TestAigScalableConstraintsAndStopCallback(t *testing.T) {
	graph := NewAigManager()
	input, _ := graph.CreateCI()
	graph.CreateCO(AigConst1.Not())
	graph.CreateCO(input)
	if err := graph.SetConstraintCount(1); err != nil {
		t.Fatal(err)
	}
	result, err := graph.ScalableBoundedCheck(AigScalableBMCOptions{Frames: 1})
	if err != nil || result.Status != BmcCounterexample || result.Property != 1 {
		t.Fatal(result.Status, result.Property, err)
	}
	result, err = graph.ScalableBoundedCheck(AigScalableBMCOptions{Frames: 1, SolveAll: true, OnFailure: func(int, *Counterexample) bool { return true }})
	if err != nil || result.Status != BmcCounterexample || result.Failed != 1 {
		t.Fatal(result, err)
	}
}

func TestAigScalableLimitsCancellationAndJump(t *testing.T) {
	graph := aigIntervalTruthGraph(t, 0x8, 0, 0xa)
	for _, options := range []AigScalableBMCOptions{
		{Frames: 0},
		{Frames: 2, MaxObjects: 1},
		{Frames: 2, MaxMappingEntries: 1},
		{Frames: 2, MaxVariables: 1},
		{Frames: 2, MaxClauses: 1},
		{Frames: 2, MaxLiterals: 1},
	} {
		result, err := graph.ScalableBoundedCheck(options)
		if options.Frames == 0 {
			if !errors.Is(err, ErrDimension) {
				t.Fatal(options, err)
			}
		} else if err != nil || result.Reason != BmcResourceLimit {
			t.Fatal(options, result.Reason, err)
		}
	}
	result, err := graph.ScalableBoundedCheck(AigScalableBMCOptions{Frames: 2, SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || result.Reason != BmcCancelled || result.SatReason != SatCancelled {
		t.Fatal(result.Reason, result.SatReason, err)
	}
	jumpGraph := NewAigManager()
	root, _ := jumpGraph.CreateCI()
	for index := 0; index < 8; index++ {
		input, _ := jumpGraph.CreateCI()
		root, _ = jumpGraph.And(root, input)
	}
	jumpGraph.CreateCO(root)
	result, err = jumpGraph.ScalableBoundedCheck(AigScalableBMCOptions{Frames: 3, FramesJump: 2, SAT: SatLimits{Propagations: 1}, JumpSAT: SatLimits{}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != BmcCounterexample || result.Frame != 2 || result.SkippedFrames != 1 {
		t.Fatal(result.Status, result.Frame, result.SkippedFrames)
	}
	result, err = jumpGraph.ScalableBoundedCheck(AigScalableBMCOptions{Frames: 1, SolveAll: true, DropUnknownOutputs: true, SAT: SatLimits{Propagations: 1}})
	if err != nil || result.Status != BmcUnknown || result.Reason != BmcSatLimit || result.Dropped != 1 || !result.Outputs[0].Dropped {
		t.Fatal(result.Status, result.Reason, result.Dropped, result.Outputs, err)
	}
}

func TestAigScalableDeepCone(t *testing.T) {
	graph := NewAigManager()
	a, _ := graph.CreateCI()
	b, _ := graph.CreateCI()
	root := a
	for index := 0; index < 10000; index++ {
		var err error
		root, err = graph.And(root, b.NotIf(index&1 != 0))
		if err != nil {
			t.Fatal(index, err)
		}
	}
	graph.CreateCO(root)
	result, err := graph.ScalableBoundedCheck(AigScalableBMCOptions{Frames: 1, MaxObjects: 30000, MaxMappingEntries: 30000, MaxVariables: 30000})
	if err != nil || result.Status != BmcNoCounterexample && result.Status != BmcCounterexample {
		t.Fatal(result.Status, err)
	}
}

func FuzzAigScalableSolveAll(f *testing.F) {
	f.Add(uint8(0x8), uint8(0x2), uint8(0xa), uint8(3))
	f.Add(uint8(0), uint8(0xf), uint8(0xc), uint8(4))
	f.Fuzz(func(t *testing.T, output0, output1, next, frameByte uint8) {
		frames := int(frameByte%4) + 1
		graph := aigIntervalTruthGraph(t, output0&15, output1&15, next&15)
		result, err := graph.ScalableBoundedCheck(AigScalableBMCOptions{Frames: frames, SolveAll: true, StoreWitness: true, MaxVariables: 20000})
		if err != nil {
			t.Fatal(err)
		}
		reachable := [2]bool{}
		for sequence := 0; sequence < 1<<frames; sequence++ {
			state := []bool{false}
			for frame := 0; frame < frames; frame++ {
				outputs, nextState, stepErr := graph.Step([]bool{sequence>>frame&1 != 0}, state)
				if stepErr != nil {
					t.Fatal(stepErr)
				}
				for property := range reachable {
					reachable[property] = reachable[property] || outputs[property]
				}
				state = nextState
			}
		}
		for property, want := range reachable {
			got := result.Outputs[property].Status == BmcCounterexample
			if got != want {
				t.Fatalf("tables=%x/%x/%x frames=%d property=%d got=%v want=%v", output0&15, output1&15, next&15, frames, property, got, want)
			}
			if got && result.Outputs[property].Witness == nil {
				t.Fatal("missing witness", property)
			}
		}
	})
}
