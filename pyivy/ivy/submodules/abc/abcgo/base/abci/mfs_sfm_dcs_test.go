package abci

import (
	"strings"
	"testing"
)

// "&mfs -l" (fUseDcs) can simplify a node to a constant while it keeps its
// fanins. Sfm_TruthToCnf then constrained the first fanin instead of the node
// output, the next window became inconsistent, and native ABC crashed
// (CBUGS_FIXED.md #3).
const mfsDcsScript = "read ../../../i10.aig; &get; &if -K 4; &mfs -a -l; &ps"

// The native result after the fix.
const mfsDcsMapping = "Mapping (K=4)  :  lut =    798  edge =    2681  lev =   17 (6.57)"

func TestNativeMfsDcsConstantNodeDoesNotCrash(t *testing.T) {
	if out := runNativeABC(t, mfsDcsScript); !strings.Contains(out, mfsDcsMapping) {
		t.Fatalf("want %q in native output:\n%s", mfsDcsMapping, out)
	}
}

func TestMfsDcsConstantNodeMatchesNative(t *testing.T) {
	if out := runScript(t, mfsDcsScript); !strings.Contains(out, mfsDcsMapping) {
		t.Fatalf("want %q in output:\n%s", mfsDcsMapping, out)
	}
}
