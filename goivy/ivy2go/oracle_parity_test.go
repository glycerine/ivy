package ivy2go

// oracle_parity_test.go is Phase C Task 0 of the divergence-remediation
// plan. It compares the binary stdout of ivy2cpp-emitted C++ vs
// ivy2go-emitted Go for the 14 fixtures under
// `ivy2cpp/test_vec/oracle/`. The harness reports per-fixture status
// and a `N parity / M blocked-go / Y blocked-cpp / Z divergent`
// summary line.
//
// First iteration uses `iters=0` (no test loop iterations). Both
// binaries should emit only `test_completed`, so byte-for-byte stdout
// match is meaningful without worrying about RNG divergence between
// Go's `math/rand` and C++'s `std::mt19937`. Future iterations can
// add an action-count parity tier once Tier 1 OMITTED entries are
// closed.
//
// Gated behind `ORACLE=1 SLOW_GO_TEST=1` because the harness runs
// `go build` + `g++` (with Z3 link) on every fixture — too expensive
// for the default test suite. Required toolchain: `g++`, `z3` library
// (Linux) or framework (macOS), plus the in-tree Go module.

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy/ivy2cpp"
)

// oracleFixtures mirrors ivy2cpp/oracle_test.go's catalog of the 14
// fixtures under `ivy2cpp/test_vec/oracle/`. Listed in the same order
// for grep-ability.
var oracleFixtures = []string{
	"empty.ivy",
	"basic_assign.ivy",
	"forall_assign.ivy",
	"bv_arithmetic.ivy",
	"enum_dispatch.ivy",
	"range_bounds.ivy",
	"destructor_record.ivy",
	"variant_simple.ivy",
	"variant_recursive.ivy",
	"hash_thunk_assign.ivy",
	"native_block.ivy",
	"callback_thunk.ivy",
	"progress_property.ivy",
	"isolate_two_parts.ivy",
}

// oraclePipelineStage is the point at which a per-fixture pipeline
// failed. Empty string indicates success through the full pipeline.
type oraclePipelineStage string

const (
	oracleStageGenerate oraclePipelineStage = "generate"
	oracleStageBuild    oraclePipelineStage = "build"
	oracleStageRun      oraclePipelineStage = "run"
)

// oracleOutcome captures the per-fixture comparison result.
type oracleOutcome struct {
	Fixture   string
	GoStage   oraclePipelineStage // "" if Go pipeline reached run+success
	GoErr     string
	GoStdout  string
	CppStage  oraclePipelineStage // "" if C++ pipeline reached run+success
	CppErr    string
	CppStdout string
	Verdict   string // "parity" | "divergent" | "blocked-go-<stage>" | "blocked-cpp-<stage>" | "blocked-both"
}

func TestOracleParity(t *testing.T) {
	if os.Getenv("ORACLE") == "" {
		t.Skip("set ORACLE=1 to compare ivy2cpp-emitted vs ivy2go-emitted binary stdout")
	}
	if !SlowGoTest {
		t.Skip("set SLOW_GO_TEST=1 to enable build-and-run oracle parity")
	}
	outcomes := make([]oracleOutcome, 0, len(oracleFixtures))
	for _, fixture := range oracleFixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			outcome := runOracleParityFixture(t, fixture)
			outcomes = append(outcomes, outcome)
			switch outcome.Verdict {
			case "parity":
				// silent on success — keeps -v output digestible.
			case "divergent":
				t.Errorf("DIVERGENT stdout:\n  go (%d bytes): %q\n  cpp (%d bytes): %q",
					len(outcome.GoStdout), trim(outcome.GoStdout, 200),
					len(outcome.CppStdout), trim(outcome.CppStdout, 200))
			default:
				t.Logf("BLOCKED: %s\n  go-err: %s\n  cpp-err: %s",
					outcome.Verdict, trim(outcome.GoErr, 400), trim(outcome.CppErr, 400))
			}
		})
	}
	t.Cleanup(func() { logOracleParitySummary(t, outcomes) })
}

func runOracleParityFixture(t *testing.T, fixture string) oracleOutcome {
	t.Helper()
	srcAbs := mustAbs(t, oracleParityFixturePath(fixture))
	out := oracleOutcome{Fixture: fixture}

	// --- Go pipeline ---------------------------------------------------
	out.GoStage, out.GoErr, out.GoStdout = runOracleGoPipeline(t, srcAbs, fixture)

	// --- C++ pipeline --------------------------------------------------
	out.CppStage, out.CppErr, out.CppStdout = runOracleCppPipeline(t, srcAbs, fixture)

	// --- verdict -------------------------------------------------------
	switch {
	case out.GoStage != "" && out.CppStage != "":
		out.Verdict = "blocked-both"
	case out.GoStage != "":
		out.Verdict = "blocked-go-" + string(out.GoStage)
	case out.CppStage != "":
		out.Verdict = "blocked-cpp-" + string(out.CppStage)
	default:
		// Both pipelines reached a clean run. Compare stdouts.
		if normalizeOracleStdout(out.GoStdout) == normalizeOracleStdout(out.CppStdout) {
			out.Verdict = "parity"
		} else {
			out.Verdict = "divergent"
		}
	}
	return out
}

