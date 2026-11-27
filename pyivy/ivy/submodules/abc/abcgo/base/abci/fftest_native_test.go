package abci

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// &fftest was a restructured reimplementation of bmcFault.c: no header or
// options line, Mf-based CNF instead of Cnf_Derive (529 instead of 341
// variables on mul4), its own cardinality encoders and solver variable
// allocation, no "Reading N pre-computed test patterns" lines, a
// condensed untestable-fault check and different file contents. It now
// is a line-by-line port, so the whole output must equal native ABC's
// (times masked), as must every file the command writes.
func TestFftestLikeNative(t *testing.T) {
	if os.Getenv("ABC_BINARY") == "" {
		t.Skip("set ABC_BINARY to a native abc built from ../src")
	}
	dir := t.TempDir()
	designs := map[string]string{}
	for _, name := range []string{"native_mul4_gen.aig", "native_mul8_canon.aig", "native_s27_drw.blif"} {
		data, err := os.ReadFile(filepath.Join("test_vectors", name))
		if err != nil {
			t.Fatal(err)
		}
		designs[name] = filepath.Join(dir, name)
		if err := os.WriteFile(designs[name], data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	mul4, mul8 := designs["native_mul4_gen.aig"], designs["native_mul8_canon.aig"]
	s27 := "read " + designs["native_s27_drw.blif"] + "; strash; &get; &fftest "
	// a pattern file for the <file> argument
	pats := filepath.Join(dir, "pats.txt")
	if err := os.WriteFile(pats, []byte("00000000\n11111111\n01010101\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// fault model given in a file (-F)
	form := filepath.Join(dir, "form.txt")
	if err := os.WriteFile(form, []byte("((a&b)^p)\n"), 0644); err != nil {
		t.Fatal(err)
	}
	// Runs ending in a timeout (-T) are left out: where they stop depends
	// on the speed of the solver.
	scripts := []string{
		"&r " + mul4 + "; &fftest -A 2 -N 3",
		"&r " + mul4 + "; &fftest -A 2",
		"&r " + mul4 + "; &fftest -A 3",
		"&r " + mul4 + "; &fftest -A 2 -N 3 -v",
		"&r " + mul4 + "; &fftest -A 3 -N 2 -u",
		"&r " + mul4 + "; &fftest -A 2 -c",
		"&r " + mul4 + "; &fftest -A 2 -u -v",
		"&r " + mul4 + "; &fftest -A 2 -b -c",
		"&r " + mul4 + "; &fftest -A 3 -K 2",
		"&r " + mul4 + "; &fftest -A 3 -K 3 -k -N 2 -c -d",
		"&r " + mul4 + "; &fftest -A 2 -s",
		"&r " + mul4 + "; &fftest -A 2 -d -u " + pats,
		"&r " + mul4 + "; &fftest -A 3 -s -d -c -N 4",
		"&r " + mul4 + "; &fftest -A 0 -S (((a&b)&~p)|q)",
		"&r " + mul4 + "; &fftest -A 0 -S ((a&b)^p) -d -u",
		"&r " + mul4 + "; &fftest -A 0 -S (((a&~p)|q)&b) -K 2 -u",
		"&r " + mul4 + "; &fftest -A 0 -F " + form + " -u",
		"&r " + mul4 + "; &fftest -A 0 -S (a&b)",
		"&r " + mul4 + "; &fftest -A 0 -S (a&b",
		s27 + "-A 1 -v -c",
		s27 + "-A 1 -d -e",
		s27 + "-A 1 -s -u -v",
		s27 + "-A 0 -S ((a&b)^p) -f",
		s27 + "-A 0 -S (a?(b?~s:r):(b?q:p)) -u -v",
		s27 + "-A 4 -N 3 -u -d",
		s27 + "-A 4 -b -c",
		s27 + "-A 0 -S (((a^p)&(b^q))^r) -c",
		"&r " + mul8 + "; &fftest -A 3 -b -d",
	}
	for _, script := range scripts {
		native, nativeFiles := fftestRun(t, dir, script, runNativeABC)
		port, portFiles := fftestRun(t, dir, script, runScript)
		if c, g := fftestClean(native), fftestClean(port); c != g {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, c, g)
		}
		if len(nativeFiles) == 0 && strings.Contains(script, " -d") {
			t.Errorf("%s: no test file written", script)
		}
		for name, want := range nativeFiles {
			if got, ok := portFiles[name]; !ok || got != want {
				t.Errorf("%s: file %s differs:\nnative %q\nport   %q", script, name, want, got)
			}
		}
		for name := range portFiles {
			if _, ok := nativeFiles[name]; !ok {
				t.Errorf("%s: port wrote extra file %s", script, name)
			}
		}
	}
}

// fftestRun runs script and collects (then removes) the *_tests.txt and
// *_untest.txt files it writes next to the designs.
func fftestRun(t *testing.T, dir, script string, run func(*testing.T, string) string) (string, map[string]string) {
	t.Helper()
	out := run(t, script)
	files := map[string]string{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "_tests.txt") || strings.HasSuffix(e.Name(), "_untest.txt") {
			name := filepath.Join(dir, e.Name())
			data, err := os.ReadFile(name)
			if err != nil {
				t.Fatal(err)
			}
			files[e.Name()] = string(data)
			os.Remove(name)
		}
	}
	return out, files
}

func fftestClean(s string) string {
	return secTimes.ReplaceAllString(bmc3Normalize(s), "T sec")
}

// &fftest -N n with a formula (-A 0) or delay (-A 1) fault model, which
// keep no fault map, called Gia_ManFaultAnalyze with an empty vMap and
// read past it: native ABC segfaulted (CBUGS_FIXED.md #24). Fixed, the
// check is skipped for these models and the run is as without -N.
func TestNativeFftestIterCheckWithoutFaultMap(t *testing.T) {
	for _, script := range []string{
		"&r test_vectors/native_mul4_gen.aig; &fftest -A 0 -S ((a&b)^p) -N 2",
		"&r test_vectors/native_mul4_gen.aig; &fftest -A 0 -S (((a&b)&~p)|q) -N 1 -c",
		"read test_vectors/native_s27_drw.blif; strash; &get; &fftest -A 1 -N 2",
	} {
		native := fftestClean(runNativeABC(t, script))
		port := fftestClean(runScript(t, script))
		if native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
		without := fftestClean(runNativeABC(t, strings.Replace(strings.Replace(script, " -N 2", "", 1), " -N 1", "", 1)))
		if native != without {
			t.Errorf("%s: -N changed the run:\n%s\nwithout -N:\n%s", script, native, without)
		}
	}
}
