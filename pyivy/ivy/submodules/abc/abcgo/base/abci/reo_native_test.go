package abci

import (
	"github.com/glycerine/abcgo/internal/testoracle"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// The commands below score or minimize BDDs with the REO engine
// (bdd/reo, a port of src/bdd/reo): renode -b scores each cut by the BDD
// size after Extra_Reorder and reorders the local functions of the result
// (Abc_NtkBddReorder); reorder and eliminate -s call Abc_NtkBddReorder.
// The Go port used to rebuild the diagrams on every swap with its own
// sifting, which gave other orders (i10: nd = 642 instead of 657).
var reoNativeBenches = []string{
	"test_vectors/native_mul4_gen.aig",
	"test_vectors/seq/s713.blif",
	"test_vectors/seq/s1494.blif",
}

var reoNativeScripts = []string{
	"renode -b",
	"renode -b -v",
	"renode -b -K 6",
	"renode; bdd; reorder -v",
	"if -K 6; eliminate -s",
}

// the larger benchmarks only with renode -b, to keep the test fast
var reoNativeLarge = []string{
	"test_vectors/i10_native_demo.aig",
	"test_vectors/native_mul8_canon.aig",
}

var reoTimeOrMemory = regexp.MustCompile(`[0-9.]+ (sec|MB)`)

// reoComparable drops the command-line echo and masks times and memory.
func reoComparable(out string) string {
	var lines []string
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, "ABC command line:") {
			continue
		}
		lines = append(lines, reoTimeOrMemory.ReplaceAllString(line, "<t>"))
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// reoBlifBody is a BLIF file without its comment header.
func reoBlifBody(t *testing.T, file string) string {
	t.Helper()
	data, err := testoracle.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	var lines []string
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

func TestReoCommandsMatchNative(t *testing.T) {
	if testoracle.NativeABCPath() == "" {
		t.Skip("set ABC_BINARY to a native abc built from ../src")
	}
	dir := t.TempDir()
	type job struct{ bench, cmd string }
	var jobs []job
	for _, bench := range reoNativeBenches {
		for _, cmd := range reoNativeScripts {
			jobs = append(jobs, job{bench, cmd})
		}
	}
	for _, bench := range reoNativeLarge {
		jobs = append(jobs, job{bench, "renode -b"})
	}
	for _, j := range jobs {
		bench, cmd := j.bench, j.cmd
		{
			script := func(out string) string {
				return "read " + bench + "; strash; " + cmd + "; print_stats; write_blif " + out
			}
			cBlif, goBlif := filepath.Join(dir, "c.blif"), filepath.Join(dir, "go.blif")
			want := reoComparable(runNativeABC(t, script(cBlif)))
			got := reoComparable(runScript(t, script(goBlif)))
			if got != want {
				t.Errorf("%s | %s: output differs\nnative:\n%s\ngo:\n%s", filepath.Base(bench), cmd, want, got)
				continue
			}
			if reoBlifBody(t, goBlif) != reoBlifBody(t, cBlif) {
				t.Errorf("%s | %s: written BLIF differs", filepath.Base(bench), cmd)
			}
		}
	}
}

// TestRenodeBddMatchesNativeStats needs no native binary: the statistics
// of renode -b printed by native ABC.
func TestRenodeBddMatchesNativeStats(t *testing.T) {
	for _, tc := range []struct{ bench, want string }{
		{"test_vectors/i10_native_demo.aig", "nd =   687  edge =   2348  bdd  =  2369  lev = 18"},
		{"test_vectors/native_mul4_gen.aig", "nd =    15  edge =     67  bdd  =   156  lev = 3"},
		{"test_vectors/seq/s1494.blif", "nd =   131  edge =    641  bdd  =   762  lev = 5"},
	} {
		out := runScript(t, "read "+tc.bench+"; strash; renode -b; print_stats")
		if !strings.Contains(statsLine(out), tc.want) {
			t.Errorf("%s: renode -b gives\n%s\nnative: %s", tc.bench, statsLine(out), tc.want)
		}
	}
}
