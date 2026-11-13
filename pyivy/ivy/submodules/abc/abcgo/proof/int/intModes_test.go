package int

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestInterpolationModesAgainstReachability(t *testing.T) {
	modes := []struct {
		name      string
		configure func(*Inter_ManParams_t)
	}{
		{"combinational", func(p *Inter_ManParams_t) { p.FCheckKstep = 0 }},
		{"backward", func(p *Inter_ManParams_t) { p.FUseBackward = 1 }},
		{"transition-loop", func(p *Inter_ManParams_t) { p.FTransLoop = 1 }},
		{"rewrite", func(p *Inter_ManParams_t) { p.FRewrite = 1 }},
		{"two-step", func(p *Inter_ManParams_t) { p.NFramesK = 2 }},
	}
	for _, mode := range modes {
		t.Run(mode.name, func(t *testing.T) {
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
				mode.configure(&pars)
				var frame int32
				got := Inter_ManPerformInterpolation(g, &pars, &frame)
				want := int32(1)
				if unsafe {
					want = 0
				}
				if got != want {
					t.Fatalf("unsafe=%v got=%d want=%d frame=%d", unsafe, got, want, frame)
				}
			}
		})
	}
}
