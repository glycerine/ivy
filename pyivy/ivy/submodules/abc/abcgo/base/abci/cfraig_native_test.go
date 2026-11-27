package abci

import (
	"os"
	"path/filepath"
	"testing"
)

// &cfraig (sscCore.c) on a sequential AIG: Ssc_GiaResetPiPattern sizes the
// pattern array by CI count but Gia_ObjSimWords (gia.h) divided it by the
// PI count, so Gia_ObjSimPi pointed past the array for latch outputs and
// Ssc_GiaTransferPiPattern overflowed the heap (CBUGS_FIXED.md #46).
func TestCFraigLikeNative(t *testing.T) {
	dir := t.TempDir()
	for _, design := range []string{
		"test_vectors/native_s27_drw.blif",
		"test_vectors/seq/s713.blif",
		"test_vectors/i10_native_demo.aig",
	} {
		for _, cmd := range []string{"&cfraig"} {
			goFile, nativeFile := filepath.Join(dir, "go.aig"), filepath.Join(dir, "native.aig")
			script := "read " + design + "; strash; &get; " + cmd + "; &ps; &w "
			native := printSymmNormalize(runNativeABC(t, script+nativeFile))
			port := printSymmNormalize(runScript(t, script+goFile))
			if native != port {
				t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
			}
			g, _ := os.ReadFile(goFile)
			n, _ := os.ReadFile(nativeFile)
			if string(g) != string(n) {
				t.Errorf("%s: written AIG differs from native", script)
			}
		}
	}
}
