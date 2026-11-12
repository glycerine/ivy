package saig

import (
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/misc/vec"
)

func TestRetimeSATInitialStateAndUnsatCore(t *testing.T) {
	g := aig.NewGraph()
	x := g.NewInput()
	g.AddOutput(x.Not())
	initial := Saig_ManRetimeInitState(g)
	if initial == nil || vec.Vec_IntSize(initial) != 1 || vec.Vec_IntEntry(initial, 0) != 1 {
		t.Fatalf("initial assignment = %v", initial)
	}
	h := aig.NewGraph()
	y := h.NewInput()
	h.AddOutput(y)
	h.AddOutput(y.Not())
	if output := Saig_ManRetimeUnsatCore(h, 0); output < 0 || output > 1 {
		t.Fatalf("unexpected unsat-core output index %d", output)
	}
}

func TestRetimeExcludeHideExposeBadRegisters(t *testing.T) {
	g := aig.NewGraph()
	pi, lo0, lo1 := g.NewInput(), g.NewInput(), g.NewInput()
	driver := g.AND(pi, lo0)
	g.AddOutput(pi)
	g.AddOutput(driver)
	g.AddOutput(driver.Not())
	aig.Aig_ManSetRegNum(g, 2)
	bad := Saig_ManGetRegistersToExclude(g)
	if got := vec.Vec_PtrSize(bad); got != 2 {
		t.Fatalf("bad registers = %d, want 2", got)
	}
	if got := Saig_ManHideBadRegs(g, bad); got != 2 || Saig_ManRegNum(g) != 0 || Saig_ManPiNum(g) != 3 || Saig_ManPoNum(g) != 3 {
		t.Fatalf("after hiding: moved=%d regs=%d PIs=%d POs=%d", got, Saig_ManRegNum(g), Saig_ManPiNum(g), Saig_ManPoNum(g))
	}
	Saig_ManExposeBadRegs(g, 2)
	if Saig_ManRegNum(g) != 2 || Saig_ManPiNum(g) != 1 || Saig_ManPoNum(g) != 1 {
		t.Fatalf("after exposing: regs=%d PIs=%d POs=%d", Saig_ManRegNum(g), Saig_ManPiNum(g), Saig_ManPoNum(g))
	}
	_ = lo1
}

func TestRetimeBackwardDuplicationPreservesOutputs(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	next := g.AND(pi, lo)
	g.AddOutput(pi)
	g.AddOutput(next)
	aig.Aig_ManSetRegNum(g, 1)
	cut := vec.Vec_PtrAlloc(1)
	vec.Vec_PtrPush(cut, aig.Aig_Regular(next))
	init := vec.Vec_IntAlloc(1)
	vec.Vec_IntPush(init, 0)
	got := Saig_ManRetimeDupBackward(g, cut, init)
	if got == nil || aig.Aig_ManRegNum(got) != 1 || Saig_ManPiNum(got) != 1 || Saig_ManPoNum(got) != 1 {
		t.Fatalf("bad backward result: graph=%v", got)
	}
	if err := got.Check(); err != nil { t.Fatalf("invalid retimed graph: %v", err) }
}
