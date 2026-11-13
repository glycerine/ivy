package ssw

import (
	"github.com/glycerine/abcgo/aig/aig"
	"testing"
)

func TestSignalCorrespondenceRunsInitializedBmc(t *testing.T) {
	g := aig.NewGraph()
	pi, lo := g.NewInput(), g.NewInput()
	out := g.AND(pi, lo)
	g.AddOutput(out)
	g.AddOutput(pi)
	aig.Aig_ManSetRegNum(g, 1)
	var pars Ssw_Pars_t
	Ssw_ManSetDefaultParams(&pars)
	pars.nStepsMax = 0
	reduced := Ssw_SignalCorrespondence(g, &pars)
	if reduced == nil || aig.Aig_ManRegNum(reduced) != 1 || pars.nIters != 0 {
		t.Fatalf("BMC correspondence failed: result=%v iterations=%d", reduced, pars.nIters)
	}
}
