package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Networks strashed inside Abc_NtkPrepareTwoNtks (miter, cec, dsec with
// file arguments) must be built as the strash command builds them: ABC's
// two halves of 'read X; strash; miter -c X' hash together to no nodes.
// base/abc's own Abc_NtkStrash decomposed each node's function its own way,
// so the halves differed (76 AND nodes on s713).
func TestPrepareTwoNtksStrashesLikeStrashCommand(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s713.blif; strash; miter -c test_vectors/seq/s713.blif; print_stats")
	if !strings.Contains(out, "and =      0") {
		t.Fatalf("miter of a network with itself is not empty:\n%s", out)
	}
}

// The sequential miter of a network with itself is built node for node as
// ABC builds it: the strash of the second network inside
// Abc_NtkPrepareTwoNtks, and Abc_AigXor's two products in the order the GCC
// build of abcAig.c creates them (right to left).
func TestMiterStructureLikeNative(t *testing.T) {
	dir := t.TempDir()
	for _, design := range []string{"test_vectors/native_s27_drw.blif", "test_vectors/seq/s713.blif", "test_vectors/seq/s1494.blif"} {
		for _, script := range []string{
			"miter " + design + " " + design,
			"read " + design + "; strash; miter -c " + design,
		} {
			goFile, nativeFile := filepath.Join(dir, "go.bench"), filepath.Join(dir, "native.bench")
			runScript(t, script+"; write_bench "+goFile)
			runNativeABC(t, script+"; write_bench "+nativeFile)
			got, err := os.ReadFile(goFile)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(nativeFile)
			if err != nil {
				t.Fatal(err)
			}
			// the header line carries the time of writing
			if blifStamp.ReplaceAllString(string(got), "") != blifStamp.ReplaceAllString(string(want), "") {
				t.Errorf("%s: miter differs from native", script)
			}
		}
	}
}
