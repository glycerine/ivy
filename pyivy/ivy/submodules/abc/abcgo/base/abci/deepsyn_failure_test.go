package abci

import (
	"strings"
	"testing"
)

// TestNativeDeepSynSurvivesFailedCommand: without abc.rc, compress2rs is
// unknown, so Gia_ManDeepSynOne returns NULL and Gia_ManDeepSyn took its
// AND count, crashing native ABC (CBUGS_FIXED.md #13). Now the search stops
// and the best AIG so far (the input) is kept.
func TestNativeDeepSynSurvivesFailedCommand(t *testing.T) {
	out := runNativeABC(t, "read test_vectors/bench/mul8.aig; strash; &get; &deepsyn -S 1; &ps")
	if !strings.Contains(out, "Something did not work out") || !strings.Contains(out, "and =     424") {
		t.Fatalf("native &deepsyn:\n%s", out)
	}
}

func TestDeepSynSurvivesFailedCommand(t *testing.T) {
	out := runScript(t, "read test_vectors/bench/mul8.aig; strash; &get; &deepsyn -S 1; &ps")
	if !strings.Contains(out, "Something did not work out") || !strings.Contains(out, "and =     424") ||
		strings.Contains(out, "Tranformation has failed") {
		t.Fatalf("&deepsyn:\n%s", out)
	}
}
