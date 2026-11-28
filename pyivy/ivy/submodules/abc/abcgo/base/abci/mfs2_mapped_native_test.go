package abci

import (
	"os"
	"path/filepath"
	"testing"
)

// mfs2 (Abc_NtkPerformMfs) converts a mapped network to SOPs and inserts
// the new functions with Abc_SopCreateFromTruthIsop on pNtk->pManFunc. The
// port's SOP conversion leaves no SOP manager (SOPs are Go strings), and
// Abc_NtkInsertMfs asserted it as *mem.Mem_Flex_t: mfs2 after if panicked
// ("interface {} is nil, not *mem.Mem_Flex_t") where ABC gives 129 nodes.
func TestMfs2OnMappedNetworkLikeNative(t *testing.T) {
	dir := t.TempDir()
	for _, prep := range []string{"if -K 6", "if -K 6; sop", "if -K 4"} {
		goFile, nativeFile := filepath.Join(dir, "go.blif"), filepath.Join(dir, "native.blif")
		script := "read test_vectors/bench/mul8.aig; strash; " + prep + "; mfs2; print_stats; write_blif "
		native := runNativeABC(t, script+nativeFile)
		out, _ := runScriptStatus(t, script+goFile)
		if printSymmNormalize(native) != printSymmNormalize(out) {
			t.Errorf("%s: output differs\nnative:\n%s\nport:\n%s", prep, native, out)
			continue
		}
		g, _ := os.ReadFile(goFile)
		n, _ := os.ReadFile(nativeFile)
		if blifStamp.ReplaceAllString(string(g), "") != blifStamp.ReplaceAllString(string(n), "") {
			t.Errorf("%s: written BLIF differs from native", prep)
		}
	}
}
