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
	pars.NFramesMax = 2
	pars.NRatioMin = 0
	pars.NRatioMin2 = 0
	pars.NFramesNoChangeLim = 100
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
	pars.FUseSimple = 1
	pars.NFramesMax = 1
	pars.NRatioMin = 0
	pars.NRatioMin2 = 0
	pars.NFramesNoChangeLim = 100
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
	pars.NFramesMax = 4
	pars.NRatioMin = 0
	pars.NRatioMin2 = 0
	pars.NFramesNoChangeLim = 100
	if result := Gia_ManPerformGla(g, pars); result != 0 || g.SequentialCounterexample() == nil || g.SequentialCounterexample().IFrame != 2 || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
		t.Fatalf("result %d, witness %+v", result, g.SequentialCounterexample())
	}
}

func TestGlaDriverConstantTrueAtLaterFrameSourceAssertion(t *testing.T) {
	g := gia.Gia_ManStart(6)
	gia.Gia_ManAppendCi(g) // unused PI, matching the source fixture
	q := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, q)
	gia.Gia_ManAppendCo(g, 1)
	gia.Gia_ManSetRegNum(g, 1)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	pars.NFramesMax = 3
	pars.NRatioMin = 0
	pars.NRatioMin2 = 0
	pars.NFramesNoChangeLim = 100
	defer func() {
		if got := recover(); got != "abs: GLA target literal must exceed one" {
			t.Fatalf("panic=%v, want source literal assertion", got)
		}
	}()
	Gia_ManPerformGla(g, pars)
}

func TestGlaDriverLayerModeTerminatesWithRealWitness(t *testing.T) {
	g := gia.Gia_ManStart(8)
	x := gia.Gia_ManAppendCi(g)
	y := gia.Gia_ManAppendCi(g)
	out := gia.Gia_ManAppendAnd(g, x, y)
	gia.Gia_ManAppendCo(g, out)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	pars.FAddLayer = 1
	pars.NFramesMax = 1
	pars.NRatioMin = 0
	pars.NRatioMin2 = 0
	if result := Gia_ManPerformGla(g, pars); result != 0 || g.SequentialCounterexample() == nil || gia.Gia_ManVerifyCex(g, g.SequentialCounterexample(), 0) != 1 {
		t.Fatalf("result %d witness %+v", result, g.SequentialCounterexample())
	}
}

func TestGlaDriverTimesOutOnConstantFalseFrames(t *testing.T) {
	g := gia.Gia_ManStart(6)
	q := gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, q)
	gia.Gia_ManAppendCo(g, 0)
	gia.Gia_ManSetRegNum(g, 1)
	pars := &Abs_Par_t{}
	Abs_ParSetDefaults(pars)
	pars.NFramesMax = 0
	pars.NTimeOut = 1
	pars.NRatioMin = 0
	pars.NRatioMin2 = 0
	pars.NFramesNoChangeLim = 100
	if result := Gia_ManPerformGla(g, pars); result != -1 || g.SequentialCounterexample() != nil {
		t.Fatalf("result %d witness %+v", result, g.SequentialCounterexample())
	}
}

func TestGlaFrameLimitWithTimeoutAndConflictLimit(t *testing.T) {
	// The first bad state is frame 2. C tests the frame bound before
	// entering a frame; the enabled timeout and conflict limit must not
	// turn a bounded unknown into a witness or a proof.
	for _, frames := range []int32{2, 3} {
		g := gia.Gia_ManStart(8)
		x := gia.Gia_ManAppendCi(g)
		q1, q2 := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, q2)
		gia.Gia_ManAppendCo(g, x)
		gia.Gia_ManAppendCo(g, q1)
		gia.Gia_ManSetRegNum(g, 2)
		pars := NewParameters()
		o := pars.Options()
		o.FramesMax, o.TimeoutSeconds, o.ConflictLimit = frames, 30, 1000
		o.RatioMin, o.RatioMinRefine, o.FramesNoChangeLimit = 0, 0, 100
		pars.SetOptions(o)
		got := Gia_ManPerformGla(g, pars)
		cex := g.SequentialCounterexample()
		if frames == 2 {
			if got != -1 || cex != nil {
				t.Fatalf("frames=%d result=%d witness=%+v", frames, got, cex)
			}
		} else if got != 0 || cex == nil || cex.IFrame != 2 || gia.Gia_ManVerifyCex(g, cex, 0) != 1 {
			t.Fatalf("frames=%d result=%d witness=%+v", frames, got, cex)
		}
	}
}

// ABC skips checking frames up to and including nFramesStart or iFrameProved
// while still unrolling their transitions (absGla.c, frame loop).
func TestGlaSkippedFramesStillUnrollState(t *testing.T) {
	for _, mode := range []string{"start", "proved"} {
		g := gia.Gia_ManStart(8)
		x := gia.Gia_ManAppendCi(g)
		q1, q2 := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
		gia.Gia_ManAppendCo(g, q2)
		gia.Gia_ManAppendCo(g, x)
		gia.Gia_ManAppendCo(g, q1)
		gia.Gia_ManSetRegNum(g, 2)
		p := NewParameters()
		p.NFramesMax, p.NFramesNoChangeLim = 4, 100
		p.NRatioMin, p.NRatioMin2 = 0, 0
		if mode == "start" {
			p.NFramesStart = 2
		} else {
			p.FUseSkip, p.iFrameProved = 1, 2
		}
		got := Gia_ManPerformGla(g, p)
		cex := g.SequentialCounterexample()
		if got != 0 || cex == nil || cex.IFrame != 3 || gia.Gia_ManVerifyCex(g, cex, 0) != 1 {
			t.Fatalf("mode=%s result=%d witness=%+v", mode, got, cex)
		}
	}
}
