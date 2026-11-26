package abci

import (
	"strings"
	"testing"
)

// Native: after drw on a sequential AIG, the latches keep their names
// (Abc_NtkFromDar starts the new network from the old one), so dsec can
// miter the result against the original.
func TestDrwKeepsLatchNamesLikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s27.blif; strash; drw; dsec")
	if !strings.Contains(out, "Networks are equivalent.  Time =") {
		t.Fatalf("dsec after drw:\n%s", out)
	}
}
