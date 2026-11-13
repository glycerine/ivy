package abs

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestGlaDriverFindsAndReplaysCounterexample(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a := gia.Gia_ManAppendCi(g)
	b := gia.Gia_ManAppendCi(g)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	pars.nFramesMax = 2
	pars.nRatioMin = 0
	pars.nRatioMin2 = 0
	pars.nFramesNoChangeLim = 100
	got := Gia_ManPerformGla(g, pars)
	if got != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
		t.Fatalf("result %d invalid witness", got)
	}
}
func TestGlaDriverHandlesDirectInputsAndConstants(t *testing.T) {
	for _, kind := range []int{0, 1, 2, 3} {
		g := gia.Gia_ManStart(5)
		x := gia.Gia_ManAppendCi(g)
		lit := int32(kind & 1)
		if kind >= 2 {
			lit = x ^ int32(kind&1)
		}
		gia.Gia_ManAppendCo(g, lit)
		pars := &Abs_Par_t{}
		Abs_ParSetDefaults(pars)
		got := Gia_ManPerformGla(g, pars)
		if kind == 0 {
			if got != 1 {
				t.Fatal("constant false")
			}
		} else if got != 0 || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
			t.Fatalf("kind %d invalid result %d", kind, got)
		}
	}
}

func TestGlaDriverRefinesSpuriousCounterexample(t *testing.T) {
	g := gia.Gia_ManStart(12)
	x := gia.Gia_ManAppendCi(g)
	y := gia.Gia_ManAppendCi(g)
	a := gia.Gia_ManAppendAnd(g, x, y)
	b := gia.Gia_ManAppendAnd(g, x^1, y)
	gate := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, gate)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	pars.fUseSimple = 1
	pars.nFramesMax = 1
	pars.nRatioMin = 0
	pars.nRatioMin2 = 0
	pars.nFramesNoChangeLim = 100
	if result := Gia_ManPerformGla(g, pars); result != -1 || g.SequentialCounterexample() != nil || pars.iFrameProved != 0 {
		t.Fatalf("result %d, frame %d, witness %v", result, pars.iFrameProved, g.SequentialCounterexample())
	}
}

func TestGlaDriverSequentialCounterexample(t *testing.T) {
	g := gia.Gia_ManStart(24)
	x := gia.Gia_ManAppendCi(g)
	q1 := gia.Gia_ManAppendCi(g)
	q2 := gia.Gia_ManAppendCi(g)
	out := gia.Gia_ManAppendAnd(g, x, q2)
	gia.Gia_ManAppendCo(g, out)
	gia.Gia_ManAppendCo(g, x)
	gia.Gia_ManAppendCo(g, q1)
	gia.Gia_ManSetRegNum(g, 2)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	pars.nFramesMax = 4
	pars.nRatioMin = 0
	pars.nRatioMin2 = 0
	pars.nFramesNoChangeLim = 100
	if result := Gia_ManPerformGla(g, pars); result != 0 || g.SequentialCounterexample() == nil || g.SequentialCounterexample().IFrame != 2 || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
		t.Fatalf("result %d, witness %+v", result, g.SequentialCounterexample())
	}
}
