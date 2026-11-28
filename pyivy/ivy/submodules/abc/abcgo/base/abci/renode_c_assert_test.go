package abci

import (
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNativeRenodeIfAssertionsHold compiles src/map/if/ifMap.c with
// assertions enabled, links it ahead of the native library named by
// ABC_LIBABC, and runs renode on i10 after if -K 6; lutpack; strash. Every
// cut of node 709 violated its required time after the filter had dropped
// the saved best cut, so If_ObjPerformMappingAnd asserted nCuts > 0; release
// builds went on and adopted a stale cut slot (CBUGS_FIXED.md #15).
func TestNativeRenodeIfAssertionsHold(t *testing.T) {
	lib := testoracle.NativeLibABCPath()
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
		{"gcc", "-O1", "-DABC_USE_STDINT_H", "-UNDEBUG", "-I", src, "-c", filepath.Join(src, "map/if/ifMap.c"), "-o", filepath.Join(dir, "ifMap.o")},
		{"g++", filepath.Join(dir, "main.o"), filepath.Join(dir, "ifMap.o"), lib, "-lreadline", "-lpthread", "-lm", "-ldl", "-o", bin},
	} {
		if out, err := testoracle.Command(t, c[0], c[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c, err, out)
		}
	}
	aig, _ := filepath.Abs("test_vectors/i10_lutpack_strash.aig")
	out, err := testoracle.Command(t, bin, "read "+aig+"; renode; print_stats").CombinedOutput()
	if err != nil || !strings.Contains(string(out), "nd =") {
		t.Fatalf("renode with assertions: %v\n%s", err, out)
	}
}

// TestRenodeKeepsBestCutWhenAllCutsViolateRequired: the Go port asserted
// (panicked) where C's release build adopted a stale cut; both now keep the
// previous best cut and agree on the result.
func TestRenodeKeepsBestCutWhenAllCutsViolateRequired(t *testing.T) {
	out := runScript(t, "read test_vectors/i10_lutpack_strash.aig; renode; print_stats")
	if !strings.Contains(out, "nd =  2350  edge =   4856  aig  =  2651  lev = 31") {
		t.Fatalf("renode result differs from fixed native:\n%s", out)
	}
}

// TestRenodeBddCompletes: renode -b scores cuts with Extra_Reorder, which
// fills pOrder only for the support's variables (the cut can be smaller
// than the manager); the port copied a whole-manager permutation into the
// cut's order and panicked. (reo_native_test.go compares the results
// with native ABC.)
func TestRenodeBddCompletes(t *testing.T) {
	out := runScript(t, "read test_vectors/native_s27_drw.blif; strash; renode -b; print_stats")
	if !strings.Contains(out, "nd =") {
		t.Fatalf("renode -b failed:\n%s", out)
	}
}
