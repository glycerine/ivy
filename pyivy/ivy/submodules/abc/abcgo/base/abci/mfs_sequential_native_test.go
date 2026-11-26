package abci

import (
	"strings"
	"testing"
)

const mfsSequentialRefusal = `Abc_CommandAbc9Mfs(): The current AIG is sequential; "&mfs" works only for combinational networks.`

// TestNativeMfsRefusesSequentialAig runs &mfs on mapped s27 through the
// native binary named by ABC_BINARY. Gia_ManPerformMfs asserts a
// combinational AIG; release builds skipped the assertion and dropped all
// registers and nearly all logic (CBUGS_FIXED.md #7). Now the command
// refuses, leaving the design intact.
func TestNativeMfsRefusesSequentialAig(t *testing.T) {
	out := runNativeABC(t, "read test_vectors/seq/s27.blif; strash; &get; &if -K 4; &mfs; &ps")
	if !strings.Contains(out, mfsSequentialRefusal) || !strings.Contains(out, "ff =      3") {
		t.Fatalf("native &mfs on s27:\n%s", out)
	}
}

// The port refuses the same way.
func TestMfsRefusesSequentialAig(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s27.blif; strash; &get; &if -K 4; &mfs; &ps")
	if !strings.Contains(out, mfsSequentialRefusal) || !strings.Contains(out, "ff =      3") {
		t.Fatalf("&mfs on s27:\n%s", out)
	}
}
