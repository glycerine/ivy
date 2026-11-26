package abci

import (
	"strings"
	"testing"
)

const anormNoBox = "Cannot find arithmetic boxes."

// TestNativeAnormWithoutAdders: Acec_Normalize passed the NULL box that
// Acec_DeriveBox returns for an AIG without adder trees to Acec_InsertBox,
// crashing native ABC (CBUGS_FIXED.md #11). It now reports it, as
// Acec_ManDecla does, and keeps the AIG.
func TestNativeAnormWithoutAdders(t *testing.T) {
	out := runNativeABC(t, "read test_vectors/seq/s27.blif; strash; &get; &anorm; &ps")
	if !strings.Contains(out, anormNoBox) || !strings.Contains(out, "and =       8") {
		t.Fatalf("native &anorm on s27:\n%s", out)
	}
}

func TestAnormWithoutAdders(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s27.blif; strash; &get; &anorm; &ps")
	if !strings.Contains(out, anormNoBox) || !strings.Contains(out, "and =       8") {
		t.Fatalf("&anorm on s27:\n%s", out)
	}
}
