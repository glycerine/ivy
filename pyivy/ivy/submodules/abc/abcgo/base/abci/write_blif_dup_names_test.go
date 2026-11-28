package abci

import (
	"path/filepath"
	"strings"
	"testing"
)

// Running l2s twice gives the network repeated CI names (two SAVE_BIERE
// inputs, two SAVED_LO latches, ...). Abc_NtkLogicToNetlist (abcNetlist.c)
// reports them with Abc_NtkCheckUniqueCiNames ("Abc_NtkCheck: Repeated CI
// names: X and X.") and then merges the repeated names into one net, so C's
// write_blif writes a BLIF whose circuit differs from the network. The port
// panicked ("permits one fanin"). It now prints the same warnings and then
// refuses to write ("Writing BLIF has failed."), a deliberate difference:
// it does not write the merged, wrong circuit.
func TestWriteBlifRepeatedCiNames(t *testing.T) {
	file := filepath.Join(t.TempDir(), "dup.blif")
	script := "read test_vectors/live/l2s_basic.blif; strash; l2s; l2s; write_blif " + file
	native := runNativeABC(t, script)
	out, _ := runScriptStatus(t, script)
	var nativeWarnings, portWarnings []string
	for _, line := range strings.Split(native, "\n") {
		if strings.HasPrefix(line, "Abc_NtkCheck: Repeated") {
			nativeWarnings = append(nativeWarnings, line)
		}
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "Abc_NtkCheck: Repeated") {
			portWarnings = append(portWarnings, line)
		}
	}
	if len(nativeWarnings) == 0 || strings.Join(nativeWarnings, "\n") != strings.Join(portWarnings, "\n") {
		t.Fatalf("repeated-name warnings differ:\nnative:\n%s\nport:\n%s", strings.Join(nativeWarnings, "\n"), strings.Join(portWarnings, "\n"))
	}
	if !strings.Contains(out, "Writing BLIF has failed.") {
		t.Fatalf("port did not refuse to write:\n%s", out)
	}
}
