package abci

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestIsoReduceLikeNative: iso keeps one PO per class of outputs whose
// cones have the same canonical form (Saig_ManFindIsoPerm's refinement,
// Saig_ManDupIsoCanonical, the AIGER bytes compared). The port's colour
// refinement found no classes on i10.
func TestIsoReduceLikeNative(t *testing.T) {
	out := runScript(t, "read ../../../i10.aig; strash; iso; print_stats")
	for _, want := range []string{"Reduced 224 outputs to 149 outputs.", "i/o =  252/  149"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in\n%s", want, out)
		}
	}
}

// TestWriteAigerCanonicalLikeNative: write_aiger -u writes
// Saig_ManDupIsoCanonical's AIG; mul8's must equal native's byte for byte
// (up to the dated comment section).
func TestWriteAigerCanonicalLikeNative(t *testing.T) {
	path := filepath.Join(t.TempDir(), "canon.aig")
	runScript(t, "read test_vectors/native_mul8_if6.blif; strash; write_aiger -u "+path)
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile("test_vectors/native_mul8_canon.aig")
	if err != nil {
		t.Fatal(err)
	}
	cut := func(b []byte) []byte {
		if i := bytes.Index(b, []byte("\nc\n")); i >= 0 {
			return b[:i]
		}
		return b
	}
	if !bytes.Equal(cut(got), cut(want)) {
		t.Fatalf("canonical AIG differs from native (%d vs %d bytes)", len(cut(got)), len(cut(want)))
	}
}
