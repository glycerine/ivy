package abc

// Go-only semantic, callback, and fuzz tests for src/sat/bmc/bmcBmcG.c.

import (
	"errors"
	"slices"
	"testing"
)

func TestGiaIncrementalBoundedCheckWitnessAndCallbacks(t *testing.T) {
	g := cexToolSequentialGraph(t)
	g.Name = "incremental"
	before := slices.Clone(g.objects)
	type event struct {
		frame, property int
		failed          bool
	}
	var events []event
	result, err := g.IncrementalBoundedCheck(GiaIncrementalBMCOptions{
		Frames: 3, FrameBatch: 2,
		OnFrameDone: func(frame, property int, failed bool) {
			events = append(events, event{frame, property, failed})
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != BmcCounterexample || result.Frame != 1 || result.Property != 0 || result.Counterexample == nil || !slices.Equal(before, g.objects) {
		t.Fatal(result)
	}
	want := []event{{0, 0, false}, {1, 0, true}}
	if !slices.Equal(events, want) {
		t.Fatal(events, want)
	}
	aig, _, _ := g.ToAig()
	model, _ := NewAigerModel(aig, 1, 0, 0, []LatchInit{LatchInitZero})
	if err = model.ValidateCounterexample(result.Counterexample); err != nil {
		t.Fatal(err)
	}
}

func TestGiaIncrementalBoundedCheckSafeSelectionAndErrors(t *testing.T) {
	g := NewGiaManager()
	p, _ := g.AppendCI()
	r, _ := g.AppendCI()
	g.AppendCO(GiaConst0)
	g.AppendCO(r)
	g.AppendCO(p.Not())
	g.SetRegisterCount(1)
	g.SetConstraintCount(1)
	var callbacks int
	result, err := g.IncrementalBoundedCheck(GiaIncrementalBMCOptions{
		Frames: 3, Properties: []int{0}, OnFrameDone: func(_, _ int, failed bool) {
			if failed {
				t.Fatal("safe property callback failed")
			}
			callbacks++
		},
	})
	if err != nil || result.Status != BmcNoCounterexample || result.CheckedFrames != 3 || callbacks != 3 {
		t.Fatal(result, callbacks, err)
	}
	for _, options := range []GiaIncrementalBMCOptions{
		{}, {Frames: 1, FrameBatch: -1}, {Frames: 1, Properties: []int{1}},
	} {
		if _, checkErr := g.IncrementalBoundedCheck(options); checkErr == nil {
			t.Fatal("accepted invalid/limited options", options)
		}
	}
	for _, options := range []GiaIncrementalBMCOptions{{Frames: 1, MaxObjects: 1}, {Frames: 1, MaxMappingEntries: 1}, {Frames: 1, MaxVariables: 1}} {
		limited, checkErr := g.IncrementalBoundedCheck(options)
		if checkErr == nil && limited.Reason != BmcResourceLimit {
			t.Fatal("resource limit", options, limited, checkErr)
		}
	}
	cancelled, err := g.IncrementalBoundedCheck(GiaIncrementalBMCOptions{Frames: 2, SAT: SatLimits{Stop: func() bool { return true }}})
	if err != nil || cancelled.Status != BmcUnknown || cancelled.Reason != BmcCancelled {
		t.Fatal(cancelled, err)
	}
	extended := NewGiaManager()
	p, _ = extended.AppendCI()
	r, _ = extended.AppendCI()
	x, _ := extended.AppendXorReal(p, r)
	extended.AppendCO(x)
	extended.AppendCO(r)
	extended.SetRegisterCount(1)
	if _, err = extended.IncrementalBoundedCheck(GiaIncrementalBMCOptions{Frames: 1}); !errors.Is(err, ErrUnsupported) {
		t.Fatal("extended graph", err)
	}
}

func TestGiaIncrementalBoundedCheckDeepCone(t *testing.T) {
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
	result, err := g.IncrementalBoundedCheck(GiaIncrementalBMCOptions{
		Frames: 1, MaxObjects: 30_000, MaxMappingEntries: 30_000, MaxVariables: 30_000,
	})
	if err != nil || result.Status != BmcNoCounterexample {
		t.Fatal(result, err)
	}
}

func FuzzGiaIncrementalBoundedCheck(f *testing.F) {
	for _, seed := range [][3]uint8{{8, 10, 3}, {6, 9, 2}, {14, 3, 1}, {1, 12, 3}} {
		f.Add(seed[0], seed[1], seed[2])
	}
	f.Fuzz(func(t *testing.T, outputTable, nextTable, framesByte uint8) {
		frames := int(framesByte%3) + 1
		g := cexDepthTruthGraph(t, outputTable, nextTable)
		result, err := g.IncrementalBoundedCheck(GiaIncrementalBMCOptions{Frames: frames})
		if err != nil {
			t.Fatal(err)
		}
		wantFrame := -1
		for candidate := 0; candidate < frames && wantFrame < 0; candidate++ {
			for inputs := 0; inputs < 1<<(candidate+1); inputs++ {
				state, failed := false, false
				for frame := 0; frame <= candidate; frame++ {
					primary := inputs&(1<<frame) != 0
					failed = cexDepthTruth(outputTable, primary, state)
					state = cexDepthTruth(nextTable, primary, state)
				}
				if failed {
					wantFrame = candidate
					break
				}
			}
		}
		if wantFrame < 0 {
			if result.Status != BmcNoCounterexample {
				t.Fatal("false witness", outputTable&15, nextTable&15, result)
			}
		} else if result.Status != BmcCounterexample || result.Frame != wantFrame || result.Counterexample == nil {
			t.Fatal("missed/late witness", outputTable&15, nextTable&15, wantFrame, result)
		}
	})
}
