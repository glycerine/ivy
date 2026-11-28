package abci

import (
	"github.com/glycerine/abcgo/internal/testoracle"
	"path/filepath"
	"testing"
)

// Extra_ProfileWidth (extraBddCas.c) read the pointer-sized records of its
// st table through (char**)&LevelStart, an int on the stack: on 64-bit hosts
// the write overflowed it and 'cascade' aborted ("stack smashing detected")
// (CBUGS_FIXED.md #51); and Abc_CascadeExperiment (casCore.c) copied the
// design's name into 100-byte file-name buffers, so a long path aborted it. With the fix Go and C produce the same cascade.
func TestCascadeLikeNative(t *testing.T) {
	dir := t.TempDir()
	// a long (absolute) path overflowed casCore.c's 100-byte file names
	long, err := filepath.Abs("test_vectors/native_s27_drw.blif")
	if err != nil {
		t.Fatal(err)
	}
	for _, design := range []string{"test_vectors/native_s27_drw.blif", "test_vectors/native_mul4_gen.aig", long} {
		for _, cmd := range []string{"cascade", "cascade -v"} {
			goFile, nativeFile := filepath.Join(dir, "go.blif"), filepath.Join(dir, "native.blif")
			script := "read " + design + "; strash; " + cmd + "; print_stats; write_blif "
			native := printSymmNormalize(runNativeABC(t, script+nativeFile))
			port := printSymmNormalize(runScript(t, script+goFile))
			if native != port {
				t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
			}
			g, _ := testoracle.ReadFile(goFile)
			n, _ := testoracle.ReadFile(nativeFile)
			if blifStamp.ReplaceAllString(string(g), "") != blifStamp.ReplaceAllString(string(n), "") {
				t.Errorf("%s: written network differs from native", script)
			}
		}
	}
}
