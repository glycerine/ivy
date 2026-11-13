package int

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/util"
)

func TestCounterexampleFromUnrolling(t *testing.T) {
	g := aig.NewGraph()
	x, lo := g.NewInput(), g.NewInput()
	g.AddOutput(lo)
	g.AddOutput(g.XOR(x, lo))
	aig.Aig_ManSetRegNum(g, 1)
	if got := Inter_ManGetCounterExample(g, 1, 0); got != nil {
		t.Fatal("initial state cannot satisfy property", got)
	}
	cex := Inter_ManGetCounterExample(g, 2, 0)
	if cex == nil || cex.IFrame != 1 || cex.NRegs != 1 || cex.NPis != 1 {
		t.Fatalf("incorrect counterexample metadata: %+v", cex)
	}
	if util.Abc_InfoHasBit(cex.PData, 0) || !util.Abc_InfoHasBit(cex.PData, 1) {
		t.Fatalf("trace must start in zero state and toggle via first input: %+v", cex)
	}
}

// The Go witness builder searches one exact bounded unrolling: an unsafe
// state reached after one transition is absent from a one-frame query and is
// found by the two-frame query. It also records each frame's PIs in Cex
// frame-major order and enforces the optional constraint output in every
// frame. ABC's intCore.c instead delegates this case to scalable BMC, so this
// test characterizes the reconstruction boundary rather than claiming that
// the two search procedures are equivalent.
func TestCounterexampleUnrollingBoundAndConstrainedTrace(t *testing.T) {
	g := aig.NewGraph()
	a, b, state := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(state)       // bad property
	g.AddOutput(g.AND(a, b)) // constraint violation
	g.AddOutput(g.OR(a, b))  // next state
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	if got := Inter_ManGetCounterExample(g, 1, 0); got != nil {
		t.Fatal("the all-zero initial state is safe")
	}
	cex := Inter_ManGetCounterExample(g, 2, 0)
	if cex == nil || cex.IFrame != 1 || cex.NRegs != 1 || cex.NPis != 2 {
		t.Fatalf("incorrect two-frame witness metadata: %+v", cex)
	}
	if util.Abc_InfoHasBit(cex.PData, 0) {
		t.Fatal("initial register state must be zero")
	}
	if !interVerifyCounterExample(g, cex) {
		t.Fatal("extracted trace did not verify")
	}
	invalid := *cex
	invalid.PData = append([]uint32(nil), cex.PData...)
	invalid.PData[0] &^= 1 << 1
	invalid.PData[0] &^= 1 << 2
	if interVerifyCounterExample(g, &invalid) {
		t.Fatal("trace that fails to reach the bad state passed verification")
	}
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		x := util.Abc_InfoHasBit(cex.PData, 1+2*frame)
		y := util.Abc_InfoHasBit(cex.PData, 2+2*frame)
		out, err := g.Evaluate([]bool{x, y, frame != 0})
		if err != nil {
			t.Fatal(err)
		}
		if out[1] {
			t.Fatalf("constraint violated at frame %d", frame)
		}
		if frame == 0 && !(x || y) {
			t.Fatal("frame-zero inputs must reach the bad state")
		}
		if frame == 1 && !out[0] {
			t.Fatal("last frame must violate the property")
		}
	}
}
