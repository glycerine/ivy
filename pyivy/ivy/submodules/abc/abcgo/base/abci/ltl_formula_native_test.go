package abci

import (
	"os"
	"path/filepath"
	"testing"
)

// Io_ReadBlifMv (ioReadBlifMv.c) pushes each .ltlformula into
// vGlobalLtlArray while parsing, but allocated that array only after
// parsing (and then copied the empty array into the network): reading a
// BLIF file with an LTL formula crashed native ABC (CBUGS_FIXED.md #58), and
// no formula ever reached the network. The port rejected the directive.
//
// With the formulas read, l3s (liveness.c's LTL path) builds their monitors:
// ltl_formula.blif names an input (the signal check fails), ltl_gf.blif's
// ":G :F o" is translated.
func TestBlifLtlFormulaLikeNative(t *testing.T) {
	dir := t.TempDir()
	for _, design := range []string{"test_vectors/live/ltl_formula.blif", "test_vectors/live/ltl_gf.blif"} {
		for _, cmd := range []string{"print_stats", "strash; print_stats", "strash; l3s; print_stats", "strash; l3s -s; print_stats", "strash; l3s -1; print_stats"} {
			goFile, nativeFile := filepath.Join(dir, "go.aig"), filepath.Join(dir, "native.aig")
			script := "read " + design + "; " + cmd + "; write_aiger "
			native := printSymmNormalize(runNativeABC(t, script+nativeFile))
			out, _ := runScriptStatus(t, script+goFile)
			if port := printSymmNormalize(out); native != port {
				t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
			}
			g, _ := os.ReadFile(goFile)
			n, _ := os.ReadFile(nativeFile)
			if aigerStamp.ReplaceAllString(string(g), "") != aigerStamp.ReplaceAllString(string(n), "") {
				t.Errorf("%s: written AIG differs from native", script)
			}
		}
	}
}
