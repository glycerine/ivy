package abci

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	base_main "github.com/glycerine/abcgo/base/main"
	"github.com/glycerine/abcgo/map/mapper"
	"github.com/glycerine/abcgo/map/mio"
	"github.com/glycerine/abcgo/map/super"
)

// superStamp masks the volatile lines of Super_WriteFileHeaderStr
// (superGate.c): the time stamp and the generation time.
var superStamp = regexp.MustCompile(`(?m)( on .*\.$|^# Generation time .*$)`)

// superLikeNativeFile runs "super <flags> lib.genlib" on a private copy of
// the genlib (super writes lib.super next to it) and returns the written
// supergate library with the directory and volatile lines masked.
func superLikeNativeFile(t *testing.T, lib, flags string, native bool) string {
	t.Helper()
	dir := t.TempDir()
	data, err := os.ReadFile(lib)
	if err != nil {
		t.Fatal(err)
	}
	genlib := filepath.Join(dir, "lib.genlib")
	if err = os.WriteFile(genlib, data, 0644); err != nil {
		t.Fatal(err)
	}
	script := "super " + flags + " " + genlib
	if native {
		runNativeABC(t, script)
	} else {
		out, status := runScriptStatusWith(t, script, func(f *base_main.Frame) { super.Super_Init(f) })
		if status != 0 {
			t.Fatalf("%s: status %d\n%s", script, status, out)
		}
	}
	res, err := os.ReadFile(filepath.Join(dir, "lib.super"))
	if err != nil {
		t.Fatal(err)
	}
	return superStamp.ReplaceAllString(strings.ReplaceAll(string(res), dir, "DIR"), "")
}

// super (superGate.c) enumerates supergates in the order of its stmm hash
// table and glibc qsort; the written library, including the fanin order of
// every gate, the header statistics and the trailing NUL, must be identical.
func TestSuperLikeNative(t *testing.T) {
	cases := map[string][]string{
		"../../map/super/test_vectors/tiny.genlib":    {"-I 5 -L 1", "-I 2 -L 1", "-I 3 -L 2", "-I 4 -L 2", "-I 3 -L 3", "-I 4 -L 2 -s", "-I 5 -L 2", "-I 5 -L 2 -D 3", "-I 4 -L 3 -A 3", "-I 5 -L 2 -N 20", "-I 6 -L 1", "-I 6 -L 2 -D 3"},
		"../../map/amap/test_vectors/amap_mux.genlib": {"-I 5 -L 1", "-I 3 -L 2", "-I 4 -L 2 -s", "-I 5 -L 2 -D 3", "-I 4 -L 2 -A 3", "-I 6 -L 1"},
	}
	for _, lib := range []string{"../../map/super/test_vectors/tiny.genlib", "../../map/amap/test_vectors/amap_mux.genlib"} {
		for _, flags := range cases[lib] {
			native := superLikeNativeFile(t, lib, flags, true)
			port := superLikeNativeFile(t, lib, flags, false)
			if native != port {
				t.Errorf("%s: super %s differs\nnative:\n%q\nport:\n%q", lib, flags, native, port)
			}
		}
	}
}

// map with the supergate library derived from the genlib
// (Map_SuperLibDeriveFromGenlib runs Super_PrecomputeStr) must give the same
// mapped network, fanin order included.
func TestMapTinyGenlibLikeNative(t *testing.T) {
	designs := map[string]string{
		"mul8": "test_vectors/bench/mul8.aig",
		"i10":  "../../../i10.aig",
	}
	dir := t.TempDir()
	read := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return blifStamp.ReplaceAllString(string(data), "")
	}
	for _, run := range []struct{ lib, name, flags string }{
		{mfs3Genlib, "mul8", ""}, {mfs3Genlib, "mul8", "-a"}, {mfs3Genlib, "mul8", "-D 60"}, {mfs3Genlib, "mul8", "-D 100"},
		{mfs3Genlib, "i10", ""}, {mfs3Genlib, "i10", "-a"}, {mfs3Genlib, "i10", "-D 60"}, {mfs3Genlib, "i10", "-D 100"},
		{"../../map/amap/test_vectors/amap_mux.genlib", "mul8", ""}, {"../../map/amap/test_vectors/amap_mux.genlib", "i10", "-a"},
	} {
		lib, name, flags := run.lib, run.name, run.flags
		goFile, nativeFile := filepath.Join(dir, "go.blif"), filepath.Join(dir, "native.blif")
		script := "read_genlib " + lib + "; read " + designs[name] + "; strash; map " + flags + "; print_stats; write_blif "
		native := printSymmNormalize(runNativeABC(t, script+nativeFile))
		out, _ := runScriptStatusWith(t, script+goFile, func(f *base_main.Frame) { mapper.Map_Init(f); mio.Mio_Init(f) })
		port := printSymmNormalize(out)
		if native != port {
			t.Errorf("%s / map %s: output differs\nnative:\n%s\nport:\n%s", name, flags, native, port)
			continue
		}
		if nb, gb := read(nativeFile), read(goFile); nb != gb {
			nl, gl := strings.Split(nb, "\n"), strings.Split(gb, "\n")
			diff := 0
			for i := range nl {
				if i >= len(gl) || nl[i] != gl[i] {
					diff++
				}
			}
			t.Errorf("%s / map %s: written BLIF differs (%d of %d lines)", name, flags, diff, len(nl))
		}
	}
}
