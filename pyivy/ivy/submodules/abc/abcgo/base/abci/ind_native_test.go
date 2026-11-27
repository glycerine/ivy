package abci

import (
	"strings"
	"testing"
)

// ind: when the frame limit is reached with the induction step still
// satisfiable, Saig_ManInduction leaves RetValue at -1 (undecided). The
// port returned 0, so 'ind' reported "Networks are NOT EQUIVALENT" for a
// property it had merely failed to prove.
func TestInductionFrameLimitIsUndecided(t *testing.T) {
	for _, frames := range []string{"1", "2", "5"} {
		script := "read test_vectors/native_s27_drw.blif; strash; ind -v -F " + frames + "; print_status"
		out := runScript(t, script)
		for _, want := range []string{
			"Completed " + frames + " iterations.\n",
			"Networks are UNDECIDED.",
			"Status = -1 ",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("%s: missing %q:\n%s", script, want, out)
			}
		}
	}
}

func TestInductionLikeNative(t *testing.T) {
	for _, script := range []string{
		"read test_vectors/native_s27_drw.blif; strash; ind -v -F 2; print_status",
		"read test_vectors/seq/s713.blif; strash; ind -v -F 3; print_status",
		"read test_vectors/seq/s1494.blif; strash; ind -v -u -F 3; print_status",
	} {
		if native, port := bmc3Normalize(runNativeABC(t, script)), bmc3Normalize(runScript(t, script)); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}
