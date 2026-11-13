package int

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestInterpolationDefaultsAndInitialFailure(t *testing.T) {
	var pars Inter_ManParams_t
	Inter_ManSetDefaultParams(&pars)
	if pars.nFramesK != 1 || pars.fCheckKstep != 1 || pars.iFrameMax != -1 {
		t.Fatalf("unexpected defaults: %+v", pars)
	}
	g := aig.NewGraph()
	x := g.NewInput()
	g.NewInput()
	g.AddOutput(x)
	g.AddOutput(g.Const0())
	aig.Aig_ManSetRegNum(g, 1)
	frame := int32(42)
	if got := Inter_ManPerformInterpolation(g, &pars, &frame); got != 0 || frame != -1 || g.SequentialCounterexample() == nil {
		t.Fatalf("initial failure: result %d frame %d witness %v", got, frame, g.SequentialCounterexample())
	}
}

func TestInterpolationFrameLimit(t *testing.T) {
	var pars Inter_ManParams_t
	Inter_ManSetDefaultParams(&pars)
	pars.nFramesMax = 1
	g := aig.NewGraph()
	g.NewInput()
	state := g.NewInput()
	g.AddOutput(state)
	g.AddOutput(state)
	aig.Aig_ManSetRegNum(g, 1)
	var frame int32
	if got := Inter_ManPerformInterpolation(g, &pars, &frame); got != -1 || frame != -1 {
		t.Fatalf("got %d, frame %d", got, frame)
	}
}

func TestInterpolationEndToEnd(t *testing.T) {
	for _, unsafe := range []bool{false, true} {
		g := aig.NewGraph()
		x, state := g.NewInput(), g.NewInput()
		g.AddOutput(state)
		if unsafe {
			g.AddOutput(x)
		} else {
			g.AddOutput(g.Const0())
		}
		aig.Aig_ManSetRegNum(g, 1)
		var pars Inter_ManParams_t
		Inter_ManSetDefaultParams(&pars)
		pars.nFramesMax = 5
		var frame int32
		got := Inter_ManPerformInterpolation(g, &pars, &frame)
		want := int32(1)
		if unsafe {
			want = 0
		}
		if got != want {
			t.Fatalf("unsafe=%v: got %d want %d frame %d", unsafe, got, want, frame)
		}
		if unsafe && (frame != 1 || g.SequentialCounterexample() == nil) {
			t.Fatalf("missing frame-one witness: frame %d", frame)
		}
	}
}
