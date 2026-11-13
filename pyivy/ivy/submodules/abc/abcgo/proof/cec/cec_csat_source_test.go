package cec

import (
	"github.com/glycerine/abcgo/aig/gia"
	"testing"
)

// cecSolve.c:837-847 unconditionally checks that every CBS counterexample
// record was consumed, including after the fCheckMiter early break.
func TestCircuitSATCheckMiterAssertsOnTrailingRecords(t *testing.T) {
	g := gia.Gia_ManStart(8)
	a, b := gia.Gia_ManAppendCi(g), gia.Gia_ManAppendCi(g)
	gia.Gia_ManAppendCo(g, a)
	gia.Gia_ManAppendCo(g, b)
	var p Cec_ParSat_t
	Cec_ManSatSetDefaultParams(&p)
	p.fCheckMiter = 1
	defer func() {
		if recover() == nil {
			t.Fatal("source requires trailing-store assertion")
		}
	}()
	Cec_ManSatSolveCSat(nil, g, &p)
}
