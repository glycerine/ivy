package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// cecCore.c:552-555 changes iOutFail only when a miter failure occurred.
func TestSourceSweepPreservesFailureIndexWithoutFailure(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	n := gia.Gia_ManAppendAnd(g, a, b)
	gia.Gia_ManAppendCo(g, n)
	var pars Cec_ParFra_t
	Cec_ManFraSetDefaultParams(&pars)
	pars.NItersMax = 0
	pars.NWords = 1
	pars.NRounds = 1
	pars.iOutFail = 123
	out := Cec_ManSatSweeping(g, &pars, 1)
	if pars.Error != nil || out == nil {
		t.Fatalf("sweep %v %v", out, pars.Error)
	}
	if pars.iOutFail != 123 {
		t.Fatalf("failure index changed to %d", pars.iOutFail)
	}
}
