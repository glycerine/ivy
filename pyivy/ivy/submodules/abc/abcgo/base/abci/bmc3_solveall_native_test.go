package abci

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// bmc3 -a (solve all outputs) on s713: output 0 fails in frame 0 and the
// same assignment fails other outputs too. Native crashed writing the
// duplicate's iPo through the "no CEX stored" sentinel (CBUGS_FIXED.md #19).
const bmc3SolveAllScript = "read test_vectors/seq/s713.blif; strash; bmc3 -F 4 -a; print_status"

var bmc3Timing = regexp.MustCompile(`Time =\s+[0-9.]+ sec`)

var bmc3Asserted = regexp.MustCompile(`^Output +(\d+) was asserted in frame +(\d+) \(solved +\d+ out of +\d+ outputs\)\.$`)

// bmc3Normalize drops timing, and reduces the "asserted" lines to the sorted
// set of failing outputs and frames. bmc3 itself is now a line-by-line port
// whose order equals C's (bmc3Exact); other commands' tests still use this.
func bmc3Normalize(out string) string {
	var lines, asserted []string
	for _, line := range strings.Split(bmc3Timing.ReplaceAllString(out, "Time = T"), "\n") {
		if strings.HasPrefix(line, "ABC command line") || strings.TrimSpace(line) == "" {
			continue
		}
		line = strings.TrimRight(line, " ")
		if m := bmc3Asserted.FindStringSubmatch(line); m != nil {
			asserted = append(asserted, fmt.Sprintf("output %3s frame %s", m[1], m[2]))
			continue
		}
		lines = append(lines, line)
	}
	sort.Strings(asserted)
	return strings.Join(append(asserted, lines...), "\n")
}

func TestNativeBmc3SolveAllDoesNotCrash(t *testing.T) {
	if out := runNativeABC(t, bmc3SolveAllScript); !strings.Contains(out, "Status array contains") {
		t.Fatalf("no status after bmc3 -a:\n%s", out)
	}
}

func TestBmc3SolveAllLikeNative(t *testing.T) {
	native := bmc3Exact(runNativeABC(t, bmc3SolveAllScript))
	port := bmc3Exact(runScript(t, bmc3SolveAllScript))
	if native != port {
		t.Fatalf("native:\n%s\nport:\n%s", native, port)
	}
}

// print_status reports C's iFrame: the last frame explored without a
// failure, -1 when frame 0 fails (the port reported the failing frame).
func TestBmc3FramesLikeNative(t *testing.T) {
	for _, script := range []string{
		"read test_vectors/native_s27_drw.blif; strash; bmc3; print_status",
		"read test_vectors/seq/s1494.blif; strash; bmc3 -S 2 -F 4; print_status",
	} {
		out := runScript(t, script)
		if !strings.Contains(out, "Frames = -1 ") {
			t.Errorf("%s:\n%s", script, out)
		}
	}
}
