package abci

import (
	"strings"
	"testing"
)

// Acec_OrderTreeRoots (acecXor.c) looks for the adder tree that no other
// tree feeds; when there is none (two independent multipliers, as in this
// dual-output miter) the assert is compiled out and the loop reads one
// entry past vXorRoots, and Acec_RankTrees then writes through the garbage
// root: '&acec -m' segfaulted (C bug 32). The port panicked at the same point.
func TestAcecOrderRootsWithoutChainLikeNative(t *testing.T) {
	for _, script := range []string{
		"&r test_vectors/acec_mul4_dual_miter.aig; &acec -m",
		"&r test_vectors/acec_mul4_dual_miter.aig; &acec -m -v",
	} {
		native := runNativeABC(t, script)
		got, _ := runScriptStatus(t, script)
		if !strings.Contains(native, "Networks are") {
			t.Errorf("%s: native did not finish:\n%s", script, native)
		}
		if bmc3Normalize(got) != bmc3Normalize(native) {
			t.Errorf("%s:\nnative:\n%s\ngo:\n%s", script, native, got)
		}
	}
}
