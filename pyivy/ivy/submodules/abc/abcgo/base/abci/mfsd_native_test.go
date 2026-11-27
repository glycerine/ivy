package abci

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// &mfsd (opt/sbd, Sbd_NtkPerform) was a reimplementation on aig.Graph whose
// node IDs, cut order and LUT synthesis differed from ABC's; on these
// designs it lost 4 AND nodes or panicked with "no LUT-sized cut". The
// port now follows sbdCore.c/sbdCut.c/sbdCut2.c/sbdWin.c/sbdLut.c on the
// GIA itself; its &ps, '-v' report (times masked) and &w output must equal
// native ABC's.

var mfsdDesigns = []string{"native_mul4_gen", "native_mul8_canon", "i10_c_rw", "i10_native_demo"}

// The port's &r keeps an AIGER symbol table that native &r drops, so the
// port's &w writes it back; that is the AIGER reader's difference, not
// &mfsd's, and is removed before comparing.
var aigerSymbolLines = regexp.MustCompile(`(?:[ilo][0-9]+ [^\n]*\n)+`)

var abcCommandLine = regexp.MustCompile(`^ABC command line: [^\n]*\n\n`)

// Times of 10 seconds and more are one character wider.
var mfsdTimes = regexp.MustCompile(`= *[0-9]+\.[0-9]+ sec`)

func mfsdSameAiger(native, port []byte) bool {
	native = giaStamp.ReplaceAll(native, nil)
	port = giaStamp.ReplaceAll(port, nil)
	if bytes.Equal(native, port) {
		return true
	}
	i := 0
	for i < len(native) && i < len(port) && native[i] == port[i] {
		i++
	}
	if loc := aigerSymbolLines.FindIndex(port[i:]); loc != nil && loc[0] == 0 {
		port = append(port[:i:i], port[i+loc[1]:]...)
	}
	return bytes.Equal(native, port)
}

func compareMfsdWithNative(t *testing.T, design, pre, opts string) {
	t.Helper()
	dir := t.TempDir()
	script := func(out string) string {
		return "&r test_vectors/" + design + ".aig; " + pre + "&mfsd " + opts + " -v; &ps; &w " + out
	}
	nativeFile, portFile := filepath.Join(dir, "native.aig"), filepath.Join(dir, "port.aig")
	clean := func(s, file string) string {
		s = abcCommandLine.ReplaceAllString(strings.ReplaceAll(s, file, "OUT"), "")
		return percents.ReplaceAllString(mfsdTimes.ReplaceAllString(mbTimes.ReplaceAllString(s, "M MB"), "= T sec"), "(P %)")
	}
	native := clean(runNativeABC(t, script(nativeFile)), nativeFile)
	port := clean(ansiEscape.ReplaceAllString(runScript(t, script(portFile)), ""), portFile)
	if native != port {
		t.Errorf("%s %s&mfsd %s\nnative:\n%s\nport:\n%s", design, pre, opts, native, port)
		return
	}
	nb, err := os.ReadFile(nativeFile)
	if err != nil {
		t.Fatal(err)
	}
	pb, err := os.ReadFile(portFile)
	if err != nil {
		t.Fatal(err)
	}
	if !mfsdSameAiger(nb, pb) {
		t.Errorf("%s %s&mfsd %s: &w output differs from native", design, pre, opts)
	}
}

func TestMfsdLikeNative(t *testing.T) {
	for _, design := range mfsdDesigns {
		compareMfsdWithNative(t, design, "", "")
	}
}

func TestMfsdOptionsLikeNative(t *testing.T) {
	for _, c := range []struct{ design, pre, opts string }{
		{"native_mul8_canon", "&if -K 4; ", "-p"},
		{"i10_native_demo", "&if -K 4; ", "-p"},
		{"native_mul8_canon", "", "-c"},
		{"i10_c_rw", "", "-c -m"},
		{"native_mul8_canon", "", "-d"},
		{"i10_native_demo", "", "-m"},
		{"native_mul4_gen", "", "-K 4 -S 6"},
		{"i10_c_rw", "", "-S 1"},
		{"i10_native_demo", "", "-S 2 -N 6"},
		{"native_mul8_canon", "", "-W 3"},
		{"i10_c_rw", "", "-F 2"},
		{"native_mul8_canon", "", "-C 10 -P 20"},
		{"i10_native_demo", "", "-M 50 -w"},
	} {
		compareMfsdWithNative(t, c.design, c.pre, c.opts)
	}
}

