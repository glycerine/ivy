package abci

import (
	"os"
	"regexp"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

// bmTime masks bmGateWay's "<phase> Time = 0.06" lines.
var bmTime = regexp.MustCompile(`Time = [0-9.]+`)

// bm and bm2 (Boolean matching) are combinational: on a sequential network
// the functional supports include latch outputs, which getDependencies
// (abcBm.c) used as indices into the per-PI array (heap overflow), and bm2
// crashed in saucyGateWay; the port panicked (CBUGS_FIXED.md #50). Both now
// refuse sequential networks; combinational ones behave as before.
func TestBmSequentialLikeNative(t *testing.T) {
	t.Cleanup(func() { os.Remove("IOmatch.txt") }) // bm writes its matching here
	for _, design := range []string{"test_vectors/native_s27_drw.blif", "test_vectors/native_mul4_gen.aig"} {
		for _, cmd := range []string{"bm", "bm -P", "bm2"} {
			script := "read " + design + "; strash; " + cmd
			// bm2 permutes with C's rand(); start from a fresh process's
			// stream, as the native run does
			util.C_Srand(1)
			native := bmTime.ReplaceAllString(printSymmNormalize(runNativeABC(t, script)), "Time = T")
			out, _ := runScriptStatus(t, script)
			if port := bmTime.ReplaceAllString(printSymmNormalize(out), "Time = T"); native != port {
				t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
			}
		}
	}
}
