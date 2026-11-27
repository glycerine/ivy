package abci

import (
	"strings"
	"testing"
)

// TestGiaFraigIterationsLikeNative: Cec_ManSatSolve answers each query with
// Cec_ManSatCheckNode (the frontier/supergate/MUX CNF of cecSolve.c), whose
// SAT models become the patterns that refine the classes. The port used a
// plain per-AND encoding, got other models, and refined differently from
// the second iteration on.
func TestGiaFraigIterationsLikeNative(t *testing.T) {
	out := runScript(t, "read ../../../i10.aig; &get; &fraig -v")
	for _, want := range []string{
		"  2 : P =      0. D =      7. F =     0. M =       0. And =    2370.",
		"  3 : P =    105. D =      4. F =     0. M =       0. And =    2251.",
		"Performed 360 SAT calls: P = 336  D = 24  F = 0",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}
