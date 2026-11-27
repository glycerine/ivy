package abci

import (
	"strings"
	"testing"
)

// &if maps through Gia_ManToIf, which creates one mapper AND per GIA object
// (the IDs must stay aligned), so XOR and MUX objects of a '&st -m' graph
// were mapped as ANDs and ABC's result was not equivalent to the input
// (C bug 29). &if now refuses such graphs, as &mfsd does (C bug 28); the
// port must print the same thing and never return a wrong mapping.
func TestIfRejectsMuxGraphsLikeNative(t *testing.T) {
	for _, design := range []string{"test_vectors/native_mul4_gen.aig", "test_vectors/native_mul8_canon.aig"} {
		for _, flow := range []string{"&st -m; &if -K 4", "&st -m; &if -K 6", "&st -L 1 -m; &if -K 4"} {
			script := "&r " + design + "; " + flow + "; &put; cec -n " + design
			native := runNativeABC(t, script)
			got, _ := runScriptStatus(t, script)
			if strings.Contains(native, "NOT EQUIVALENT") || strings.Contains(got, "NOT EQUIVALENT") {
				t.Errorf("%s: mapping changed the function\nnative:\n%s\ngo:\n%s", script, native, got)
			}
			if bmc3Normalize(got) != bmc3Normalize(native) {
				t.Errorf("%s:\nnative:\n%s\ngo:\n%s", script, native, got)
			}
		}
	}
}
