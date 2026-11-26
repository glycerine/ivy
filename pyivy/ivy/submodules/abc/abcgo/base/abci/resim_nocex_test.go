package abci

import (
	"strings"
	"testing"
)

const resimNoCex = "Cec_ManSeqResimulateCounter(): Counter-example is not available."

// TestNativeResimWithoutCexDoesNotCrash: &resim reported the missing
// counter-example and then read pAbc->pCex->iFrame, crashing native ABC
// (CBUGS_FIXED.md #8).
func TestNativeResimWithoutCexDoesNotCrash(t *testing.T) {
	out := runNativeABC(t, "read test_vectors/seq/s27.blif; strash; &get; &resim; &ps")
	if !strings.Contains(out, resimNoCex) || !strings.Contains(out, "ff =      3") {
		t.Fatalf("native &resim without a counter-example:\n%s", out)
	}
}

func TestResimWithoutCexDoesNotPanic(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s27.blif; strash; &get; &resim; &ps")
	if !strings.Contains(out, resimNoCex) || !strings.Contains(out, "ff =      3") {
		t.Fatalf("&resim without a counter-example:\n%s", out)
	}
}
