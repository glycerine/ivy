package abci

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

var blifStamp = regexp.MustCompile(`(?m)^# Benchmark .*$`)

// dsdOutputTime matches Dsd_Decompose's per-output runtime ("T= 0.01 s.").
var dsdOutputTime = regexp.MustCompile(`T= *[0-9.]+ s\.`)

// dsd (abcDsd.c) is ported: the same decomposition, statistics and
// written network as native ABC (whose crash was C bug 43).
func TestDsdLikeNative(t *testing.T) {
	dir := t.TempDir()
	read := func(path string) string {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return blifStamp.ReplaceAllString(string(data), "")
	}
	designs := []string{
		"test_vectors/native_s27_drw.blif",
		"test_vectors/native_mul4_gen.aig",
		"test_vectors/native_mul8_canon.aig",
		"test_vectors/seq/s713.blif",
	}
	cmds := []string{"dsd", "dsd -v", "dsd -r", "dsd -p", "dsd -ps", "collapse; dsd -g",
		"dsd -pv", "dsd -pvs", "collapse; dsd -g -v", "collapse; dsd -g -r -v"}
	for _, design := range designs {
		for _, cmd := range cmds {
			goFile, nativeFile := filepath.Join(dir, "go.blif"), filepath.Join(dir, "native.blif")
			script := "read " + design + "; strash; " + cmd + "; print_stats; write_blif "
			native := printSymmNormalize(dsdOutputTime.ReplaceAllString(runNativeABC(t, script+nativeFile), "T"))
			port := printSymmNormalize(dsdOutputTime.ReplaceAllString(runScript(t, script+goFile), "T"))
			if native != port {
				t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
			}
			if read(nativeFile) != read(goFile) {
				t.Errorf("%s: written network differs from native", script)
			}
		}
	}
}

// i10's many wide global functions drive the PRIME and CASE 4 paths;
// -pvs also prints C's "%c" short names past 'z' as single bytes.
func TestDsdLikeNativeI10(t *testing.T) {
	if testing.Short() {
		t.Skip("i10 global BDDs take seconds")
	}
	for _, cmd := range []string{"dsd -v", "dsd -pvs"} {
		script := "read test_vectors/i10_native_demo.aig; strash; " + cmd + "; print_stats"
		native := printSymmNormalize(dsdOutputTime.ReplaceAllString(runNativeABC(t, script), "T"))
		port := printSymmNormalize(dsdOutputTime.ReplaceAllString(runScript(t, script), "T"))
		if native != port {
			t.Errorf("%s: output differs from native", script)
		}
	}
}
