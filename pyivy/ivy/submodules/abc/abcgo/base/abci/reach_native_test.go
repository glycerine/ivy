package abci

import (
	"strings"
	"testing"
)

// &reachm (llb1Reach.c) passed the miter name twice to its report, so the
// frame was a pointer value; &reachn (llb3Nonlin.c) passed no name, so %s
// read nIters as a string ("(null)" for frame 0, a crash for a later one).
// CBUGS_FIXED.md #21. s27 fails in frame 0, as &reachy reports.
func TestNativeReachReportsFailingFrame(t *testing.T) {
	for _, command := range []string{"&reachm", "&reachn", "&reachy"} {
		out := runNativeABC(t, "read test_vectors/native_s27_drw.blif; strash; &get; "+command)
		if want := `Output 0 of miter "s27.bench" was asserted in frame 0.`; !strings.Contains(out, want) {
			t.Errorf("%s: want %q:\n%s", command, want, out)
		}
	}
}
