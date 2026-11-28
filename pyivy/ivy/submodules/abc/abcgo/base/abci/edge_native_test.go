package abci

import (
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// '&edge' runs Seg_ManComputeDelay (giaSatEdge.c) by default, '-o' runs
// Gia_ManComputeEdgeDelay and '-o -r' Gia_ManComputeEdgeDelay2 (giaEdge.c),
// '-p' Gia_ManConvertPackingToEdges and '-m' Sle_ManExplore (giaSatLE.c);
// '&ps' evaluates the edges with Gia_ManEvalEdgeDelay. The SAT search was a
// reimplementation that passed per-call assumptions to bsat's Solve instead
// of C's sat_solver_push/sat_solver_solve_internal with the random and
// polarity settings, so it found other optimal assignments (i10 K=4: 174
// edges instead of 181). Gia_ObjEvalEdgeDelay asserts that a LUT has at
// most four fanins, but release builds compile that out and the loop
// handles any count; the port panicked on '&if -K 5/6; &edge; &ps'.
var edgeDesigns = []string{"i10_native_demo", "native_mul8_canon", "native_mul4_gen"}

var edgeOptions = []string{"", "-E 2", "-F 3", "-E 2 -F 2", "-D 9", "-C 10", "-o", "-o -E 2", "-o -r", "-o -r -E 2", "-v", "-E 2 -v", "-o -v", "-d"}

var edgeMB = regexp.MustCompile(`[0-9.]+ MB`)

func edgeCompare(t *testing.T, script string) {
	t.Helper()
	native := runNativeABC(t, script)
	got, _ := runScriptStatus(t, script)
	if edgeMB.ReplaceAllString(bmc3Normalize(got), "MB") != edgeMB.ReplaceAllString(bmc3Normalize(native), "MB") {
		t.Errorf("%s:\nnative:\n%s\ngo:\n%s", script, native, got)
	}
}

func TestEdgeOnSixInputLutsLikeNative(t *testing.T) {
	for _, script := range []string{
		"&r test_vectors/i10_native_demo.aig; &if -K 6; &edge; &ps",
		"&r test_vectors/i10_native_demo.aig; &if -K 5; &edge -E 2; &ps",
		"&r test_vectors/native_mul8_canon.aig; &if -K 6; &edge; &ps",
	} {
		edgeCompare(t, script)
	}
}

func TestEdgeOptionsLikeNative(t *testing.T) {
	for _, design := range edgeDesigns {
		for _, k := range []string{"3", "4", "5", "6"} {
			for _, opt := range edgeOptions {
				edgeCompare(t, "&r test_vectors/"+design+".aig; &if -K "+k+"; &edge "+opt+"; &ps")
			}
		}
	}
}

// Commands that read, write or reuse the edges.
func TestEdgeConsumersLikeNative(t *testing.T) {
	for _, design := range edgeDesigns {
		r := "&r test_vectors/" + design + ".aig; "
		for _, script := range []string{
			"&if -K 4; &edge; &edge -o; &ps",
			"&if -K 4; &edge -o; &edge; &ps",
			"&if -K 4; &pack; &ps; &edge -p; &ps",
			"&if -K 4; &pack -N 3; &edge -p; &ps",
			"&if -K 4; &pack -N 4; &edge -p; &ps",
			"&if -K 4; &edge -p",
			"&if -K 4; &edge; &satlut -d; &ps",
			"&if -K 4; &edge -E 2; &satlut -d -Q 2; &ps",
			"&if -K 6; &edge; &satlut -d; &ps",
			"&lf -K 4; &edge; &ps",
			"&jf -K 4; &edge -o; &ps",
		} {
			edgeCompare(t, r+script)
		}
	}
}

// The AIGER comment written by '&w' ends with the time of day.
var aigerStamp = regexp.MustCompile(`on [A-Z][a-z]{2} [A-Z][a-z]{2} [ 0-9]\d \d\d:\d\d:\d\d \d{4}`)

// '&edge -m' (Sle_ManExplore, giaSatLE.c) solved per-call assumptions with
// bsat's Solve instead of sat_solver_push/sat_solver_solve_internal with
// C's random, polarity and resource settings (mul4 K=4: other edges). Runs
// on native_mul8_canon take minutes, so it is checked on the smaller designs.
func TestEdgeMappingLikeNative(t *testing.T) {
	for _, script := range []string{
		"&r test_vectors/native_mul4_gen.aig; &if -K 4; &edge -m -v; &ps",
		"&r test_vectors/native_mul4_gen.aig; &edge -m -v; &ps",
		"&r test_vectors/native_mul4_gen.aig; &if -K 3; &edge -m -E 2 -v; &ps",
		"&r test_vectors/native_mul4_gen.aig; &if -K 4; &edge; &edge -m -d -v; &ps",
		"&r test_vectors/native_mul4_gen.aig; &if -K 4; &edge -m -C 100 -v; &ps",
		"&r test_vectors/native_mul4_gen.aig; &if -K 6; &edge -m; &ps",
		"&r test_vectors/i10_native_demo.aig; &if -K 4; &edge -m -v; &ps",
	} {
		edgeCompare(t, script)
	}
}

// '&w' stores the edges in the AIGER "w" extension.
func TestEdgeWriteReadLikeNative(t *testing.T) {
	dir := t.TempDir()
	for _, design := range edgeDesigns {
		for _, k := range []string{"4", "6"} {
			cFile := filepath.Join(dir, design+k+"_c.aig")
			goFile := filepath.Join(dir, design+k+"_go.aig")
			script := "&r test_vectors/" + design + ".aig; &if -K " + k + "; &edge; &w "
			runNativeABC(t, script+cFile)
			runScriptStatus(t, script+goFile)
			c, err1 := testoracle.ReadFile(cFile)
			g, err2 := testoracle.ReadFile(goFile)
			if err1 != nil || err2 != nil || aigerStamp.ReplaceAllString(string(c), "") != aigerStamp.ReplaceAllString(string(g), "") {
				t.Errorf("%s: written files differ (%v, %v)", script, err1, err2)
			}
			// Without -s the AIG is rehashed and the "w" extension is skipped.
			edgeCompare(t, "&r "+cFile+"; &ps; &edge -o; &ps")
			edgeCompare(t, "&r -s "+cFile+"; &ps; &edge -o; &ps; &w "+goFile+"; &r -s "+goFile+"; &ps")
		}
	}
}

// '&edge -D' below the smallest feasible delay: no assignment is found, and
// Seg_ManComputeDelay passed the NULL result to Gia_ManEdgeFromArray, so
// native ABC segfaulted (CBUGS_FIXED.md #31). '&pack -N 5' builds blocks of
// five LUTs, which Gia_ManConvertPackingToEdges copied into int nNodes[4]
// (stack overflow; '-N 8' was "stack smashing detected", CBUGS_FIXED.md #30).
var edgeCrashScripts = []string{
	"&if -K 4; &edge -D 5; &ps",
	"&if -K 4; &edge -D 5 -v; &ps",
	"&if -K 4; &edge -o; &edge -D 3; &ps",
	"&if -K 4; &pack -N 5; &edge -p; &ps",
	"&if -K 4; &pack -N 8; &edge -p; &ps",
	"&if -K 6; &pack -N 6; &edge -p; &ps",
}

func TestEdgeCrashesFixedLikeNative(t *testing.T) {
	for _, design := range edgeDesigns {
		for _, script := range edgeCrashScripts {
			edgeCompare(t, "&r test_vectors/"+design+".aig; "+script)
		}
	}
}

// Compiles giaEdge.c and giaSatEdge.c with AddressSanitizer against the
// native libabc.a named by ABC_LIBABC. Record it against a release library
// (assertions off, like the native binary): ABC_LIBABC=.scratch/abc-release/libabc.a.
func TestNativeEdgeStaysInBounds(t *testing.T) {
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
	bin := filepath.Join(dir, "abc_asan")
	asan := []string{"-g", "-O1", "-fsanitize=address", "-fno-omit-frame-pointer", "-DABC_USE_STDINT_H", "-DNDEBUG", "-I", src}
	objs := []string{filepath.Join(dir, "main.o")}
	cmds := [][]string{append([]string{"gcc"}, append(asan, "-c", main, "-o", objs[0])...)}
	for _, f := range []string{"giaEdge", "giaSatEdge"} {
		obj := filepath.Join(dir, f+".o")
		objs = append(objs, obj)
		cmds = append(cmds, append([]string{"gcc"}, append(asan, "-c", filepath.Join(src, "aig/gia", f+".c"), "-o", obj)...))
	}
	cmds = append(cmds, append(append([]string{"g++", "-fsanitize=address"}, objs...), lib, "-lreadline", "-lpthread", "-lm", "-ldl", "-o", bin))
	for _, c := range cmds {
		if out, err := testoracle.Command(t, c[0], c[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c[:3], err, out)
		}
	}
	design, _ := filepath.Abs("test_vectors/i10_native_demo.aig")
	scripts := append([]string{"&if -K 5; &edge; &edge -o; &edge -o -r; &ps", "&if -K 6; &edge -E 2; &edge -o -E 2; &ps"}, edgeCrashScripts...)
	for _, script := range scripts {
		out, _ := testoracle.Command(t, bin, "&r "+design+"; "+script).CombinedOutput()
		if strings.Contains(string(out), "AddressSanitizer") || !strings.Contains(string(out), "and =") {
			t.Errorf("%s:\n%s", script, out)
		}
	}
}
