package abci

import (
	"os"
	"path/filepath"
	"testing"
)

// Io_WriteAiger (ioWriteAiger.c) writes the extended header when a latch is
// not initialized to 0: "aig M I L 0 A B C" with both B (POs) and C
// (constraints), C written even when 0. The port dropped the trailing zero
// count ("aig 6 1 2 0 3 1").
func TestWriteAigerInit1LatchLikeNative(t *testing.T) {
	dir := t.TempDir()
	for _, prep := range []string{"strash", "strash; dc2"} {
		goFile, nativeFile := filepath.Join(dir, "go.aig"), filepath.Join(dir, "native.aig")
		script := "read test_vectors/seq/latch_init1.blif; " + prep + "; write_aiger "
		runNativeABC(t, script+nativeFile)
		runScript(t, script+goFile)
		g, _ := os.ReadFile(goFile)
		n, _ := os.ReadFile(nativeFile)
		if aigerStamp.ReplaceAllString(string(g), "") != aigerStamp.ReplaceAllString(string(n), "") {
			t.Errorf("%s: written AIG differs from native:\nnative %q\nport   %q", prep, firstLine(n), firstLine(g))
		}
	}
}

func firstLine(b []byte) string {
	for i, c := range b {
		if c == '\n' {
			return string(b[:i])
		}
	}
	return string(b)
}
