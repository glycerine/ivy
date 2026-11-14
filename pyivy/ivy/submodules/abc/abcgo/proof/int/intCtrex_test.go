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
	cex := Inter_ManGetCounterExample(g, 0, 0)
	if cex == nil || cex.IFrame != 1 || cex.NRegs != 1 || cex.NPis != 1 {
		t.Fatalf("incorrect counterexample metadata: %+v", cex)
	}
	if util.Abc_InfoHasBit(cex.PData, 0) || !util.Abc_InfoHasBit(cex.PData, 1) {
		t.Fatalf("trace must start in zero state and toggle via first input: %+v", cex)
	}
	if got := Inter_ManGetCounterExample(g, 1, 0); got == nil || got.IFrame != 1 {
		t.Fatalf("start frame 1 should return the frame-1 witness: %+v", got)
	}
	if got := Inter_ManGetCounterExample(g, 2, 0); got != nil {
		t.Fatalf("complete one-register search must stop before frame 2: %+v", got)
	}
}

func TestScalableBmcWitnessStartsAtRequestedFrameAndMapsPis(t *testing.T) {
	g := aig.NewGraph()
	a, b, state := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(state)
	g.AddOutput(g.OR(a, b))
	aig.Aig_ManSetRegNum(g, 1)
	cex := Inter_ManGetCounterExample(g, 1, 0)
	if cex == nil || cex.IFrame != 1 || cex.IPo != 0 || cex.NRegs != 1 || cex.NPis != 2 {
		t.Fatalf("unexpected scalable-BMC witness: %+v", cex)
	}
	if util.Abc_InfoHasBit(cex.PData, 0) {
		t.Fatal("initial register bit must be zero")
	}
	x0 := util.Abc_InfoHasBit(cex.PData, 1)
	y0 := util.Abc_InfoHasBit(cex.PData, 2)
	if !x0 && !y0 {
		t.Fatal("frame-zero inputs do not reach the frame-one bad state")
	}
	for frame := int32(0); frame <= cex.IFrame; frame++ {
		x := util.Abc_InfoHasBit(cex.PData, 1+2*frame)
		y := util.Abc_InfoHasBit(cex.PData, 2+2*frame)
		badState := frame != 0
		out, err := g.Evaluate([]bool{x, y, badState})
		if err != nil {
			t.Fatal(err)
		}
		if frame == cex.IFrame && !out[0] {
			t.Fatalf("property is false at returned frame %d", frame)
		}
	}
}

func TestScalableBmcSourceContractTreatsConstraintOutputAsTarget(t *testing.T) {
	g := aig.NewGraph()
	x, state := g.NewInput(), g.NewInput()
	g.AddOutput(state) // ordinary property is false in the initial state
	g.AddOutput(x)     // trailing constraint output is independently targeted by C BMC
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	aig.Aig_ManSetConstrNum(g, 1)
	cex := Inter_ManGetCounterExample(g, 0, 0)
	if cex == nil || cex.IFrame != 0 || cex.IPo != 1 {
		t.Fatalf("C scalable-BMC output-target behavior changed: %+v", cex)
	}
	if !util.Abc_InfoHasBit(cex.PData, 1) {
		t.Fatal("constraint-output witness is missing its frame-zero PI")
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
	if got := Inter_ManGetCounterExampleConstrained(g, 1, 0); got != nil {
		t.Fatal("the all-zero initial state is safe")
	}
	cex := Inter_ManGetCounterExampleConstrained(g, 2, 0)
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
