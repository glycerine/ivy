package int

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestInterpolationDefaultsAndInitialFailure(t *testing.T) {
	var pars Inter_ManParams_t
	Inter_ManSetDefaultParams(&pars)
	if pars.NFramesK != 1 || pars.FCheckKstep != 1 || pars.IFrameMax != -1 {
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
	pars.NFramesMax = 1
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
		pars.NFramesMax = 5
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

// Exhaustively compare all one-register transition functions with explicit
// state exploration. The property is state AND input, so an error may need
// multiple frames to become reachable.
func TestInterpolationAgainstExplicitReachability(t *testing.T) {
	for table := 0; table < 16; table++ {
		g := aig.NewGraph()
		x, state := g.NewInput(), g.NewInput()
		g.AddOutput(g.AND(state, x))
		next := g.Const0()
		for mask := 0; mask < 4; mask++ {
			if table&(1<<mask) == 0 {
				continue
			}
			a, b := x, state
			if mask&1 == 0 {
				a = a.Not()
			}
			if mask&2 == 0 {
				b = b.Not()
			}
			next = g.OR(next, g.AND(a, b))
		}
		g.AddOutput(next)
		aig.Aig_ManSetRegNum(g, 1)
		seen := [2]bool{true, false}
		unsafe := false
		for changed := true; changed; {
			changed = false
			for s := 0; s < 2; s++ {
				if !seen[s] {
					continue
				}
				for input := 0; input < 2; input++ {
					if s == 1 && input == 1 {
						unsafe = true
					}
					n := (table >> uint(input+2*s)) & 1
					if !seen[n] {
						seen[n] = true
						changed = true
					}
				}
			}
		}
		var pars Inter_ManParams_t
		Inter_ManSetDefaultParams(&pars)
		pars.NFramesMax = 5
		var frame int32
		got := Inter_ManPerformInterpolation(g, &pars, &frame)
		want := int32(1)
		if unsafe {
			want = 0
		}
		if got != want {
			t.Fatalf("table=%04b: got %d want %d frame %d", table, got, want, frame)
		}
	}
}
