package abci

import (
	"regexp"
	"strings"
	"testing"
)

// Ivy_FraigProve (ivyFraig.c) prints its resource limits, an ITERATION line
// per round and Ivy_FraigMiterPrint's 'Nodes = ... Levels = ...' after each
// SAT/fraiging stage under -v; the port printed none of them (iprove -v,
// dprove -v). prove -v printed extra blank lines.
func TestProveVerboseLikeNative(t *testing.T) {
	times := regexp.MustCompile(`[0-9]+\.[0-9]+ ?(sec|MB|Mb|%)`)
	for _, pair := range [][2]string{
		{"test_vectors/native_mul8_canon.aig", "test_vectors/cec_mul8_syn2.aig"},
		{"test_vectors/native_mul8_canon.aig", "test_vectors/cec_mul8_mut.aig"},
		{"test_vectors/native_mul4_gen.aig", "test_vectors/cec_mul4_syn2.aig"},
	} {
		for _, cmd := range []string{"iprove -v", "dprove -v", "prove -v"} {
			script := "miter " + pair[0] + " " + pair[1] + "; " + cmd
			native := runNativeABC(t, script)
			// drop native's 'ABC command line' header and its blank line;
			// blank lines are compared (C's fflush was once a newline)
			if i := strings.Index(native, "\n\n"); strings.HasPrefix(native, "ABC command line") && i >= 0 {
				native = native[i+2:]
			}
			native = times.ReplaceAllString(native, "X")
			got, _ := runScriptStatus(t, script)
			got = times.ReplaceAllString(got, "X")
			if got != native {
				t.Errorf("%s:\nnative:\n%s\ngo:\n%s", script, native, got)
			}
		}
	}
}
