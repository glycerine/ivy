package abc

// Stand-alone tests for public BMC parameter records from src/sat/bmc/bmc.h.

import (
	"reflect"
	"testing"
)

func TestBMCHeaderParameterDefaults(t *testing.T) {
	exact := DefaultBMCExactSynthesisParameters()
	if exact.LUTSize != 2 || !exact.Verbose {
		t.Fatal("exact synthesis defaults", exact)
	}
	exact.LUTSize = 0
	if fresh := DefaultBMCExactSynthesisParameters(); fresh.LUTSize != 2 {
		t.Fatal("exact synthesis defaults alias state", fresh)
	}

	scalable := DefaultAigScalableBMCParameters()
	if scalable.LearnedStart != 10000 || scalable.LearnedDelta != 2000 || scalable.LearnedPercent != 80 || scalable.ExploredFrame != -1 {
		t.Fatal("scalable defaults", scalable)
	}
	scalable.LearnedStart = 0
	if fresh := DefaultAigScalableBMCParameters(); fresh.LearnedStart != 10000 {
		t.Fatal("scalable defaults alias state", fresh)
	}

	ordinary := DefaultGiaAndBMCParameters()
	if ordinary.AdditionalFrames != 50 || ordinary.LUTSize != 6 || ordinary.Workers != 1 {
		t.Fatal("AND BMC defaults", ordinary)
	}
	super := DefaultGiaSuperBMCParameters()
	if super.AdditionalFrames != 1 || super.LUTSize != 0 || super.Workers != 1 {
		t.Fatal("super BMC defaults", super)
	}
	if reflect.DeepEqual(ordinary, super) {
		t.Fatal("distinct command defaults collapsed")
	}

	if core := DefaultGiaBoundedCoreParameters(); core.Frame != 10 || core.Output != 0 || core.Timeout != 0 || core.PivotFile != "" || core.ProofFile != "" || core.Verbose {
		t.Fatal("bounded-core defaults", core)
	}
	if multi := DefaultMultiProveParameters(); multi.GlobalTimeout != 30 || multi.LocalTimeout != 2 || multi.TimeoutIncrease != 100 || multi.GapTimeout != 0 || multi.PerOutputTimeout != 0 {
		t.Fatal("multi-prove defaults", multi)
	}
	if fault := DefaultGiaFaultParameters(); !reflect.DeepEqual(fault, GiaFaultParameters{}) {
		t.Fatal("fault defaults", fault)
	}
}

func TestBMCHeaderParameterCallbacksAreConcrete(t *testing.T) {
	failures, stops, frames := 0, 0, 0
	scalable := DefaultAigScalableBMCParameters()
	scalable.OnFailure = func(property int, witness *Counterexample) bool {
		failures += property + 1
		return witness == nil
	}
	scalable.Stop = func(runID int) bool {
		stops += runID
		return true
	}
	if !scalable.OnFailure(2, nil) || !scalable.Stop(7) || failures != 3 || stops != 7 {
		t.Fatal(failures, stops)
	}
	and := DefaultGiaAndBMCParameters()
	and.OnFrameDone = func(frame, property int, status BmcStatus) {
		frames += frame + property + int(status)
	}
	and.OnFrameDone(2, 3, BmcCounterexample)
	if frames != 7 {
		t.Fatal(frames)
	}
}
