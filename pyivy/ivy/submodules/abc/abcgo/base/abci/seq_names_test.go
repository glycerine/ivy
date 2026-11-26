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

// Native: abc -c "read mul8.aig; &get; &sweep; &put; print_stats".
func TestGiaSweepDefaultsMatchNative(t *testing.T) {
	expectStats(t, "read test_vectors/bench/mul8.aig; &get; &sweep; &put; print_stats",
		"i/o =   16/   16  lat =    0  and =    424  lev = 53")
}

// Native: abc -c "read mul8.aig; strash; fraig_store; balance; fraig_store;
// fraig_restore; print_stats". Storing a strashed network restrashes it.
func TestFraigStoreRestoreMatchesNative(t *testing.T) {
	expectStats(t, "read test_vectors/bench/mul8.aig; strash; fraig_store; balance; fraig_store; fraig_restore; print_stats",
		"i/o =   16/   16  lat =    0  and =    424  lev = 53")
}

// Native (release) ABC: on a combinational network absec and simsec report
// that demitering failed; their register-count assertions are compiled out.
func TestSecOnCombinationalNetworkLikeNative(t *testing.T) {
	for _, cmd := range []string{"absec", "simsec"} {
		out, _ := runScriptStatus(t, "read test_vectors/bench/mul8.aig; strash; "+cmd)
		if !strings.Contains(out, "Demitering has failed.") {
			t.Errorf("%s:\n%s", cmd, out)
		}
	}
}

// Native: abc -c "read mul8.aig; strash; dframes; print_stats". ABC's
// Saig_ManTimeframeSimplify has no register precondition.
func TestDframesCombinationalLikeNative(t *testing.T) {
	expectStats(t, "read test_vectors/bench/mul8.aig; strash; dframes; print_stats",
		"i/o =   80/   80  lat =    0  and =   2120  lev = 53")
}

// &print_truth computes truth tables over all CIs; with registers s27 has
// 7 and is refused (CBUGS_FIXED.md #5: ABC read past its tables).
func TestPrintTruthRefusesMoreThanSixCis(t *testing.T) {
	out, _ := runScriptStatus(t, "read test_vectors/seq/s27.blif; strash; &get; &print_truth")
	if !strings.Contains(out, "The number of inputs of the AIG exceeds 6.") {
		t.Fatalf("&print_truth on s27:\n%s", out)
	}
}

// Native: abc -c "read s27.blif; strash; &get; &setregnum; &ps": the
// command overwrites the register count (here with 0).
func TestSetRegNumOverwritesLikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s27.blif; strash; &get; &setregnum; &ps")
	if !strings.Contains(out, "i/o =      7/      4  and =       8  lev =    5 (3.75)") {
		t.Fatalf("&setregnum:\n%s", out)
	}
}

// Native: abc -c "read s27.blif; strash; &get; &atree -v". Release ABC runs
// on sequential AIGs and prints C's "inf" for a division by zero adders.
func TestAtreeSequentialLikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s27.blif; strash; &get; &atree -v")
	if !strings.Contains(out, "Hashed/Adders = inf.") || !strings.Contains(out, "   1 : Supp =   6  Cone =   7  \n") {
		t.Fatalf("&atree on s27:\n%s", out)
	}
}

// Native: abc -c "read s27.blif; strash; &get; &satenum" enumerates the 16
// PI assignments making some output 1 (giaQbf.c has no single-output or
// combinational precondition).
func TestSatEnumLikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s27.blif; strash; &get; &satenum")
	if !strings.Contains(out, "Enumerated the complete set of 16 assignments.  Time =") {
		t.Fatalf("&satenum on s27:\n%s", out)
	}
}

// Native: abc -c "read mul8.aig; strash; inter". Interpolating the network
// against itself leaves dangling nodes, which release ABC's
// Abc_NtkFromDar tolerates.
func TestInterOnMul8DoesNotPanic(t *testing.T) {
	out, _ := runScriptStatus(t, "read test_vectors/bench/mul8.aig; strash; inter")
	if !strings.Contains(out, "Error: Deriving new circuit structure for the current network.") {
		t.Fatalf("inter on mul8:\n%s", out)
	}
}

// Native: abc -c "read and2x2.blif; strash; saucy3" (f = a&b, g = b&c).
// Sim_ManStop prints the support statistics, and the colors array is sized
// by the number of outputs and inputs.
func TestSaucyOnSmallDesignLikeNative(t *testing.T) {
	out, _ := runScriptStatus(t, "read test_vectors/seq/and2x2.blif; strash; saucy3")
	for _, want := range []string{"Total func supps   =        4.", "Sat runs UNSAT     =        0.", "average support = 4.00", "Runtime =     0.00 sec"} {
		if !strings.Contains(out, want) {
			t.Fatalf("saucy3 missing %q:\n%s", want, out)
		}
	}
}

// Native: abc -c "read s1494.blif; strash; irws; print_stats" gives
// lat = 117, and = 599. Ivy_NodeRewriteSeq re-prepares the best subgraph
// with the best cut's leaves and latches; the port kept those of the last
// cut evaluated with the shared subgraph and overstated the gain.
func TestIrwsOnS1494LikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s1494.blif; strash; irws; print_stats")
	if got := statsLine(out); !strings.Contains(got, "lat =  117") || !strings.Contains(got, "and =    599") {
		t.Fatalf("irws on s1494: %q", got)
	}
}

// Native gives lat = 17, and = 136 on s713. Ivy_GraphToNetworkSeqCountSeq
// stops walking a fanin's latches once one is not in the table; the port
// kept going and built a latch ghost from the missing (nil) node.
func TestIrwsOnS713LikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s713.blif; strash; irws; print_stats")
	if got := statsLine(out); !strings.Contains(got, "lat =   17") || !strings.Contains(got, "and =    136") {
		t.Fatalf("irws on s713: %q", got)
	}
}
