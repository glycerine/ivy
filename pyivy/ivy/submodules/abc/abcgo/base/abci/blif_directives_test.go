package abci

import (
	"strings"
	"testing"
)

// Native: abc -c "read test_vectors/seq/s27.blif; print_stats". ABC skips
// BLIF directives it does not know (here .wire_load_slope), saying so.
func TestBlifUnknownDirectiveIsSkippedLikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s27.blif; print_stats")
	if !strings.Contains(out, `Line 4: Skipping line ".wire_load_slope 0.00".`) {
		t.Errorf("missing skip notice:\n%s", out)
	}
	if !strings.Contains(out, "i/o =    4/    1  lat =    3  nd =    10  edge =     18  cube =    13  lev = 6") {
		t.Errorf("network not read like native:\n%s", out)
	}
}
