package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Three native crashes on a combinational AIG read by absolute path
// (CBUGS_FIXED.md #53): &splitprove read a CEX from a freed manager
// (cecSplit.c), l2ssim ANDed a NULL register-equality accumulator when there
// are no registers (liveness_sim.c), and &mlgen built its output file names
// in 100-byte buffers (giaGen.c). Go and native now print the same for
// &splitprove and &mlgen.
func TestMul4CrashesLikeNative(t *testing.T) {
	design, err := filepath.Abs("test_vectors/native_mul4_gen.aig")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	// &mlgen writes next to the design; a path over 100 bytes overflowed
	long := filepath.Join(dir, strings.Repeat("d", 80))
	if err := os.MkdirAll(long, 0o755); err != nil {
		t.Fatal(err)
	}
	copyPath := filepath.Join(long, "mul4.aig")
	data, err := os.ReadFile(design)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(copyPath, data, 0o644); err != nil {
		t.Fatal(err)
	}
	// l2ssim's Go port (proof/live) is not yet a port of liveness_sim.c; only
	// the native crash is checked for it
	if out := runNativeABC(t, "read "+copyPath+"; strash; &get; l2ssim"); !strings.Contains(out, "liveness property") {
		t.Errorf("l2ssim: unexpected native output:\n%s", out)
	}
	for _, cmd := range []string{"&splitprove", "&mlgen"} {
		script := "read " + copyPath + "; strash; &get; " + cmd
		native := printSymmNormalize(runNativeABC(t, script))
		out, _ := runScriptStatus(t, script)
		if port := printSymmNormalize(out); native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}
