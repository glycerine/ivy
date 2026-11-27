package abci

import "testing"

// kcs (kliveness.c) needs an output named csLiveTarget_*: collectCSSignals
// asserts one, and without it (release build) createConstrained0LiveCone
// dereferenced NULL (CBUGS_FIXED.md #47). Both now report the missing
// target.
func TestKcsWithoutTargetLikeNative(t *testing.T) {
	for _, prep := range []string{"strash", "logic"} {
		for _, cmd := range []string{"kcs", "kcs -c", "kcs -m", "kcs -C", "kcs -g"} {
			script := "read test_vectors/native_s27_drw.blif; " + prep + "; " + cmd
			native := printSymmNormalize(runNativeABC(t, script))
			out, _ := runScriptStatus(t, script)
			if port := printSymmNormalize(out); native != port {
				t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
			}
		}
	}
}
