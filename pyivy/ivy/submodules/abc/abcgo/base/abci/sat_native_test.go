package abci

import (
	"fmt"
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestMiterSatLikeNative compares 'sat -v' (and the counter-example that
// write_cex -n saves) with native ABC's. The port's Abc_NtkMiterSat built
// its own Tseitin CNF on sat_solver2 for strashed networks instead of
// abcSat.c's Abc_NtkMiterSatCreateInt (supergates, MUX clauses, BFS variable
// order) on sat_solver, and for BDD logic networks it numbered variables by
// node and skipped the PO clauses; it found other models, printed a one-line
// summary instead of the MINISAT table and Sat_SolverPrintStats, and so made
// saucy3 -v (which solves its miters through Abc_NtkMulti and
// Abc_NtkMiterSat) diverge.
func TestMiterSatLikeNative(t *testing.T) {
	vectors, err := filepath.Abs("test_vectors")
	if err != nil {
		t.Fatal(err)
	}
	// Run in a scratch directory: the BDD-logic path writes trace.cnf, as
	// the C code does.
	t.Chdir(t.TempDir())
	cex := filepath.Join(t.TempDir(), "cex.txt")
	mul8 := filepath.Join(vectors, "native_mul8_canon.aig")
	mul4 := filepath.Join(vectors, "native_mul4_gen.aig")
	i10 := filepath.Join(vectors, "i10_native_demo.aig")
	s713 := filepath.Join(vectors, "seq", "s713.blif")
	comb := filepath.Join(t.TempDir(), "s713_comb.blif")
	scripts := []string{
		"read " + mul8 + "; strash; cone -O 0; sat -v; write_cex -n " + cex,
		"read " + mul8 + "; strash; cone -O 3; sat -v; write_cex -n " + cex,
		"read " + mul8 + "; strash; cone -O 7; sat; write_cex -n " + cex,
		"read " + mul8 + "; strash; cone -O 12; sat -v; write_cex -n " + cex,
		"read " + mul8 + "; strash; cone -O 3; logic; sat -v; write_cex -n " + cex,
		"read " + mul4 + "; strash; sat -v; write_cex -n " + cex,
		"read " + i10 + "; strash; cone -O 20; sat -v; write_cex -n " + cex,
		"read " + s713 + "; comb; strash; write_blif " + comb + "; dc2; miter " + comb + "; sat -v",
		"miter " + mul8 + " " + filepath.Join(vectors, "cec_mul8_mut.aig") + "; sat -v; write_cex -n " + cex,
		// saucy3 on mul8 ends in exit(1) after a SAT timeout (both C and Go),
		// which would end the test binary: use the 4-bit multipliers.
		"read " + mul4 + "; strash; saucy3 -v",
		"read " + filepath.Join(vectors, "cec_mul4_syn2.aig") + "; strash; saucy3 -v",
	}
	times := regexp.MustCompile(`[0-9]+\.[0-9]+ ?(sec|MB|Mb|%)`)
	bin := testoracle.NativeABCPath()
	if bin == "" {
		t.Skip("set ABC_BINARY to a native abc built from ../src")
	}
	// Compare exit statuses too.
	run := func(script string, native bool) string {
		os.Remove(cex)
		var out string
		var status int
		if native {
			data, err := testoracle.Command(t, bin, "-c", script).CombinedOutput()
			if _, ok := err.(*testoracle.ExitError); err != nil && !ok {
				t.Fatal(err)
			}
			out = ansiEscape.ReplaceAllString(string(data), "")
			if i := strings.Index(out, "\n\n"); strings.HasPrefix(out, "ABC command line") && i >= 0 {
				out = out[i+2:]
			}
			if err != nil {
				status = err.(*testoracle.ExitError).ExitCode()
			}
		} else {
			out, status = runScriptStatus(t, script)
		}
		out += fmt.Sprintf("--- status %d ---\n", status)
		if data, err := testoracle.ReadFile(cex); err == nil {
			out += "--- cex ---\n" + string(data)
		}
		return times.ReplaceAllString(out, "X")
	}
	for _, script := range scripts {
		want := run(script, true)
		if got := run(script, false); got != want {
			t.Errorf("%s:\nnative:\n%s\ngo:\n%s", script, want, got)
		}
	}
}

// abcSaucy.c's reduceDB advances its write index only when it is behind the
// read index ("else if (j < i) { ...; j++; }"), so every pass also drops
// the first surviving counter-example. The port compacted the list
// properly, which changed saucy3 -v's search on i10 after ~460 lines
// ("Backtrack by SAT from level 54 to 52" in C, "to 51" in Go).
func TestSaucyReduceDBLikeC(t *testing.T) {
	s := &saucy{satCounterExamples: Vec_PtrAlloc(MAX_LEARNTS), activityInc: float64(MAX_LEARNTS)}
	// The first pass's limit is activityInc/size = 1 and drops entry 2
	// (and, as in C, entry 0); each later pass (limits 2, 4, ..., all below
	// 1e6) drops the first entry until at most 0.7*MAX_LEARNTS remain.
	for i := 0; i < MAX_LEARNTS; i++ {
		activity := 1e6
		if i == 2 {
			activity = 0
		}
		Vec_PtrPush(s.satCounterExamples, &sim_result{activity: activity, inVecSignature: int32(i)})
	}
	reduceDB(s)
	if got, want := Vec_PtrSize(s.satCounterExamples), int32(35); got != want {
		t.Fatalf("size after reduceDB = %d, want %d", got, want)
	}
	for i := int32(0); i < Vec_PtrSize(s.satCounterExamples); i++ {
		if got := Vec_PtrEntry(s.satCounterExamples, i).(*sim_result).inVecSignature; got != i+15 {
			t.Fatalf("entry %d is counter-example %d, want %d", i, got, i+15)
		}
	}
}
