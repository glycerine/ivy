package abci

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestNativeSuperclAssertionsHold compiles src/base/abci/abcLut.c with
// assertions enabled, links it ahead of the native library named by
// ABC_LIBABC, and runs supercl. Abc_NodeLutMap asserted that a node's level
// is 0 although the driver no longer clears it (see CBUGS_FIXED.md #4).
func TestNativeSuperclAssertionsHold(t *testing.T) {
	lib := os.Getenv("ABC_LIBABC")
	if lib == "" {
		t.Skip("set ABC_LIBABC to a native libabc.a built from ../src")
	}
	lib, _ = filepath.Abs(lib)
	src, _ := filepath.Abs("../../../src")
	dir := t.TempDir()
	main := filepath.Join(dir, "main.c")
	if err := os.WriteFile(main, []byte(`#include "base/main/main.h"
int main(int argc, char **argv) { Abc_Start(); return Cmd_CommandExecute(Abc_FrameGetGlobalFrame(), argv[1]); }
`), 0o644); err != nil {
		t.Fatal(err)
	}
	bin := filepath.Join(dir, "abc_assert")
	for _, c := range [][]string{
		{"gcc", "-O1", "-DABC_USE_STDINT_H", "-I", src, "-c", main, "-o", filepath.Join(dir, "main.o")},
		{"gcc", "-O1", "-DABC_USE_STDINT_H", "-UNDEBUG", "-I", src, "-c", filepath.Join(src, "base/abci/abcLut.c"), "-o", filepath.Join(dir, "abcLut.o")},
		{"g++", filepath.Join(dir, "main.o"), filepath.Join(dir, "abcLut.o"), lib, "-lreadline", "-lpthread", "-lm", "-ldl", "-o", bin},
	} {
		if out, err := exec.Command(c[0], c[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, out)
		}
	}
	bench, _ := filepath.Abs("test_vectors/bench/mul8.aig")
	out, err := exec.Command(bin, "read "+bench+"; strash; supercl; print_stats").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "nd =   483") {
		t.Fatalf("supercl with assertions: %v\n%s", err, out)
	}
}
