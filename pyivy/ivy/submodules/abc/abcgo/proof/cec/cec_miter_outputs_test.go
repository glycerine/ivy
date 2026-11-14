package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

func TestMiterOutputAnalysisMatchesSourcePhaseNormalization(t *testing.T) {
	for _, dual := range []bool{false, true} {
		g := gia.Gia_ManStart(4)
		gia.Gia_ManAppendCo(g, 0)
		if dual {
			gia.Gia_ManAppendCo(g, 1)
		}
		pars := Cec_ParSim_t{NWords: 1, FCheckMiter: 1}
		if dual {
			pars.FDualOut = 1
		}
		sim := Cec_ManSimStart(g, &pars)
		if dual {
			cecInfo(sim.vCoSimInfo, 1)[0] = ^uint32(0)
		} else {
			cecInfo(sim.vCoSimInfo, 0)[0] = ^uint32(0)
		}
		if got := Cec_ManSimAnalyzeOutputs(sim); got != 0 || sim.pCexComb != nil {
			t.Fatalf("dual %v: got %d witness %+v", dual, got, sim.pCexComb)
		}
	}
}
