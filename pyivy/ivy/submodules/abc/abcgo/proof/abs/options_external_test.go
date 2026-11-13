package abs_test

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
	"github.com/glycerine/abcgo/proof/abs"
)

func TestOptionsDefaultsAndRealOldGLARun(t *testing.T) {
	p := abs.NewParameters()
	o := p.Options()
	if o != abs.DefaultOptions() || o.FramesPast != 4 || o.LearnedMax != 1000 || !o.PropagateFanout {
		t.Fatalf("unexpected defaults: %+v", o)
	}
	if got := p.Progress(); got.Frame != -1 || got.FrameProved != -1 {
		t.Fatalf("initial progress = %+v", got)
	}
	// Exercise a non-default configuration through the actual public driver.
	o.FramesMax = 2
	o.AddLayer = true
	o.ConflictLimit = 10000
	p.SetOptions(o)
	if got := p.Options(); got != o {
		t.Fatalf("options round trip: %+v", got)
	}
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	state := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, state)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManSetRegNum(g, 1)
	if got := abs.Gia_ManPerformGlaOld(g, p, 0); got != 0 {
		t.Fatalf("expected sequential counterexample, got %d", got)
	}
	witness := g.SequentialCounterexample()
	if witness == nil || witness.IFrame != 1 || gia.Gia_ManVerifyCex(g, witness, 0) != 1 {
		t.Fatalf("invalid witness: %+v", witness)
	}
	// Applying options does not erase output fields updated by the driver.
	progress := p.Progress()
	p.SetOptions(p.Options())
	if p.Progress() != progress {
		t.Fatal("setting options erased progress")
	}
}

func TestLegacyOptionsAndProgress(t *testing.T) {
	p := abs.NewLegacyParameters()
	o := p.Options()
	if o != abs.DefaultLegacyOptions() || o.FramesMax != 10 || o.ConflictLimit != 10000 || !o.Dynamic {
		t.Fatalf("unexpected legacy defaults: %+v", o)
	}
	if got := p.Progress(); got.Status != -1 || got.FramesDone != -1 {
		t.Fatalf("initial progress: %+v", got)
	}
	o.Algorithm, o.FramesMax, o.BMCFrames = 2, 4, 7
	o.Dynamic, o.Constraints, o.UseBDDs, o.Verbose = false, true, true, true
	o.TimeoutSeconds, o.VTATimeoutSeconds, o.BobParameter = 9, 11, 13
	p.SetOptions(o)
	if got := p.Options(); got != o {
		t.Fatalf("legacy options round trip: %+v", got)
	}
	if got := p.Progress(); got.Status != -1 || got.FramesDone != -1 {
		t.Fatalf("options erased progress: %+v", got)
	}
}

func TestAllOptionsRoundTrip(t *testing.T) {
	o := abs.Options{
		FramesMax: 1, FramesStart: 2, FramesPast: 3, ConflictLimit: 4,
		LearnedMax: 5, LearnedStart: 6, LearnedDelta: 7, LearnedPercent: 8,
		TimeoutSeconds: 9, RatioMin: 10, RatioMinRefine: 11, RatioMax: 12,
		UseTerminalVariables: true, UseRollback: true, PropagateFanout: true,
		AddLayer: true, NewRefinement: true, SkipIntermediate: true,
		SimpleCNF: true, SkipHash: true, UseFullProof: true,
		DumpVariableAbstraction: true, DumpMappedAbstraction: true,
		CallProver: true, SimplifyBeforeProver: true,
		VariableAbstractionFile: "example.aig", Verbose: true, VeryVerbose: true,
		FramesNoChangeLimit: 13,
	}
	p := abs.NewParameters()
	p.SetOptions(o)
	if got := p.Options(); got != o {
		t.Fatalf("all options round trip: %+v", got)
	}
	p.SetOptions(abs.Options{})
	if got := p.Options(); got != (abs.Options{}) {
		t.Fatalf("clearing options: %+v", got)
	}
}