// runOracleGoPipeline emits target=test Go for fixture, builds it,
// runs it with `--iters=0 --seed=1`, and returns whatever stdout the
// binary produced (or the stage that failed).
func runOracleGoPipeline(t *testing.T, srcAbs, fixture string) (oraclePipelineStage, string, string) {
	t.Helper()
	dir := playpenDir(t)
	out, err := CompileAndGenerate(srcAbs, nil, Config{
		Target:      "test",
		PackageName: "main",
	})
	if err != nil {
		return oracleStageGenerate, fmt.Sprintf("ivy2go.CompileAndGenerate: %v", err), ""
	}
	if err := WriteOutput(out, dir); err != nil {
		return oracleStageGenerate, fmt.Sprintf("ivy2go.WriteOutput: %v", err), ""
	}
	pkgDir := outputDirectory(dir, out.BaseName)
	binName := "oracle_bin"
	build := exec.Command("go", "build", "-o", binName, ".")
	build.Dir = pkgDir
	if buildOut, err := build.CombinedOutput(); err != nil {
		return oracleStageBuild, fmt.Sprintf("go build: %v\n%s", err, string(buildOut)), ""
	}
	run := exec.Command(filepath.Join(pkgDir, binName), "--iters=0", "--seed=1")
	runOut, err := run.CombinedOutput()
	if err != nil {
		return oracleStageRun, fmt.Sprintf("run: %v\n%s", err, string(runOut)), string(runOut)
	}
	return "", "", string(runOut)
}

// runOracleCppPipeline emits target=test C++ for fixture, builds it
// with g++, runs it with `iters=0 seed=1`, and returns whatever stdout
// the binary produced (or the stage that failed).
func runOracleCppPipeline(t *testing.T, srcAbs, fixture string) (oraclePipelineStage, string, string) {
	t.Helper()
	dir := playpenDir(t)
	className := strings.TrimSuffix(fixture, ".ivy")
	className = strings.ReplaceAll(className, "-", "_")
	batch, err := ivy2cpp.CompileAndGenerateAll(srcAbs, map[string]string{
		"target":    "test",
		"classname": className,
		"outdir":    dir,
	}, ivy2cpp.Config{})
	if err != nil {
		return oracleStageGenerate, fmt.Sprintf("ivy2cpp.CompileAndGenerateAll: %v", err), ""
	}
	if err := ivy2cpp.WriteBatchOutput(batch, dir); err != nil {
		return oracleStageGenerate, fmt.Sprintf("ivy2cpp.WriteBatchOutput: %v", err), ""
	}
	if len(batch.Outputs) == 0 {
		return oracleStageGenerate, "ivy2cpp produced no outputs", ""
	}
	// pick the first (single-isolate fixtures expose one Output)
	out := batch.Outputs[0]
	binPath, err := ivy2cpp.BuildOutput(out, dir)
	if err != nil {
		return oracleStageBuild, fmt.Sprintf("ivy2cpp.BuildOutput: %v", err), ""
	}
	run := exec.Command(binPath, "iters=0", "seed=1")
	runOut, err := run.CombinedOutput()
	if err != nil {
		return oracleStageRun, fmt.Sprintf("run: %v\n%s", err, string(runOut)), string(runOut)
	}
	return "", "", string(runOut)
}

// normalizeOracleStdout strips trailing whitespace and collapses CRLF
// to LF so byte-for-byte comparison isn't tripped by platform quirks.
func normalizeOracleStdout(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimRight(s, "\n \t")
}

func oracleParityFixturePath(fixture string) string {
	return filepath.Join("..", "ivy2cpp", "test_vec", "oracle", fixture)
}

func mustAbs(t *testing.T, p string) string {
	t.Helper()
	abs, err := filepath.Abs(p)
	if err != nil {
		t.Fatalf("filepath.Abs(%q): %v", p, err)
	}
	return abs
}

func trim(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + fmt.Sprintf("... [%d more bytes]", len(s)-n)
}

// logOracleParitySummary prints the N parity / M blocked / Z divergent
// roll-up the user-facing acceptance gate watches.
func logOracleParitySummary(t *testing.T, outcomes []oracleOutcome) {
	t.Helper()
	counts := map[string]int{}
	var byVerdict []string
	for _, o := range outcomes {
		counts[o.Verdict]++
	}
	for k := range counts {
		byVerdict = append(byVerdict, k)
	}
	sort.Strings(byVerdict)
	parts := []string{fmt.Sprintf("%d total", len(outcomes))}
	for _, v := range byVerdict {
		parts = append(parts, fmt.Sprintf("%d %s", counts[v], v))
	}
	t.Logf("oracle parity summary: %s", strings.Join(parts, " / "))

	// Per-fixture line — useful when running with -v.
	for _, o := range outcomes {
		t.Logf("  %-28s %s", o.Fixture, o.Verdict)
	}
}
