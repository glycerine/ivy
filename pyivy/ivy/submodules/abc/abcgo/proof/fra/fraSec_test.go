package fra

import (
	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
	"testing"
)

func TestSecDefaultsAndTrivialMiter(t *testing.T) {
	var opts Fra_Sec_t
	Fra_SecSetDefaultParams(&opts)
	if opts.nFramesMax != 4 || opts.nBTLimit != 1000 || opts.fUsePdr != 1 || opts.fInterpolation != 1 {
		t.Fatalf("defaults %+v", opts)
	}
	for _, fails := range []bool{false, true} {
		g := aig.NewGraph()
		g.NewInput()
		g.NewInput()
		g.AddOutput(g.Const0())
		if fails {
			g.AddOutput(g.Const1())
		} else {
			g.AddOutput(g.Const0())
		}
		g.AddOutput(g.Const0())
		aig.Aig_ManSetRegNum(g, 1)
		opts.fSilent = 1
		got := Fra_FraigSecVerified(g, &opts, nil)
		want := int32(1)
		if fails {
			want = 0
		}
		if got != want {
			t.Fatalf("fails=%v got=%d want=%d", fails, got, want)
		}
		if fails {
			cex := g.SequentialCounterexample()
			if cex == nil || cex.IPo != 1 || saig.Saig_ManVerifyCexNoClear(g, cex) == 0 {
				t.Fatalf("invalid witness %+v", cex)
			}
		}
	}
}

func TestSecSequentialMiterSafeAndFailing(t *testing.T) {
	for _, tc := range []struct {
		name string
		fail bool
		want int32
	}{{"safe", false, 1}, {"failing", true, 0}} {
		t.Run(tc.name, func(t *testing.T) {
			g := aig.NewGraph()
			input, a, b := g.NewInput(), g.NewInput(), g.NewInput()
			if tc.fail {
				g.AddOutput(g.AND(input, a))
			} else {
				g.AddOutput(g.XOR(a, b))
			}
			g.AddOutput(input)
			g.AddOutput(input)
			aig.Aig_ManSetRegNum(g, 2)
			var pars Fra_Sec_t
			Fra_SecSetDefaultParams(&pars)
			pars.fSilent = 1
			pars.fRetimeFirst = 0
			pars.fRetimeRegs = 0
			if got := Fra_FraigSec(g, &pars, nil); got != tc.want {
				t.Fatalf("status=%d want=%d", got, tc.want)
			}
			if tc.fail {
				cex := g.SequentialCounterexample()
				if cex == nil || saig.Saig_ManVerifyCex(g, cex) == 0 {
					t.Fatalf("invalid witness: %+v", cex)
				}
			}
		})
	}
}

func TestSecSequentialMiterWithDefaultRetiming(t *testing.T) {
	g := aig.NewGraph()
	input, a := g.NewInput(), g.NewInput()
	g.NewInput()
	g.AddOutput(g.AND(input, a))
	g.AddOutput(input)
	g.AddOutput(input)
	aig.Aig_ManSetRegNum(g, 2)
	var pars Fra_Sec_t
	Fra_SecSetDefaultParams(&pars)
	pars.fSilent = 1
	if got := Fra_FraigSec(g, &pars, nil); got != 0 {
		t.Fatalf("status=%d want failure", got)
	}
	if cex := g.SequentialCounterexample(); cex == nil || saig.Saig_ManVerifyCex(g, cex) == 0 {
		t.Fatalf("invalid witness: %+v", cex)
	}
}

func TestSecSafeMiterWithDefaultRetiming(t *testing.T) {
	g := aig.NewGraph()
	input, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.XOR(a, b))
	g.AddOutput(input)
	g.AddOutput(input)
	aig.Aig_ManSetRegNum(g, 2)
	var pars Fra_Sec_t
	Fra_SecSetDefaultParams(&pars)
	pars.fSilent = 1
	if got := Fra_FraigSec(g, &pars, nil); got != 1 {
		t.Fatalf("status=%d want proof", got)
	}
}

func TestSecGlobalConflictLimitReturnsUnknown(t *testing.T) {
	g := aig.NewGraph()
	input, a, b := g.NewInput(), g.NewInput(), g.NewInput()
	g.AddOutput(g.AND(a, b))
	g.AddOutput(input)
	g.AddOutput(input.Not())
	aig.Aig_ManSetRegNum(g, 2)
	var pars Fra_Sec_t
	Fra_SecSetDefaultParams(&pars)
	pars.fSilent = 1
	pars.fRetimeFirst = 0
	pars.fRetimeRegs = 0
	pars.fFraiging = 0
	pars.fInterpolation = 0
	pars.fReachability = 0
	pars.fUsePdr = 0
	pars.nBTLimitGlobal = -1
	var reduced *aig.Graph
	if got := Fra_FraigSec(g, &pars, &reduced); got != -1 {
		t.Fatalf("status=%d, want unknown", got)
	}
	if reduced == nil {
		t.Fatal("unknown result did not retain the reduced miter")
	}
	if g.SequentialCounterexample() != nil {
		t.Fatal("unknown result claimed a counterexample")
	}
}
