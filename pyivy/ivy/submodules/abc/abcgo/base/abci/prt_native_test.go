package abci

import (
	"regexp"
	"testing"
)

// TestPrtFormatLikeNative: ABC_PRT prints "%s =" then "%9.2f sec" (bmc3's
// "was asserted ... Time =     0.00 sec"); the port printed "%s = %7.2f".
func TestPrtFormatLikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/native_s27_drw.blif; strash; bmc3")
	if !regexp.MustCompile(`was asserted in frame 0\. Time = {5}\d\.\d\d sec\n`).MatchString(out) {
		t.Fatalf("time not in ABC_PRT's format:\n%s", out)
	}
}
