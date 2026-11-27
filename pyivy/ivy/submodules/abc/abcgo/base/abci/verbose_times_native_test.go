package abci

import (
	"regexp"
	"strings"
	"testing"
)

// memory figures depend on the allocator, and every "N.NN sec" is a time
var (
	mbTimes  = regexp.MustCompile(`[0-9]+\.[0-9]+ MB`)
	secTimes = regexp.MustCompile(`[0-9]+\.[0-9]+ sec`)
	percents = regexp.MustCompile(`\( *[0-9.]+ %\)`)
)

// Times are printed with Abc_PrintTime ("Time =%9.2f sec"); several ports
// printed a Go duration ("Time = 7.92ms", "Time: 1.2s") and omitted lines
// ABC prints (&jf's "Computing cuts..." and "Memory:", &lf/&nf's mux stats).
func TestVerboseReportsLikeNative(t *testing.T) {
	for _, script := range []string{
		"read ../../../i10.aig; &get; &jf -v; &ps",
		"read ../../../i10.aig; &get; &kf -v; &ps",
		"read ../../../i10.aig; &get; &lf -v; &ps",
		"read test_vectors/native_s27_drw.blif; strash; &get; &sim",
		"read test_vectors/seq/s713.blif; strash; &get; &sim -F 4",
		"read ../../../i10.aig; drwsat -v; print_stats",
		"read ../../../i10.aig; drwsat -v -b; print_stats",
		"read test_vectors/seq/s713.blif; strash; scorr -v; print_stats",
		"read test_vectors/seq/s713.blif; strash; lcorr -v; print_stats",
		"read test_vectors/seq/s1494.blif; strash; lcorr -v; print_stats",
		"read test_vectors/seq/s1494.blif; strash; scorr -v; print_stats",
		"read test_vectors/seq/s713.blif; strash; dsec -v; print_status",
		"read test_vectors/native_s27_drw.blif; strash; dsec -v; print_status",
	} {
		port := bmc3Normalize(runScript(t, script))
		if strings.Contains(port, "ms\n") || strings.Contains(port, "Time: ") {
			t.Errorf("%s: Go duration in output:\n%s", script, port)
		}
		native := bmc3Normalize(runNativeABC(t, script))
		clean := func(s string) string {
			return percents.ReplaceAllString(secTimes.ReplaceAllString(mbTimes.ReplaceAllString(s, "M MB"), "T sec"), "(P %)")
		}
		if clean(native) != clean(port) {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}

// drwsat -v labels its stages as Dar_ManRwsat does and prints the rewrite
// and refactor statistics; the port printed "Starting: 3157 nodes".
func TestDrwsatVerboseStages(t *testing.T) {
	out := runScript(t, "read ../../../i10.aig; drwsat -v; print_stats")
	for _, want := range []string{"Starting:  ../../../i10", "Rewrite:   ../../../i10", "Refactor:  ../../../i10", "NodesBeg =     2123."} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
}

// dsec -v prints Fra_FraigSec's stage lines; the port printed only the
// verdict.
func TestDsecVerboseStages(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s713.blif; strash; dsec -v; print_status")
	for _, want := range []string{
		"Original miter:       Latches =    38. Nodes =    376.\n",
		"Sequential cleanup:   Latches =    28. Nodes =    281. Time =",
		"Forward retiming:     Latches =    28. Nodes =    281. Time =",
		"Latch-corr (I=  1):   Latches =     0. Nodes =      0. Time =",
		"Fraiging:             Latches =     0. Nodes =      0. Time =",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q:\n%s", want, out)
		}
	}
}

// Ssw_SatStart allocates 1000 SAT variables and Ssw_CnfNodeAddToSolver
// grows them by hundreds (sat_solver_setnvars), which fixes the decision
// order; the port added them one at a time, so its counterexamples and
// refinement differed (scorr on s1494 took 7 iterations to ABC's 6).
func TestScorrIterationsLikeABC(t *testing.T) {
	out := runScript(t, "read test_vectors/seq/s1494.blif; strash; scorr -v; print_stats")
	if !strings.Contains(out, "  1 : C =     10. Cl =     34. LR =    73.") || strings.Contains(out, "  6 : C =") {
		t.Fatalf("scorr -v:\n%s", out)
	}
}
