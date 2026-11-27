package abci

import (
	"regexp"
	"strings"
	"testing"
)

var otherNumbers = regexp.MustCompile(` *[0-9]+`)

// timeLines keeps the lines of out that report a time, with the times,
// memory and percentages masked and every other number replaced by '#'
// (solver statistics and their widths may differ).
func timeLines(out string) string {
	var keep []string
	for _, line := range strings.Split(bmc3Normalize(out), "\n") {
		if strings.Contains(line, " sec") || strings.Contains(line, "Time") || strings.Contains(line, "time") {
			line = percents.ReplaceAllString(secTimes.ReplaceAllString(mbTimes.ReplaceAllString(line, "M MB"), "T sec"), "(P %)")
			keep = append(keep, otherNumbers.ReplaceAllString(line, " #"))
		}
	}
	return strings.Join(keep, "\n")
}

// compareTimeLinesWithNative requires the port's time-reporting lines to
// equal native ABC's where the rest of the output (e.g. the SAT model
// chosen) may legitimately differ.
func compareTimeLinesWithNative(t *testing.T, script string) {
	t.Helper()
	port := timeLines(runScript(t, script))
	native := timeLines(runNativeABC(t, script))
	if native != port {
		t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
	}
}

// sat/bmc commands printed Go durations ("Time 520.807µs", "Time: 1.2s")
// where ABC prints Abc_PrintTime's "Time =     0.00 sec".
func TestVerboseTimesBmcLikeNative(t *testing.T) {
	for _, script := range []string{
		"read_truth 8; strash; &get; &mesh",
		"read_truth 8; strash; &get; &mesh -s",
		"read_truth 6; strash; &get; &mesh -X 3 -Y 3 -T 2",
		"read_truth 6; strash; &get; &mesh -X 3 -Y 3 -T 2 -s",
		"miter test_vectors/native_s27_drw.blif test_vectors/native_s27_drw.blif; &get; &bcore -v -F 2 /dev/null",
		"read test_vectors/native_s27_drw.blif; strash; &get; &fftest -A 3",
		"read test_vectors/native_s27_drw.blif; strash; &get; &fftest -A 2",
		"read test_vectors/native_s27_drw.blif; strash; &get; &fftest -A 3 -N 4",
		"majexact -I 3 -N 1 -g",
		"majexact -I 5 -N 3 -g",
		"lutexact -I 4 -N 2 -K 3 8778",
		"lutexact -I 3 -N 1 -K 3 -g E8",
		"read test_vectors/native_s27_drw.blif; strash; &get; &sattest -v",
		"read test_vectors/seq/s713.blif; strash; &get; &inse -v",
		"read test_vectors/seq/s713.blif; strash; &get; &maxi -v",
		"read test_vectors/native_s27_drw.blif; strash; &get; &maxi -v",
	} {
		compareTimeLinesWithNative(t, script)
	}
}
