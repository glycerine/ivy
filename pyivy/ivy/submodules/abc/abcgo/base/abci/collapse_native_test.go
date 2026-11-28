package abci

import (
	"strings"
	"testing"
)

// TestCollapseBddSizesLikeNative: collapse builds global BDDs with CUDD's
// automatic symmetric sifting (reordering inside cuddUniqueInter with the
// recursion's pending results referenced) and a final symmetric sifting;
// the BDD sizes equal native ABC's (11807 and 79647 nodes).
func TestCollapseBddSizesLikeNative(t *testing.T) {
	for _, c := range []struct{ script, want string }{
		{"read test_vectors/native_mul8_if6.blif; strash; collapse; print_stats", "bdd  = 11807"},
		{"read ../../test_vectors/i10.aig; strash; collapse; print_stats", "bdd  = 79647"},
	} {
		if out := runScript(t, c.script); !strings.Contains(out, c.want) {
			t.Errorf("%s: want %q in\n%s", c.script, c.want, out)
		}
	}
}
