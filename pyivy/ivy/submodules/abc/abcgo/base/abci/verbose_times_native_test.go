package abci

import (
	"regexp"
	"strings"
	"testing"
)

// memory figures depend on the allocator
var mbTimes = regexp.MustCompile(`[0-9]+\.[0-9]+ MB`)

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
	} {
		port := bmc3Normalize(runScript(t, script))
		if strings.Contains(port, "ms\n") || strings.Contains(port, "Time: ") {
			t.Errorf("%s: Go duration in output:\n%s", script, port)
		}
		native := bmc3Normalize(runNativeABC(t, script))
		clean := func(s string) string { return mbTimes.ReplaceAllString(s, "M MB") }
		if clean(native) != clean(port) {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}
