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

// bmc2 -v prints one line per unrolled interval of Saig_BmcPerform; the
// counterexample's bits and the interval lines (frames, nodes, SAT
// variables, conflicts) depend on the exact frames manager and SAT solver
// of bmcBmc2.c, so only the times and the memory estimate are masked.
var bmc2VerboseScripts = []string{
	"read test_vectors/seq/s713.blif; strash; bmc2 -v; print_status",
	"read test_vectors/seq/s713.blif; strash; dprove -v; print_status",
	"read test_vectors/seq/s713.blif; strash; bmc2 -F 5 -v; print_status",
	"read test_vectors/seq/s1494.blif; strash; bmc2 -v; print_status",
	"read test_vectors/seq/s1494.blif; strash; dprove -v; print_status",
	"read test_vectors/seq/s1494.blif; strash; bmc2 -F 5 -v; print_status",
	"read test_vectors/native_s27_drw.blif; strash; bmc2 -v; print_status",
	"read test_vectors/native_s27_drw.blif; strash; dprove -v; print_status",
	"read test_vectors/native_s27_drw.blif; strash; bmc2 -F 5 -v; print_status",
	// -u splits each output's top OR (Abc_NtkToDarBmc) into outputs
	"read test_vectors/seq/s27.blif; strash; bmc2 -v -u; print_status",
	"read test_vectors/seq/s1494.blif; strash; bmc2 -v -u; print_status",
	"read test_vectors/seq/s713.blif; strash; bmc3 -u; print_status",
	// miters against a resynthesized copy are never falsified, so these
	// cover several intervals (-D) and the frame-limit messages
	"read test_vectors/seq/s27.blif; strash; dc2; miter test_vectors/seq/s27.blif; bmc2 -v -F 30 -D 300; print_status",
	"read test_vectors/seq/s1494.blif; strash; dc2; miter test_vectors/seq/s1494.blif; bmc2 -v -F 12 -D 300; print_status",
	"read test_vectors/seq/s713.blif; strash; dc2; miter test_vectors/seq/s713.blif; bmc2 -v -F 10 -D 300; print_status",
	"read test_vectors/seq/s713.blif; strash; dc2; miter test_vectors/seq/s713.blif; bmc2 -v -F 30 -C 50; print_status",
	"read test_vectors/seq/s713.blif; strash; dc2; miter test_vectors/seq/s713.blif; bmc2 -v -F 30 -G 200; print_status",
}

var bmc2Seconds = regexp.MustCompile(`[0-9]+\.[0-9]+ sec`)
var bmc2Megabytes = regexp.MustCompile(`[0-9]+ MB`)

func bmc2VerboseNormalize(out string) string {
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "ABC command line") || strings.TrimSpace(line) == "" {
			continue
		}
		line = bmc2Megabytes.ReplaceAllString(bmc2Seconds.ReplaceAllString(line, "T sec"), "M MB")
		lines = append(lines, strings.TrimRight(line, " "))
	}
	return strings.Join(lines, "\n")
}

func TestBmc2VerboseLikeNative(t *testing.T) {
	for _, script := range bmc2VerboseScripts {
		if native, port := bmc2VerboseNormalize(runNativeABC(t, script)), bmc2VerboseNormalize(runScript(t, script)); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}
