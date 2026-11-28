package abci

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glycerine/abcgo/misc/util"
)

// bmSessionRun runs script natively and in process, each from a fresh
// rand() stream, and returns both outputs (times masked) with the
// IOmatch.txt that bm wrote appended.
func bmSessionRun(t *testing.T, script string) (native, port string) {
	t.Helper()
	ioMatch := func() string {
		data, err := os.ReadFile("IOmatch.txt")
		os.Remove("IOmatch.txt")
		if err != nil {
			return "\n(no IOmatch.txt)"
		}
		return "\nIOmatch.txt:\n" + string(data)
	}
	os.Remove("IOmatch.txt")
	native = bmTime.ReplaceAllString(printSymmNormalize(runNativeABC(t, script)), "Time = T") + ioMatch()
	util.C_Srand(1) // a fresh native process starts from srand(1)
	out, _ := runScriptStatus(t, script)
	port = bmTime.ReplaceAllString(printSymmNormalize(out), "Time = T") + ioMatch()
	return native, port
}

// bm (abcBm.c) is a line-by-line port: the functional supports come from
// Sim_ComputeFunSupp on the strashed network, the simulation refinement
// draws SIM_RANDOM_UNSIGNED per input group, and the SAT search builds
// the same cofactors and miters, so the matching, the IOmatch.txt file and
// the rand() state left for a following bm2 (saucy) are C's.
func TestBmSessionLikeNative(t *testing.T) {
	t.Cleanup(func() { os.Remove("IOmatch.txt") })
	permuted := filepath.Join(t.TempDir(), "permuted.aig")
	type tc struct {
		design string
		cmds   []string
	}
	all := []string{"bm", "bm -P", "bm " + permuted, "bm -P " + permuted}
	cases := []tc{
		{"test_vectors/native_mul4_gen.aig", append(all, "bm; bm", "bm; bm2", "bm -P; bm2")},
		{"test_vectors/native_mul8_canon.aig", all},
		{"test_vectors/i10_native_demo.aig", all[:2]},
	}
	for _, c := range cases {
		for _, cmd := range c.cmds {
			script := "read " + c.design + "; strash; dc2; write_aiger " + permuted + "; read " + c.design + "; strash; " + cmd
			native, port := bmSessionRun(t, script)
			if native != port {
				t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
			}
		}
	}
}
