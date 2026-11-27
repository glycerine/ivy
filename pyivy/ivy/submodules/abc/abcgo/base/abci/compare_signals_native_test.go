package abci

import (
	"path/filepath"
	"strings"
	"testing"
)

// Abc_NtkCompareSignals reports why two networks' interfaces do not match
// (Abc_NtkComparePis/Pos/Boxes); the port returned 0 silently, so cec only
// said "Miter computation has failed."
func TestCompareSignalsReportsMismatch(t *testing.T) {
	unnamed := filepath.Join(t.TempDir(), "i10dc2.aig")
	runScript(t, "read ../../../i10.aig; dc2; write_aiger "+unnamed)
	script := "cec ../../../i10.aig " + unnamed
	want := `Primary input #0 is different in network 1 ( "V1(0)") and in network 2 ("pi000").` + "\nMiter computation has failed."
	if out := runScript(t, script); !strings.Contains(out, want) {
		t.Errorf("port:\n%s", out)
	}
	cases := map[string]string{
		"cec ../../../i10.aig test_vectors/seq/s713.blif": "Networks have different number of primary inputs.",
	}
	for script, want := range cases {
		if out := runScript(t, script); !strings.Contains(out, want) {
			t.Errorf("%s:\n%s", script, out)
		}
	}
}

func TestCompareSignalsLikeNative(t *testing.T) {
	unnamed := filepath.Join(t.TempDir(), "i10dc2.aig")
	runScript(t, "read ../../../i10.aig; dc2; write_aiger "+unnamed)
	for _, script := range []string{
		"cec ../../../i10.aig " + unnamed,
		"cec ../../../i10.aig test_vectors/seq/s713.blif",
		"dcec ../../../i10.aig " + unnamed,
	} {
		if native, port := bmc3Normalize(runNativeABC(t, script)), bmc3Normalize(runScript(t, script)); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}
