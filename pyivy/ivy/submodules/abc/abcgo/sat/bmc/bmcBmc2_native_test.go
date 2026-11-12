package bmc

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestSaigBmcPerformUsesNativeChecker(t *testing.T) {
	g := aig.NewGraph()
	_, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	failedFrame := int32(-1)
	if status := Saig_BmcPerform(g, 0, 2, 100, 0, 0, 0, 0, 0, &failedFrame, 0, 0); status != -1 {
		t.Fatalf("status = %d, want legacy bounded-safe (-1)", status)
	}
	if failedFrame != 1 {
		t.Fatalf("failed-frame progress = %d, want last checked frame 1", failedFrame)
	}

	unsafe := aig.NewGraph()
	pi, state := unsafe.NewInput(), unsafe.NewInput()
	unsafe.AddOutput(state)
	unsafe.AddOutput(pi)
	aig.Aig_ManSetRegNum(unsafe, 1)
	failedFrame = -1
	if status := Saig_BmcPerform(unsafe, 0, 2, 100, 0, 0, 0, 0, 0, &failedFrame, 0, 0); status != 0 {
		t.Fatalf("counterexample status = %d, want 0", status)
	}
	if failedFrame != 0 {
		t.Fatalf("reported counterexample frame = %d, want 0", failedFrame)
	}
}

func TestSaigManBmcSimplePreservesLegacyStatus(t *testing.T) {
	g := aig.NewGraph()
	pi, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	frame := int32(-1)
	if status := Saig_ManBmcSimple(g, 2, 0, 0, 0, 0, &frame, 0, 0); status != 0 {
		t.Fatalf("status = %d, want counterexample (0)", status)
	}
	if frame != 1 {
		t.Fatalf("counterexample frame = %d, want 1", frame)
	}

	safe := aig.NewGraph()
	_, state = safe.NewInput(), safe.NewInput()
	safe.AddOutput(state)
	safe.AddOutput(state)
	aig.Aig_ManSetRegNum(safe, 1)
	if status := Saig_ManBmcSimple(safe, 2, 0, 0, 0, 0, nil, 0, 0); status != -1 {
		t.Fatalf("safe status = %d, want legacy bounded-safe (-1)", status)
	}
}

func TestAbsManTernarySimulate(t *testing.T) {
	g := aig.NewGraph()
	_, state := g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	frames := Abs_ManTernarySimulate(g, 4, 0)
	if got := vec.Vec_PtrSize(frames); got != 4 {
		t.Fatalf("proved frames = %d, want 4", got)
	}
	var object aig.Edge = state
	if got := Abs_ManSimInfoGet(frames, &object, 0); got != ABS_ZER {
		t.Fatalf("frame 0 register output = %d, want zero", got)
	}

	unknown := aig.NewGraph()
	pi, ro := unknown.NewInput(), unknown.NewInput()
	unknown.AddOutput(pi)
	unknown.AddOutput(ro)
	aig.Aig_ManSetRegNum(unknown, 1)
	frames = Abs_ManTernarySimulate(unknown, 4, 0)
	if got := vec.Vec_PtrSize(frames); got != 1 {
		t.Fatalf("unknown property frames = %d, want early stop at 1", got)
	}
}

func TestAbsManSimInfoPackedAccess(t *testing.T) {
	frames := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(frames, make([]uint32, 1))
	var object aig.Edge = aig.Edge(2 << 1)
	Abs_ManSimInfoSet(frames, &object, 0, ABS_UND)
	if got := Abs_ManSimInfoGet(frames, &object, 0); got != ABS_UND {
		t.Fatalf("packed value = %d, want unknown", got)
	}
}
