package abci

import (
	"regexp"
	"strings"
	"testing"
)

// memory figures depend on the allocator, and every "N.NN sec" is a time
var (
	mbTimes  = regexp.MustCompile(`[0-9]+\.[0-9]+ MB`)
	secTimes = regexp.MustCompile(`[0-9]+\.[0-9]+ sec`)
	percents = regexp.MustCompile(`\( *[0-9.]+ %\)`)
)

// Times are printed with Abc_PrintTime ("Time =%9.2f sec"); several ports
// printed a Go duration ("Time = 7.92ms", "Time: 1.2s") and omitted lines
// ABC prints (&jf's "Computing cuts..." and "Memory:", &lf/&nf's mux stats).
func TestVerboseReportsLikeNative(t *testing.T) {
	for _, script := range []string{
		"read ../../../i10.aig; &get; &jf -v; &ps",
		"read ../../../i10.aig; &get; &kf -v; &ps",
		"read ../../../i10.aig; &get; &lf -v; &ps",
		"read test_vectors/native_s27_drw.blif; strash; &get; &sim",
		"read test_vectors/seq/s713.blif; strash; &get; &sim -F 4",
		"read ../../../i10.aig; drwsat -v; print_stats",
		"read ../../../i10.aig; drwsat -v -b; print_stats",
		"read test_vectors/seq/s713.blif; strash; scorr -v; print_stats",
		"read test_vectors/seq/s713.blif; strash; lcorr -v; print_stats",
		"read test_vectors/seq/s1494.blif; strash; lcorr -v; print_stats",
	} {
		port := bmc3Normalize(runScript(t, script))
		if strings.Contains(port, "ms\n") || strings.Contains(port, "Time: ") {
			t.Errorf("%s: Go duration in output:\n%s", script, port)
		}
		native := bmc3Normalize(runNativeABC(t, script))
		clean := func(s string) string {
			return percents.ReplaceAllString(secTimes.ReplaceAllString(mbTimes.ReplaceAllString(s, "M MB"), "T sec"), "(P %)")
		}
		if clean(native) != clean(port) {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}

// drwsat -v labels its stages as Dar_ManRwsat does and prints the rewrite
// and refactor statistics; the port printed "Starting: 3157 nodes".
func TestDrwsatVerboseStages(t *testing.T) {
	out := runScript(t, "read ../../../i10.aig; drwsat -v; print_stats")
	for _, want := range []string{"Starting:  ../../../i10", "Rewrite:   ../../../i10", "Refactor:  ../../../i10", "NodesBeg =     2123."} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
}
