package abci

import (
	"regexp"
	"strings"
	"testing"
)

// bmc2 (Saig_BmcPerform) prints its result, the time and, when undecided,
// why it stopped; the port printed nothing and reported other frame counts.

const bmc2Miter = "miter test_vectors/seq/s713.blif test_vectors/seq/s713.blif; "

var bmc2Scripts = []string{
	"read test_vectors/seq/s713.blif; strash; bmc2 -F 4; print_status",
	"read test_vectors/seq/s713.blif; strash; bmc2 -S 2 -F 4; print_status",
	bmc2Miter + "bmc2 -F 3; print_status",
	bmc2Miter + "bmc2 -S 2 -F 4; print_status",
}

// which inputs the counterexample sets depends on the SAT model
var cexOnes = regexp.MustCompile(`1s = +\d+ \( *[0-9.]+ %\)`)

func bmc2Normalize(out string) string {
	return cexOnes.ReplaceAllString(bmc3Normalize(out), "1s = N")
}

func TestBmc2ReportsLikeNative(t *testing.T) {
	for _, script := range bmc2Scripts {
		if native, port := bmc2Normalize(runNativeABC(t, script)), bmc2Normalize(runScript(t, script)); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}

func TestBmc2Reports(t *testing.T) {
	for i, want := range [][]string{
		{`Output 0 of miter "s713.bench" was asserted in frame 0. Time =`, "Frames = -1 "},
		{`Output 0 of miter "s713.bench" was asserted in frame 2. Time =`, "Frames = 1 "},
		{"No output failed in 2 frames.  Time =", "Reached limit on the number of timeframes (3).", "Frames = 2 "},
		{"No output failed in 3 frames.  Time =", "Reached limit on the number of timeframes (4).", "Frames = 3 "},
	} {
		out := runScript(t, bmc2Scripts[i])
		for _, w := range want {
			if !strings.Contains(out, w) {
				t.Errorf("%s: missing %q:\n%s", bmc2Scripts[i], w, out)
			}
		}
	}
}
