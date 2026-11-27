package abci

import (
	"strings"
	"testing"
)

// TestCascadeKeepsNetworkLikeNative: Abc_NtkCascade builds the global BDDs
// with Abc_NtkBuildGlobalBdds (500000-node limit, reordering), runs the
// cascade experiment and returns a copy of the network whatever it found.
// The port built its own BDDs without reordering, blew the limit on i10
// and failed the command.
func TestCascadeKeepsNetworkLikeNative(t *testing.T) {
	out, status := runScriptStatus(t, "read ../../../i10.aig; strash; if -K 6; cascade; print_stats")
	if status != 0 || !strings.Contains(out, "and =   2675") {
		t.Fatalf("cascade failed (status %d):\n%s", status, out)
	}
}

// TestCascadeVerboseLikeNative: Abc_CascadeExperiment prints the reordered
// MTBDD size and reordering time; the decomposer reports "The LUT size is
// too small". (C then leaks the MTBDD and Extra_StopManager reports
// "The number of referenced nodes = 12556"; Go's Cudd_CheckZeroRef counts
// only externally referenced nodes, so that report is not compared.)
func TestCascadeVerboseLikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/native_mul8_if6.blif; strash; cascade -v")
	for _, want := range []string{"MTBDD           =  13434 nodes\nMTBDD reordered =  12556 nodes\nVariable reordering time = ", "\nThe LUT size is too small\n"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}
