package abci

import (
	"strings"
	"testing"
)

// &bmc maps the unrolled AIG with Mf and derives CNF from the mapping.
// On s1494 a 4-leaf cut computes the buffer of its first leaf, whose CNF
// was cached for one variable; Mf_ManDeriveCnf put the output literal at
// position 4 and wrote past the literal array (CBUGS_FIXED.md #20):
// "malloc(): invalid size", or under ASan a heap-buffer-overflow.
func TestNativeAbc9BmcDoesNotCorruptHeap(t *testing.T) {
	out := runNativeABC(t, "read test_vectors/seq/s1494.blif; strash; &get; &bmc -F 4; print_status")
	if !strings.Contains(out, "Status =") {
		t.Fatalf("no status after &bmc:\n%s", out)
	}
}
