package abci

import "testing"

// renode -s maps with ISOP covers (Abc_NodeFromIf, abcIf.c). A cut whose
// cover is constant returns a new constant node, but the node already built
// from the cut's leaves keeps a function: C gives it Abc_SopCreateAnd. The
// port left it without one, so print_stats panicked ("node has no SOP") and
// sop/strash failed with "Converting to AIGs has failed" on i10.
func TestRenodeSopConstantCoverLikeNative(t *testing.T) {
	for _, script := range []string{
		"read ../../test_vectors/i10.aig; strash; renode -s; print_stats",
		"read ../../test_vectors/i10.aig; strash; renode -s; sop; strash; print_stats",
	} {
		native := printSymmNormalize(runNativeABC(t, script))
		out, _ := runScriptStatus(t, script)
		if port := printSymmNormalize(out); port != native {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}
