package abci

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestNativePrintTruthStaysInBounds compiles src/aig/gia/giaTruth.c and
// src/base/abci/abc.c with AddressSanitizer, links them ahead of the native
// library named by ABC_LIBABC, and runs &print_truth on s27 (4 PIs, 7 CIs).
// The command checked the PI count but computed and printed truth tables
// over all CIs, reading past s_Truths6 (CBUGS_FIXED.md #5).
func TestNativePrintTruthStaysInBounds(t *testing.T) {
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
	bin := filepath.Join(dir, "abc_asan")
	asan := []string{"-g", "-O1", "-fsanitize=address", "-fno-omit-frame-pointer", "-DABC_USE_STDINT_H", "-I", src}
	for _, c := range [][]string{
		append([]string{"gcc"}, append(asan, "-c", main, "-o", filepath.Join(dir, "main.o"))...),
		append([]string{"gcc"}, append(asan, "-c", filepath.Join(src, "aig/gia/giaTruth.c"), "-o", filepath.Join(dir, "giaTruth.o"))...),
		append([]string{"gcc"}, append(asan, "-c", filepath.Join(src, "base/abci/abc.c"), "-o", filepath.Join(dir, "abc.o"))...),
		{"g++", "-fsanitize=address", filepath.Join(dir, "main.o"), filepath.Join(dir, "giaTruth.o"), filepath.Join(dir, "abc.o"), lib, "-lreadline", "-lpthread", "-lm", "-ldl", "-o", bin},
	} {
		if out, err := exec.Command(c[0], c[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c[:3], err, out)
		}
	}
	bench, _ := filepath.Abs("test_vectors/seq/s27.blif")
	out, _ := exec.Command(bin, "read "+bench+"; strash; &get; &print_truth").CombinedOutput()
	if strings.Contains(string(out), "AddressSanitizer") || !strings.Contains(string(out), "The number of inputs of the AIG exceeds 6.") {
		t.Fatalf("&print_truth on s27:\n%s", out)
	}
}
