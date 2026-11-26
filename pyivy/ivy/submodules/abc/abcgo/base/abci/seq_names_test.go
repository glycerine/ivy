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

// Native: abc -c "read seq/constraint1.aig; print_stats; &get; &put;
// print_stats". The AIGER file has one constraint (the C header field);
// ABC keeps it through reading and &get/&put.
func TestAigerConstraintsKeptLikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/constraint1.aig; print_stats; &get; &put; print_stats")
	if n := strings.Count(out, "i/o =    2/    2(c=1)  lat =    1  and =      1  lev =  1"); n != 2 {
		t.Fatalf("constraint lost (%d of 2 stats lines):\n%s", n, out)
	}
}

// Native: after &put the network has no external spec (Abc_NtkFromAigPhase
// leaves pSpec unset), so dsec without arguments reports that.
func TestPutLeavesNoSpecLikeNative(t *testing.T) {
	out, _ := runScriptStatus(t, "read test_vectors/seq/s27.blif; strash; &get; &dc2; &put; dsec")
	if !strings.Contains(out, "The external spec is not given.") {
		t.Fatalf("dsec after &put:\n%s", out)
	}
}
