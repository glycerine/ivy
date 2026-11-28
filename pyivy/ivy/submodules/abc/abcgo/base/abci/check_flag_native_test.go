package abci

import (
	"path/filepath"
	"strings"
	"testing"
)

// Abc_NtkCheck (abcCheck.c) checks a network only when the "check" flag is
// enabled; it is off by default. The port always checked, so on the l2s;
// l2s network (repeated CI names) it printed "Abc_NtkAigToLogicSop():
// Network check has failed." and "Abc_NtkCleanup: The network check has
// failed." where ABC prints nothing, and it paid for the check after every
// command. With "set check", Abc_NtkCheckNames reports the repeated names
// through Abc_NtkCheckUniqueCiNames, as in C.
func TestNtkCheckHonorsCheckFlagLikeNative(t *testing.T) {
	file := filepath.Join(t.TempDir(), "dup.blif")
	body := "read test_vectors/live/l2s_basic.blif; strash; l2s; l2s; write_blif " + file
	checkLines := func(out string) string {
		var lines []string
		for _, line := range strings.Split(out, "\n") {
			if strings.HasPrefix(line, "Abc_NtkCheck:") || strings.Contains(line, "check has failed") {
				lines = append(lines, line)
			}
		}
		return strings.Join(lines, "\n")
	}
	for _, prefix := range []string{"", "set check; "} {
		script := prefix + body
		native := runNativeABC(t, script)
		out, _ := runScriptStatus(t, script)
		if prefix != "" {
			// C goes on to build the two-driver netlist and fails there
			// ("Abc_NtkLogicToNetlist(): Network check has failed."); the
			// port stops before it (see TestWriteBlifRepeatedCiNames).
			native = strings.Replace(native, "Abc_NtkLogicToNetlist(): Network check has failed.\n", "", 1)
		}
		if n, g := checkLines(native), checkLines(out); n != g {
			t.Errorf("%q: check output differs:\nnative:\n%s\nport:\n%s", prefix, n, g)
		}
	}
}
