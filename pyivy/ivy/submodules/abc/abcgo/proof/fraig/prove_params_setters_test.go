package fraig

import "testing"

func TestProveParamsSettersMatchCFields(t *testing.T) {
	var p Prove_Params_t
	Prove_ParamsSetDefault(&p)
	Prove_ParamsSetItersMax(&p, 5)
	Prove_ParamsSetVerbose(&p, 1)
	if p.nItersMax != 5 || p.fVerbose != 1 {
		t.Fatalf("got iterations=%d verbose=%d, want 5 and 1", p.nItersMax, p.fVerbose)
	}
}
