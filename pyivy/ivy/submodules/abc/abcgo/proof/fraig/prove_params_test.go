package fraig

import (
	"bytes"
	"strings"
	"testing"
)

func TestProveParamsDefaultsAndReport(t *testing.T) {
	p := Prove_Params_t{fVerbose: 1, fUseBdds: 1, nTotalBacktracksMade: 42, nTotalInspectsMade: 43}
	Prove_ParamsSetDefault(&p)
	want := Prove_Params_t{fUseFraiging: 1, fUseRewriting: 1, nItersMax: 6,
		nMiteringLimitStart: 5000, nMiteringLimitMulti: 2,
		nRewritingLimitStart: 3, nRewritingLimitMulti: 1,
		nFraigingLimitStart: 2, nFraigingLimitMulti: 8,
		nBddSizeLimit: 1000000, fBddReorder: 1}
	if p != want {
		t.Fatalf("defaults = %+v, want %+v", p, want)
	}
	var out bytes.Buffer
	proveParamsPrint(&out, &p)
	for _, line := range []string{"Fraiging enabled: yes", "BDD construction enabled: no", "Solver iterations: 6", "Starting mitering limit: 5000", "Starting number of conflicts in fraiging: 8.00", "BDD reordering enabled: yes", "Parameter dump complete."} {
		if !strings.Contains(out.String(), line+"\n") {
			t.Errorf("missing %q in %q", line, out.String())
		}
	}
}
