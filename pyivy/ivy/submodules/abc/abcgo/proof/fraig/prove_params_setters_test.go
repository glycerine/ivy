package fraig

import "testing"

func TestProveParamsSettersMatchCFields(t *testing.T) {
	var p Prove_Params_t
	Prove_ParamsSetDefault(&p)
	Prove_ParamsSetItersMax(&p, 5)
	Prove_ParamsSetMiteringLimitLast(&p, 17)
	Prove_ParamsSetVerbose(&p, 1)
	if p.nItersMax != 5 || p.nMiteringLimitLast != 17 || p.fVerbose != 1 {
		t.Fatalf("got iterations=%d mitering-last=%d verbose=%d, want 5, 17, and 1", p.nItersMax, p.nMiteringLimitLast, p.fVerbose)
	}
}
