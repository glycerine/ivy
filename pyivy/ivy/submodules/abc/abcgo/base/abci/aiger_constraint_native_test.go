package abci

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const constraintAig = "../io/test_vectors/constraint1.aig" // aig 4 2 1 1 1 0 1: one PO, one constraint

func aigerPrefix(t *testing.T, path string, n int) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data[:min(n, len(data))]
}

// TestAigerConstraintsReadInvertedLikeNative: Io_ReadAiger inverts the
// constraint outputs (Abc_NtkInvertConstraints) when the file has B/C/J/F
// sections, so ABC's constraint PO is the complement of the file literal.
func TestAigerConstraintsReadInvertedLikeNative(t *testing.T) {
	path := filepath.Join(t.TempDir(), "con.blif")
	runScript(t, "read "+constraintAig+"; write_blif "+path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), ".names pi0 po1\n1 1\n") {
		t.Fatalf("constraint PO not inverted on read:\n%s", data)
	}
}

// TestAigerConstraintsWriteLikeNative: Io_WriteAiger with constraints writes
// an extended header with no outputs, the POs as bad states, and the
// constraints inverted back; write_aiger -u writes the canonical AIG (no
// constraint count) with ABC's internal polarity.
func TestAigerConstraintsWriteLikeNative(t *testing.T) {
	dir := t.TempDir()
	plain, unique := filepath.Join(dir, "plain.aig"), filepath.Join(dir, "u.aig")
	runScript(t, "read "+constraintAig+"; write_aiger "+plain+"; write_aiger -u "+unique)
	if want := "aig 4 2 1 0 1 1 1\n8\n6\n3\n\x04\x02"; !bytes.Equal(aigerPrefix(t, plain, len(want)), []byte(want)) {
		t.Errorf("plain write = %q", aigerPrefix(t, plain, 40))
	}
	if want := "aig 4 2 1 2 1\n8\n6\n2\n\x04\x02"; !bytes.Equal(aigerPrefix(t, unique, len(want)), []byte(want)) {
		t.Errorf("canonical write = %q", aigerPrefix(t, unique, 40))
	}
}

// TestAigerLatchNamesLikeNative: Io_ReadAiger names a latch's output by its
// symbol, the latch box name+"L" and the latch input name+"_in".
func TestAigerLatchNamesLikeNative(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ext.blif")
	runScript(t, "read ../io/test_vectors/aiger_binary_extended.aig; write_blif "+path)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), ".latch   state_in      state  0\n") || !strings.Contains(string(data), ".names state_in out\n") {
		t.Fatalf("latch input not named state_in:\n%s", data)
	}
}

// TestNativeAigerSectionSymbolsLikePort: AIGER numbers the b, c, j and f
// symbols within their own sections. Io_ReadAiger indexed them into the PO
// list from 0, so "b0" landed on the first output (already named "out") and
// the bad and constraint POs kept default names (CBUGS_FIXED.md). Native
// and the port must both name them.
func TestNativeAigerSectionSymbolsLikePort(t *testing.T) {
	for _, run := range []func(*testing.T, string) string{runScript, runNativeABC} {
		path := filepath.Join(t.TempDir(), "ext.blif")
		run(t, "read ../io/test_vectors/aiger_binary_extended.aig; write_blif "+path)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), ".outputs out bad constraint\n") {
			t.Fatalf("bad/constraint symbols not applied:\n%s", data)
		}
	}
}
