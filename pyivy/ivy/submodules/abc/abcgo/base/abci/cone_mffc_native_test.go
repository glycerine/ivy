package abci

import (
	"os"
	"path/filepath"
	"testing"
)

// cone builds its network with abcNtk.c's Abc_NtkCreateCone,
// Abc_NtkCreateConeArray (-R), Abc_NtkCreateMffc (-m) or Abc_NtkMakeOnePo
// (-s, which prints a "Run sequential cleanup" hint); the port used one
// generic copier for all of them, which numbered logic nodes differently
// and printed no hint. Also: cone -m (Abc_NtkCreateMffc) names each new PI after its leaf in the
// original network (Abc_ObjName: "n1304" for an unnamed node) and copies the
// MFFC in Abc_NodeMffcConeSupp order. The port built the cone with a generic
// copier and named the PIs by their new IDs ("n7"), so the written network
// differed from ABC's.
func TestConeMffcLikeNative(t *testing.T) {
	dir := t.TempDir()
	for _, prep := range []string{"logic; cone -O 5 -m", "strash; cone -O 5 -m", "logic; cone -O 0 -m", "strash; cone -O 20 -m",
		// Abc_NtkCreateCone / Abc_NtkCreateConeArray / Abc_NtkMakeOnePo
		"logic; cone -O 0", "logic; cone -O 7", "logic; cone -O 20 -a", "strash; cone -O 9 -a",
		"logic; cone -O 2 -R 5", "strash; cone -O 2 -R 5", "logic; cone -O 30 -s", "strash; cone -O 3 -s",
		"logic; cone -O 4 -a -s"} {
		goFile, nativeFile := filepath.Join(dir, "go.blif"), filepath.Join(dir, "native.blif")
		script := "read test_vectors/i10_native_demo.aig; " + prep + "; print_stats; write_blif "
		native := runNativeABC(t, script+nativeFile)
		out, _ := runScriptStatus(t, script+goFile)
		if printSymmNormalize(native) != printSymmNormalize(out) {
			t.Errorf("%s: output differs\nnative:\n%s\nport:\n%s", prep, native, out)
		}
		g, _ := os.ReadFile(goFile)
		n, _ := os.ReadFile(nativeFile)
		if blifStamp.ReplaceAllString(string(g), "") != blifStamp.ReplaceAllString(string(n), "") {
			t.Errorf("%s: written BLIF differs from native:\nnative:\n%s\nport:\n%s", prep, n, g)
		}
	}
}
