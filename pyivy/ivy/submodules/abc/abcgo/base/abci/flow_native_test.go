package abci

import (
	"strings"
	"testing"
)

// TestBalanceLutLikeNative: &blut pairs supergate inputs by affinity
// counts (Str_CountBits). ABC's count was two short for words with three
// or more bits (CBUGS_FIXED.md); with that fixed, native and the port agree.
func TestBalanceLutLikeNative(t *testing.T) {
	out := runScript(t, "read test_vectors/native_mul8_if6.blif; strash; &get; &blut -K 6; &ps")
	if !strings.Contains(out, "and =     925  lev =   58 (34.12)") {
		t.Fatalf("unexpected &blut result:\n%s", out)
	}
}

// TestFlowLikeNative: the &flow script (&dsdb, &dch, &if, &lf, &syn2,
// &blut, &sopb rounds) on i10 maps as native ABC does.
func TestFlowLikeNative(t *testing.T) {
	out := runScript(t, "read ../../test_vectors/i10.aig; &get; &flow; &ps")
	if !strings.Contains(out, "lut =    561  edge =    2219  lev =    7 (4.10)") {
		t.Fatalf("unexpected &flow result:\n%s", out)
	}
}
