package abci

import (
	"regexp"
	"testing"
)

// &satlut (giaSatLut.c) was a reimplementation: it built a fresh
// cardinality solver for every window instead of rolling back to the
// bookmark, never set the cut polarities (sat_solver_set_polarity), and
// marked the solver permanently UNSAT on a failed critical-path clause, so
// it picked other windows' solutions (i10 '&if -K 4; &satlut': lev 15
// instead of 14). It is now a line-by-line port.
var satlutPercent = regexp.MustCompile(`sec \( *[0-9.]+ %\)`)

func satlutNormalize(out string) string {
	return bmc3Normalize(satlutPercent.ReplaceAllString(regexp.MustCompile(`=\s+[0-9.]+ sec`).ReplaceAllString(out, "= T sec"), "sec (P %)"))
}

func TestSatlutLikeNative(t *testing.T) {
	for _, design := range []string{"i10_native_demo", "native_mul8_canon", "native_mul4_gen"} {
		for _, script := range []string{
			"&if -K 4; &satlut; &ps",
			"&if -K 4; &satlut -d; &ps",
			"&if -K 4; &edge; &satlut -d -v; &ps",
			"&if -K 4; &edge -E 2; &satlut -d -Q 2; &ps",
			"&if -K 6; &edge; &satlut -d; &ps",
			"&if -K 6; &satlut -N 64 -r; &ps",
			"&if -K 3; &satlut -I 5 -w; &ps",
			"&if -K 4; &satlut -C 10 -D 8 -d; &ps",
			"&if -K 5; &satlut -N 16 -v; &ps",
		} {
			script = "&r test_vectors/" + design + ".aig; " + script
			native := runNativeABC(t, script)
			got, _ := runScriptStatus(t, script)
			if satlutNormalize(got) != satlutNormalize(native) {
				t.Errorf("%s:\nnative:\n%s\ngo:\n%s", script, native, got)
			}
		}
	}
}
