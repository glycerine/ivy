package abci

import (
	"path/filepath"
	"regexp"
	"testing"
)

// Reports from misc/extra (majgen, varmin), opt/sbd (&mfsd) and base/exor
// (&exorcism) print times with Abc_PrintTime; the ports printed Go durations
// ("Time = 49µs") and dropped text ABC prints ("Total =", "Covering time",
// &mfsd's per-stage ABC_PRTP lines, the exorcism stage lines).
func TestVerboseTimesMiscLikeNative(t *testing.T) {
	for _, script := range []string{
		"majgen -N 6",
		"varmin -M 3 -N 8",
		"varmin -M 3 -N 8 -v",
		"varmin -M 3 -N 8 -o -v",
		"read test_vectors/native_s27_drw.blif; strash; &get; &mfsd -v",
	} {
		compareTimesWithNative(t, script)
	}
}

// &exorcism -V 1 prints Exorcism()'s stage report. The port's minimizer is
// not ABC's, so the number of passes (one '.' each) may differ; every
// other line must match.
func TestExorcismVerboseLikeNative(t *testing.T) {
	script := "read test_vectors/native_s27_drw.blif; strash; &get; &exorcism -V 1 " + filepath.Join(t.TempDir(), "out.esop")
	dots := regexp.MustCompile(`(?m)^\.+$`)
	clean := func(s string) string {
		s = dots.ReplaceAllString(bmc3Normalize(s), "...")
		return percents.ReplaceAllString(secTimes.ReplaceAllString(mbTimes.ReplaceAllString(s, "M MB"), "T sec"), "(P %)")
	}
	port := clean(runScript(t, script))
	native := clean(runNativeABC(t, script))
	if native != port {
		t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
	}
}
