package abci

import (
	"regexp"
	"strings"
	"testing"

	base_main "github.com/glycerine/abcgo/base/main"
)

// glaClean masks what differs between runs of &gla: times, percentages and
// memory estimates (including the width of the masked numbers).
var (
	glaSecs     = regexp.MustCompile(`[0-9]+\.[0-9]+ sec`)
	glaPercents = regexp.MustCompile(`\( *[0-9.]+ %\)`)
	glaMB       = regexp.MustCompile(` *[0-9]+(\.[0-9]+)? MB`)
)

func glaClean(s string) string {
	s = glaMB.ReplaceAllString(glaPercents.ReplaceAllString(glaSecs.ReplaceAllString(s, "T sec"), "(P %)"), " M MB")
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		// a progress line ends in "\r" and is overwritten (and native abc -c
		// runs in batch mode, which skips some of them)
		if k := strings.LastIndex(line, "\r"); k >= 0 {
			line = line[k+1:]
		}
		line = strings.TrimRight(line, " ")
		if strings.HasPrefix(line, "ABC command line") || line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// runGlaScript runs script in batch mode, as "abc -c" does: the frame
// progress lines and commands such as &gla_derive depend on it.
func runGlaScript(t *testing.T, script string) string {
	t.Helper()
	out, status := runScriptStatusWith(t, script, func(f *base_main.Frame) { f.BatchMode = true })
	if status != 0 {
		t.Fatalf("script %q failed:\n%s", script, out)
	}
	return out
}

func compareGlaWithNative(t *testing.T, script string) {
	t.Helper()
	native := glaClean(runNativeABC(t, script))
	port := glaClean(runGlaScript(t, script))
	if native != port {
		t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
	}
}

// &gla (Gia_ManPerformGla, absGla.c): the port printed an invented header
// ("Running gate-level abstraction: FrameMax = ...") and none of the result
// lines ("True counter-example detected in frame %d.", "GLA finished %d
// frames and produced a %d-stable abstraction."), and it drove the SAT
// solver through a wrapper whose solve propagated root implications before
// the assumptions and never reduced learned clauses, so Vars/Clas/Lrns/
// Confl and the abstraction differed.
func TestGlaLikeNative(t *testing.T) {
	for _, script := range []string{
		"&r test_vectors/native_mul4_gen.aig; &gla; &ps",
		"&r test_vectors/native_mul4_gen.aig; &gla -v; &ps",
		"read test_vectors/native_s27_drw.blif; strash; &get; &gla; &ps",
		"read test_vectors/native_s27_drw.blif; strash; &get; &gla -v; &ps",
		"read test_vectors/native_s27_drw.blif; strash; &get; &gla -F 10 -v; &ps",
		"read test_vectors/seq/s1494.blif; strash; &get; &gla -F 10; &ps",
		"read test_vectors/seq/s1494.blif; strash; &get; &gla -v -F 20; &ps",
		"read test_vectors/seq/s1494.blif; strash; &get; &gla -F 10; &gla_derive; &ps",
		"read test_vectors/seq/s1494.blif; strash; orpos; &get; &gla -F 10; &gla_derive; &ps",
		"read test_vectors/seq/s1494.blif; strash; orpos; &get; &gla -F 6 -v; &gla_shrink -v; &gla_derive; &ps",
		"read test_vectors/seq/s713.blif; strash; orpos; &get; &gla -F 10 -v; &gla_derive; &ps",
		"read test_vectors/seq/s713.blif; strash; orpos; &get; &gla -F 10 -v; &gla_vta; &vta_gla; &gla_derive; &ps",
		"read test_vectors/native_s27_drw.blif; strash; dc2; miter test_vectors/native_s27_drw.blif; &get; &gla -v -F 20",
		"read test_vectors/seq/s713.blif; strash; dc2; miter test_vectors/seq/s713.blif; &get; &gla -v -F 12",
	} {
		compareGlaWithNative(t, script)
	}
}

// An output of s713 is driven directly by a primary input. Gia_ManPerformGla
// gave it no abstraction ID, so Ga2_ObjFindLit read vIds[-1] (C: undefined,
// printed "GLA finished ..."; port: panic). CBUGS_FIXED.md #33.
func TestGlaPrimaryInputOutputLikeNative(t *testing.T) {
	for _, script := range []string{
		"read test_vectors/seq/s713.blif; strash; &get; &gla -F 10; &ps",
		"read test_vectors/seq/s713.blif; strash; &get; &gla -v -F 20; &ps",
	} {
		compareGlaWithNative(t, script)
		if out := runGlaScript(t, script); !strings.Contains(out, "Sequential miter is trivially SAT.") {
			t.Errorf("%s: no trivially SAT output:\n%s", script, out)
		}
	}
}

// &gla -n (Gia_ManPerformGlaOld, absGlaOld.c): Rnm_ManRefine filtered every
// non-trivial refinement with Rnm_ManFilterSelected, which reads the Ga2 cut
// mapping that Gla_ManStart never creates, so native abc segfaulted on all
// of these (CBUGS_FIXED.md #34); the port had invented that mapping.
func TestGlaOldLikeNative(t *testing.T) {
	for _, script := range []string{
		"read test_vectors/native_s27_drw.blif; strash; &get; &gla -n; &ps",
		"read test_vectors/native_s27_drw.blif; strash; &get; &gla -n -v; &ps",
		"read test_vectors/native_s27_drw.blif; strash; &get; &gla -n -F 10 -v; &ps",
		"read test_vectors/seq/s1494.blif; strash; &get; &gla -n -F 10; &ps",
		"read test_vectors/seq/s1494.blif; strash; &get; &gla -n -F 10 -v; &ps",
		"read test_vectors/seq/s713.blif; strash; &get; &gla -n -F 10; &ps", // a PI-driven output: CBUGS_FIXED.md #33
		"read test_vectors/seq/s713.blif; strash; orpos; &get; &gla -n -F 10 -v; &gla_derive; &ps",
		"read test_vectors/seq/s1494.blif; strash; orpos; &get; &gla -n -F 10 -v; &gla_derive; &ps",
		"&r test_vectors/native_mul4_gen.aig; &gla -n; &ps",
	} {
		compareGlaWithNative(t, script)
	}
}

// &gla -a adds every non-PI pseudo-input; when all pseudo-inputs are PIs it
// added an empty set and started the proof logger with width 0, so
// Prf_ManSize divided by zero (native SIGFPE) in both engines
// (CBUGS_FIXED.md #35). The counter-example is now checked as without -a.
func TestGlaAddLayerAllPisLikeNative(t *testing.T) {
	for _, script := range []string{
		"read_truth 8; strash; &get; &gla -a -F 1 -v",
		"read_truth 8; strash; &get; &gla -n -a -F 1 -v",
		"read test_vectors/seq/s1494.blif; strash; &get; &gla -a -F 10; &ps",
		"read test_vectors/seq/s1494.blif; strash; &get; &gla -n -a -F 10; &ps",
	} {
		compareGlaWithNative(t, script)
	}
}

// -T: the frame reached before the timeout depends on the machine, so the
// frame counts of the timeout lines are masked.
func TestGlaTimeoutLikeNative(t *testing.T) {
	frames := regexp.MustCompile(`(frame|completed) [0-9]+ (frames )?with a [0-9]+-stable`)
	for _, script := range []string{
		"read test_vectors/seq/s713.blif; strash; &get; &gla -T 1; &ps",
		"read test_vectors/seq/s713.blif; strash; &get; &gla -n -T 1; &ps",
		"read test_vectors/seq/s713.blif; strash; orpos; &get; &gla -T 1; &gla_derive; &ps",
	} {
		native := frames.ReplaceAllString(glaClean(runNativeABC(t, script)), "$1 N with a N-stable")
		port := frames.ReplaceAllString(glaClean(runGlaScript(t, script)), "$1 N with a N-stable")
		if native != port {
			t.Errorf("%s\nnative:\n%s\nport:\n%s", script, native, port)
		}
	}
}