// The help says 2 <= K <= 6. Native &mfsd -K 5 and -K 6 aborted with "stack
// smashing detected" (Sbd_ProblemAddClauses' literal buffer held only
// SBD_SIZE_MAX+2 = 6 entries; CBUGS_FIXED.md #25), and -K 2 segfaulted on
// i10_c_rw (the 3-LUT structures of Sbd_ManExploreCut assume at most 3K-2
// cut leaves and read an unset variable otherwise; CBUGS_FIXED.md #27).
func TestMfsdLutSizesLikeNative(t *testing.T) {
	for _, c := range []struct{ design, pre, opts string }{
		{"i10_native_demo", "", "-K 5"},
		{"i10_c_rw", "", "-K 6"},
		{"native_mul4_gen", "", "-K 5 -c"},
		{"native_mul4_gen", "", "-K 3 -d"},
		{"i10_native_demo", "&if -K 6; ", "-K 6 -p"},
		{"i10_c_rw", "", "-K 2"},
		{"native_mul8_canon", "", "-K 3"},
		{"i10_native_demo", "", "-K 3"},
	} {
		compareMfsdWithNative(t, c.design, c.pre, c.opts)
	}
}

// TestNativeMfsdStaysInBounds compiles src/opt/sbd/sbdWin.c, sbdLut.c and
// sbdCore.c with AddressSanitizer, links them ahead of the native library
// named by ABC_LIBABC, and runs &mfsd with every LUT size. Sbd_ManSolve read
// past s_Truths6 (CBUGS_FIXED.md #26) and Sbd_ProblemAddClauses overflowed
// its literal buffer for K >= 5 (#25).
func TestNativeMfsdStaysInBounds(t *testing.T) {
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
	asan := []string{"-g", "-O1", "-fsanitize=address", "-fno-omit-frame-pointer", "-DABC_USE_STDINT_H", "-DNDEBUG", "-I", src}
	objs := []string{filepath.Join(dir, "main.o")}
	cmds := [][]string{append([]string{"gcc"}, append(asan, "-c", main, "-o", objs[0])...)}
	for _, f := range []string{"sbdWin", "sbdLut", "sbdCore"} {
		obj := filepath.Join(dir, f+".o")
		objs = append(objs, obj)
		cmds = append(cmds, append([]string{"gcc"}, append(asan, "-c", filepath.Join(src, "opt/sbd", f+".c"), "-o", obj)...))
	}
	cmds = append(cmds, append(append([]string{"g++", "-fsanitize=address"}, objs...), lib, "-lreadline", "-lpthread", "-lm", "-ldl", "-o", bin))
	for _, c := range cmds {
		if out, err := exec.Command(c[0], c[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", c[:3], err, out)
		}
	}
	design, _ := filepath.Abs("test_vectors/i10_native_demo.aig")
	for _, script := range []string{"&mfsd", "&if -K 6; &mfsd", "&mfsd -K 2", "&mfsd -K 3", "&mfsd -K 5", "&mfsd -K 6"} {
		out, _ := exec.Command(bin, "&r "+design+"; "+script+"; &ps").CombinedOutput()
		if strings.Contains(string(out), "AddressSanitizer") || !strings.Contains(string(out), "and =") {
			t.Errorf("%s:\n%s", script, out)
		}
	}
}

// &mfsd treats a MUX node as the AND of its first two fanins (C only
// asserts !Gia_ObjIsMux in Sbd_StoMergeCuts) and Sbd_ManDerive rebuilds it
// as an AND, so on an AND/XOR/MUX graph from '&st -m' native ABC silently
// returned a non-equivalent AIG, even when it changed nothing
// ('&st -m; &mfsd -m -M 1; &st' failed &cec against the input), and the port
// failed the assertion (CBUGS_FIXED.md #28). The command now refuses such a
// GIA, like one with barrier buffers.
func TestMfsdRejectsMuxes(t *testing.T) {
	script := "&r test_vectors/native_mul8_canon.aig; &st -m; &mfsd -m -M 1"
	want := "Abc_CommandAbc9Mfsd(): This command does not work with MUXes."
	if bin := os.Getenv("ABC_BINARY"); bin != "" {
		bin, _ = filepath.Abs(bin)
		got, _ := exec.Command(bin, "-c", script).CombinedOutput()
		if !strings.Contains(string(got), want) {
			t.Errorf("native:\n%s", got)
		}
	}
	got, status := runScriptStatus(t, script)
	if !strings.Contains(got, want) || status == 0 {
		t.Errorf("port (status %d):\n%s", status, got)
	}
}
