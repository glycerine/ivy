package abci

import (
	"io"
	"strings"
	"testing"

	base_main "github.com/glycerine/abcgo/base/main"
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

// Gia_ManDeepSyn reports a failed script with Abc_Print( 1, ...), on stdout
// with plain quotes; the port wrote it to the error stream with %q.
func TestDeepSynFailureReportedOnStdout(t *testing.T) {
	out, _ := runScriptStatusWith(t, "read test_vectors/bench/mul8.aig; strash; &get; &deepsyn -S 1; &ps",
		func(f *base_main.Frame) { f.Err = io.Discard })
	if !strings.Contains(out, `Something did not work out with the command "&dch; &if -a -K 3; &mfs -e -W 20 -L 20; &put; compress2rs; &get".`) {
		t.Fatalf("stdout:\n%s", out)
	}
}
