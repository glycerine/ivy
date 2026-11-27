package abci

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

// Gia_AigerRead (giaAiger.c) reads only '@'-tagged user IDs and break names
// from the symbol table; ordinary names are dropped, so '&r X; &w Y' writes
// no symbols and '&put' gives default names. Abc_NtkFromAigPhase creates its
// latches unnamed, so '&put' names them li00/lo00 (Abc_NtkAddDummyBoxNames).
// The port kept the names, and named latches n<id>_li/_lo.
func TestGiaReadDropsSymbolNamesLikeNative(t *testing.T) {
	dir := t.TempDir()
	stamp := regexp.MustCompile(`(?m)^This file was produced by .*$|^# Benchmark .*$`)
	read := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return stamp.ReplaceAllString(string(data), "")
	}
	named := filepath.Join(dir, "named.aig")
	bare := filepath.Join(dir, "bare.aig")
	runNativeABC(t, "read test_vectors/seq/s713.blif; strash; write_aiger -s "+named+"; write_aiger "+bare)
	for _, input := range []string{"test_vectors/i10_native_demo.aig", named, bare} {
		for _, step := range []struct{ write, ext string }{
			{"&w", "aig"},
			{"&put; write_blif", "blif"},
		} {
			goFile, nativeFile := filepath.Join(dir, "go."+step.ext), filepath.Join(dir, "native."+step.ext)
			script := "&r " + input + "; " + step.write + " "
			runScript(t, script+goFile)
			runNativeABC(t, script+nativeFile)
			if read(goFile) != read(nativeFile) {
				t.Errorf("%s: output differs from native", script)
			}
		}
	}
}
