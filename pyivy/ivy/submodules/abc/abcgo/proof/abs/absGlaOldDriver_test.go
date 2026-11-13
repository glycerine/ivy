package abs

import (
	"testing"

	"github.com/glycerine/abcgo/aig/gia"
)

func TestOldGlaDriverFindsRealCounterexample(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	n := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, n)
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	p.nFramesMax = 1
	if got := Gia_ManPerformGlaOld(g, &p, 0); got != 0 {
		t.Fatalf("SAT result %d", got)
	}
	if g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) == 0 {
		t.Fatal("invalid returned witness")
	}
}
func TestOldGlaDriverTrivialConstants(t *testing.T) {
	for _, sat := range []bool{false, true} {
		g := gia.Gia_ManStart(4)
		driver := int32(0)
		if sat {
			driver = 1
		}
		gia.Gia_ManAppendCo(g, driver)
		var p Abs_Par_t
		Abs_ParSetDefaults(&p)
		got := Gia_ManPerformGlaOld(g, &p, 0)
		if sat && (got != 0 || g.SequentialCounterexample() == nil) || !sat && got != 1 {
			t.Fatalf("sat=%t result=%d", sat, got)
		}
	}
}

func TestOldGlaDriverFindsSequentialCounterexample(t *testing.T) {
	g := gia.Gia_ManStart(8)
	input, flop := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, flop)
	gia.Gia_ManAppendCo(g, input)
	gia.Gia_ManSetRegNum(g, 1)
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	p.nFramesMax = 2
	if got := Gia_ManPerformGlaOld(g, &p, 0); got != 0 {
		t.Fatalf("sequential SAT result %d", got)
	}
	cex := g.SequentialCounterexample()
	if cex == nil || cex.IFrame != 1 || gia.Gia_ManVerifyCex(g, cex, 0) == 0 {
		t.Fatalf("invalid sequential witness %+v", cex)
	}
}

func TestOldGlaDriverRefinesHiddenLogic(t *testing.T) {
	g := gia.Gia_ManStart(20)
	inputs := make([]int32, 6)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(g)
	}
	root := inputs[0]
	for _, input := range inputs[1:] {
		root = gia.Gia_ManAppendAnd(g, root, input)
	}
	gia.Gia_ManAppendCo(g, root)
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	p.nFramesMax = 1
	if got := Gia_ManPerformGlaOld(g, &p, 0); got != 0 {
		t.Fatalf("SAT result %d", got)
	}
	if cex := g.SequentialCounterexample(); cex == nil || gia.Gia_ManVerifyCex(g, cex, 0) == 0 {
		t.Fatal("missing valid witness")
	}
}

func TestOldGlaDriverRefinesSpuriousCounterexample(t *testing.T) {
	g := gia.Gia_ManStart(20)
	inputs := make([]int32, 5)
	for i := range inputs {
		inputs[i] = gia.Gia_ManAppendCi(g)
	}
	x := gia.Gia_ManAppendAnd(g, inputs[0], inputs[1])
	x = gia.Gia_ManAppendAnd(g, x, inputs[2])
	y := gia.Gia_ManAppendAnd(g, inputs[0]^1, inputs[3])
	y = gia.Gia_ManAppendAnd(g, y, inputs[4])
	gia.Gia_ManAppendCo(g, gia.Gia_ManAppendAnd(g, x, y))
	var p Abs_Par_t
	Abs_ParSetDefaults(&p)
	p.nFramesMax = 1
	if got := Gia_ManPerformGlaOld(g, &p, 0); got != -1 {
		t.Fatalf("unexpected result %d", got)
	}
	if g.SequentialCounterexample() != nil {
		t.Fatal("unsatisfiable graph produced counterexample")
	}
}
