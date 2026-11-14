package abci

import "testing"

func TestProveParamsDefaultsMatchSource(t *testing.T) {
	var p Prove_Params_t
	Prove_ParamsSetDefault(&p)
	if p.fUseFraiging != 1 || p.fUseRewriting != 1 || p.fUseBdds != 0 || p.fVerbose != 0 {
		t.Fatalf("general defaults = %+v", p)
	}
	if p.nItersMax != 6 || p.nMiteringLimitStart != 5000 || p.nMiteringLimitMulti != 2 {
		t.Fatalf("miter defaults = %+v", p)
	}
	if p.nRewritingLimitStart != 3 || p.nRewritingLimitMulti != 1 || p.nFraigingLimitStart != 2 || p.nFraigingLimitMulti != 8 {
		t.Fatalf("rewrite/fraig defaults = %+v", p)
	}
	if p.nBddSizeLimit != 1_000_000 || p.fBddReorder != 1 || p.nMiteringLimitLast != 0 || p.nTotalBacktrackLimit != 0 || p.nTotalInspectLimit != 0 {
		t.Fatalf("last-gasp/resource defaults = %+v", p)
	}
}
