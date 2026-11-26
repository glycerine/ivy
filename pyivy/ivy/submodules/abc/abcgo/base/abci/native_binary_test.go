package abci

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runNativeABC runs a script through the native abc executable named by
// ABC_BINARY (built from this checkout's src/), skipping when it is unset.
func runNativeABC(t *testing.T, script string) string {
	t.Helper()
	bin := os.Getenv("ABC_BINARY")
	if bin == "" {
		t.Skip("set ABC_BINARY to a native abc built from ../src")
	}
	bin, err := filepath.Abs(bin)
	if err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command(bin, "-c", script).CombinedOutput()
	if err != nil {
		t.Fatalf("native abc failed on %q: %v\n%s", script, err, out)
	}
	return ansiEscape.ReplaceAllString(string(out), "")
}

// &synch2 stores its choices as GIA siblings, which &put's choice
// conversion (Gia_ManToAig with choices) cannot read; native ABC crashed
// here (CBUGS_FIXED.md #2).
const synch2PutScript = "read ../../../i10.aig; &get; &st; &synch2; &put; print_stats"

func TestNativeSynch2PutDoesNotCrash(t *testing.T) {
	if out := runNativeABC(t, synch2PutScript); !strings.Contains(out, "and =") {
		t.Fatalf("no statistics after &put:\n%s", out)
	}
}

func TestSynch2PutDoesNotPanic(t *testing.T) {
	expectStats(t, synch2PutScript, "and =   1861  lev = 28") // native result
}
