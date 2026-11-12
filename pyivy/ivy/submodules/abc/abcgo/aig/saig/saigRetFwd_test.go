package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

func TestSaigMarkAutonomousDoesNotEnterClosedLatchCycle(t *testing.T) {
	g := aig.NewGraph()
	pi, latch := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	g.AddOutput(latch) // latch feeds its own next state
	aig.Aig_ManSetRegNum(g, 1)

	Saig_ManMarkAutonomous(g)
	if !aig.Aig_ObjIsTravIdCurrent(g, pi) {
		t.Fatal("primary input should be marked autonomous")
	}
	if aig.Aig_ObjIsTravIdCurrent(g, latch) {
		t.Fatal("self-contained latch cycle should remain unreachable")
	}
}

func TestSaigRetimeForwardOneKeepsUnreachableLatch(t *testing.T) {
	g := aig.NewGraph()
	pi, latch := g.NewInput(), g.NewInput()
	g.AddOutput(pi)
	g.AddOutput(latch)
	aig.Aig_ManSetRegNum(g, 1)

	fixed, moved := int32(-1), int32(-1)
	got := Saig_ManRetimeForwardOne(g, &fixed, &moved)
	if got == nil {
		t.Fatal("forward retiming returned nil")
	}
	if fixed != 1 || moved != 0 {
		t.Fatalf("fixed=%d moved=%d, want one fixed latch and no moved gates", fixed, moved)
	}
	if aig.Aig_ManRegNum(got) != 1 || len(got.Outputs()) != 2 {
		t.Fatalf("retimed graph has %d regs and %d outputs, want 1 and 2", aig.Aig_ManRegNum(got), len(got.Outputs()))
	}
}
