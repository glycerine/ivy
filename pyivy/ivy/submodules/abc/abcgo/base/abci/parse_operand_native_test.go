package abci

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A formula ending in a binary operator ("a * ") left one operand on the
// stack: Parse_ParserPerformTopOp popped NULL for the other and Hop_And
// (parseEqn.c) or Cudd_bddAnd (parseCore.c) dereferenced it, so reading such
// an EQN file or genlib crashed native (CBUGS_FIXED.md #23).
func TestNativeParserMissingOperandFails(t *testing.T) {
	bin := os.Getenv("ABC_BINARY")
	if bin == "" {
		t.Skip("set ABC_BINARY to a native abc built from ../src")
	}
	bin, _ = filepath.Abs(bin)
	dir := t.TempDir()
	eqn := filepath.Join(dir, "syntax.eqn")
	genlib := filepath.Join(dir, "bad.genlib")
	if err := os.WriteFile(eqn, []byte("INORDER = a b;\nOUTORDER = y;\ny = a * ;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(genlib, []byte("GATE zero 0 O=CONST0;\nGATE one 0 O=CONST1;\nGATE inv 1 O=!a; PIN * INV 1 999 1 0 1 0\nGATE and2 1 O=a*; PIN * NONINV 1 999 1 0 1 0\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct{ script, want string }{
		{"read " + eqn + "; print_stats", "Reading network from file has failed."},
		{"read_genlib " + genlib, "Unknown operation"},
	} {
		out, err := exec.Command(bin, "-c", c.script).CombinedOutput()
		if err != nil {
			t.Errorf("%s: %v\n%s", c.script, err, out)
			continue
		}
		if !strings.Contains(string(out), c.want) {
			t.Errorf("%s: want %q:\n%s", c.script, c.want, out)
		}
	}
}
