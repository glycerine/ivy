package fra

import (
	"math/rand"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
	"github.com/glycerine/abcgo/aig/saig"
)

// Exercise the real SEC latch and signal correspondence paths against an
// independently explored state space. Unknown is permitted, but a proof or
// counterexample must agree with reachability and every witness must replay.
func TestSecCorrespondenceAgainstReachability(t *testing.T) {
	random := rand.New(rand.NewSource(202711))
	decided := 0
	for trial := 0; trial < 48; trial++ {
		g := aig.NewGraph()
		edges := []aig.Edge{g.NewInput(), g.NewInput(), g.NewInput()}
		for i := 0; i < 9; i++ {
			a, b := edges[random.Intn(len(edges))], edges[random.Intn(len(edges))]
			if random.Intn(2) != 0 {
				a = a.Not()
			}
			if random.Intn(2) != 0 {
				b = b.Not()
			}
			edges = append(edges, g.AND(a, b))
		}
		for i := 0; i < 4; i++ {
			e := edges[random.Intn(len(edges))]
			if random.Intn(2) != 0 {
				e = e.Not()
			}
			g.AddOutput(e)
		}
		aig.Aig_ManSetRegNum(g, 2)
		fails := exhaustiveSmallMiter(t, g)
		var pars Fra_Sec_t
		Fra_SecSetDefaultParams(&pars)
		pars.fSilent = 1
		pars.fRetimeFirst, pars.fRetimeRegs, pars.fFraiging = 0, 0, 0
		pars.fInduction = 1
		pars.nFramesMax = 2
		pars.fInterpolation, pars.fReachability, pars.fUsePdr = 0, 0, 0
		status := Fra_FraigSec(g, &pars, nil)
		switch status {
		case 1:
			decided++
			if fails {
				t.Fatalf("trial %d: proved failing miter", trial)
			}
		case 0:
			decided++
			if !fails {
				t.Fatalf("trial %d: reported failure for safe miter", trial)
			}
			if cex := g.SequentialCounterexample(); cex == nil || saig.Saig_ManVerifyCex(g, cex) == 0 {
				t.Fatalf("trial %d: invalid witness %+v", trial, cex)
			}
		case -1:
		default:
			t.Fatalf("trial %d: invalid status %d", trial, status)
		}
	}
	if decided < 32 {
		t.Fatalf("only %d of 48 cases decided", decided)
	}
}
