package ivy2cpp

import (
	"bytes"
	cryrand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// returns r >= 0
func cryptoRandNonNegInt64() (r int64) {
	b := make([]byte, 8)
	_, err := cryrand.Read(b)
	if err != nil {
		panic(err)
	}
	r = int64(binary.LittleEndian.Uint64(b))
	if r < 0 {
		if r == math.MinInt64 {
			return 0
		}
		r = -r
	}
	return r
}

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

const (
	oracleStatusPass         = "PASS"
	oracleStatusExpectedFail = "EXPECTED_FAIL"
	oracleStatusSkip         = "SKIP"
)

type oracleCase struct {
	Name             string
	Target           string
	IncludeClassName bool
	Build            bool
}

func oracleCasesForTarget(target string) []oracleCase {
	cases := []oracleCase{
		{Name: "no_class_no_build", Target: target},
		{Name: "with_class_no_build", Target: target, IncludeClassName: true},
	}
	// Python target=impl build=true tries to link an implementation with no
	// generated main and fails in the source-of-truth tool. The build flag is
	// meaningful for generated executables such as target=test.
	if target == "test" || target == "repl" {
		cases = append(cases,
			oracleCase{Name: "no_class_build", Target: target, Build: true},
			oracleCase{Name: "with_class_build", Target: target, IncludeClassName: true, Build: true},
		)
	}
	return cases
}

func (c oracleCase) className(fixture string) string {
	if !c.IncludeClassName {
		return ""
	}
	return "oracle_" + varName(oracleClassName(fixture))
}

func TestCompareCPPTokensStripsCommentsAndWhitespace(t *testing.T) {
	left := `
int main() {
  std::string s = "/* not a comment */";
  return 1 + 2;
}
`
	right := `int main(){ // comment
std::string s="/* not a comment */"; /* block */
return 1+2;}`
	diff := CompareCPPTokens("left", left, "right", right)
	if !diff.Equal {
		t.Fatalf("equivalent C++ streams differed: %s", diff.Error())
	}
}

func TestCompareCPPTokensPreservesStringLiterals(t *testing.T) {
	left := `const char *s = "a b";`
	right := `const char *s = "a  b";`
	diff := CompareCPPTokens("left", left, "right", right)
	if diff.Equal {
		t.Fatal("string-literal whitespace must remain significant")
	}
	if diff.LeftToken != `"a b"` || diff.RightToken != `"a  b"` {
		t.Fatalf("wrong divergent tokens: %#v", diff)
	}
}

func TestCompareCPPTokensReportsContext(t *testing.T) {
	diff := CompareCPPTokens("go", "int x = 1;\nint y = 2;", "python", "int x = 1;\nint y = 3;")
	if diff.Equal {
		t.Fatal("expected divergence")
	}
	for _, want := range []string{"token", "go", "python", "2", "3"} {
		if !strings.Contains(diff.Error(), want) {
			t.Fatalf("diff report missing %q: %s", want, diff.Error())
		}
	}
	if diff.LeftContext == "" || diff.RightContext == "" {
		t.Fatalf("missing context in diff: %#v", diff)
	}
}

func TestOracleFixturesExist(t *testing.T) {
	for _, fixture := range oracleFixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			if _, err := os.Stat(oracleFixturePath(fixture)); err != nil {
				t.Fatalf("missing oracle fixture %s: %v", fixture, err)
			}
			if _, err := os.Stat(oracleTesterArgsPath(fixture)); err != nil {
				t.Fatalf("missing generated tester transcript args for %s: %v", fixture, err)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(oracleDir(), "compare_cpp.go")); err != nil {
		t.Fatalf("missing standalone oracle comparator: %v", err)
	}
}

func TestOracleStatusFileUpToDate(t *testing.T) {
	statuses := readOracleStatuses(t)
	want := append([]string(nil), oracleFixtures...)
	got := make([]string, 0, len(statuses))
	for fixture, status := range statuses {
		switch status {
		case oracleStatusPass, oracleStatusExpectedFail, oracleStatusSkip:
		default:
			t.Fatalf("fixture %s has invalid oracle status %q", fixture, status)
		}
		if status == oracleStatusExpectedFail {
			t.Fatalf("fixture %s is still marked %s after TODO3 oracle promotion", fixture, status)
		}
		got = append(got, fixture)
	}
	sort.Strings(want)
	sort.Strings(got)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("STATUS.md fixtures differ from oracle catalog\nwant:\n%s\n\ngot:\n%s",
			strings.Join(want, "\n"), strings.Join(got, "\n"))
	}
}

func TestOracleSingle(t *testing.T) {
	if !oracleTestEnabled() {
		t.Skip("set ORACLE_TEST=1 to compare Go ivy2cpp output against Python ivy_to_cpp")
	}
	statuses := readOracleStatuses(t)
	var outcomes []oracleOutcome
	for _, fixture := range oracleFixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			status := statuses[fixture]
			if status == oracleStatusSkip {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "skip"})
				t.Skip("fixture marked SKIP in STATUS.md")
			}
			for _, tc := range oracleCasesForTarget("impl") {
				t.Run(tc.Name, func(t *testing.T) {
					err := compareOracleFixture(t, fixture, tc)
					recordOracleExpectation(t, fixture, status, err, &outcomes)
				})
			}
		})
	}
	t.Cleanup(func() {
		t.Logf("oracle parity summary: %s", formatOracleOutcomes(outcomes))
	})
}

func TestOracleSingleTargetTest(t *testing.T) {
	if !oracleTestEnabled() {
		t.Skip("set ORACLE_TEST=1 to compare Go ivy2cpp test output against Python ivy_to_cpp")
	}
	if !oracleTargetTestEnabled() {
		t.Skip("set ORACLE_TEST_TARGET_TEST=1 to run the broader target=test token oracle")
	}
	statuses := readOracleStatuses(t)
	var outcomes []oracleOutcome
	for _, fixture := range oracleFixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			status := statuses[fixture]
			if status == oracleStatusSkip {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "skip"})
				t.Skip("fixture marked SKIP in STATUS.md")
			}
			if _, err := readOracleTesterArgs(fixture); errors.Is(err, errOracleTesterSkipped) {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "tester skip"})
				t.Skip(err.Error())
			}
			for _, tc := range oracleCasesForTarget("test") {
				t.Run(tc.Name, func(t *testing.T) {
					err := compareOracleFixture(t, fixture, tc)
					recordOracleExpectation(t, fixture, status, err, &outcomes)
				})
			}
		})
	}
	t.Cleanup(func() {
		t.Logf("oracle target=test parity summary: %s", formatOracleOutcomes(outcomes))
	})
}

func TestOracleHermesRMWO3TargetTestDiffWE(t *testing.T) {
	//if !oracleTestEnabled() {
	//	t.Skip("set ORACLE_TEST=1 to compare Hermes Go ivy2cpp output against Python ivy_to_cpp")
	//}
	fixture := hermesRMWO3OracleFixturePath()
	if _, err := os.Stat(fixture); err != nil {
		t.Skipf("Hermes oracle fixture not available: %v", err)
	}

	// allow simultaneous runs.
	root, err := os.MkdirTemp("", "tmp.oracle.out.dir.")
	panicOn(err)
	var bad bool
	defer func() {
		if !bad {
			os.RemoveAll(root) // comment to manually inspect output.
		}
	}()
	for _, tc := range oracleCasesForTarget("test") {
		t.Run(tc.Name, func(t *testing.T) {
			goDir, pyDir := hermesOracleDirs(root, tc)
			panicOn(os.MkdirAll(goDir, 0755))
			panicOn(os.MkdirAll(pyDir, 0755))
			if err := compareOracleFixtureAtPaths(t, fixture, tc, goDir, pyDir, true); err != nil {
				bad = true
				t.Fatalf("Hermes oracle %s failed: %v", tc.Name, err)
			}
		})
	}
}

func hermesOracleDirs(root string, tc oracleCase) (string, string) {
	if tc.Name == "no_class_no_build" {
		return filepath.Join(root, "go.out"), filepath.Join(root, "py.out")
	}
	return filepath.Join(root, tc.Name, "go.out"), filepath.Join(root, tc.Name, "py.out")
}

func TestOracleCompileGo(t *testing.T) {
	if !SlowCppTest {
		t.Skip("set SLOW_CPP_TEST=1 to compile oracle Go output")
	}
	statuses := readOracleStatuses(t)
	var outcomes []oracleOutcome
	for _, fixture := range oracleFixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			status := statuses[fixture]
			if status == oracleStatusSkip {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "skip"})
				t.Skip("fixture marked SKIP in STATUS.md")
			}
			err := compileGoOracleFixture(t, fixture, "impl")
			recordOracleExpectation(t, fixture, status, err, &outcomes)
		})
	}
	t.Cleanup(func() {
		t.Logf("oracle Go compile summary: %s", formatOracleOutcomes(outcomes))
	})
}

func TestOracleCompileGoTargetTest(t *testing.T) {
	if !SlowCppTest {
		t.Skip("set SLOW_CPP_TEST=1 to compile oracle Go test output")
	}
	statuses := readOracleStatuses(t)
	var outcomes []oracleOutcome
	for _, fixture := range oracleFixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			status := statuses[fixture]
			if status == oracleStatusSkip {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "skip"})
				t.Skip("fixture marked SKIP in STATUS.md")
			}
			err := compileGoOracleFixture(t, fixture, "test")
			recordOracleExpectation(t, fixture, status, err, &outcomes)
		})
	}
	t.Cleanup(func() {
		t.Logf("oracle Go target=test compile summary: %s", formatOracleOutcomes(outcomes))
	})
}

func TestOracleCompilePython(t *testing.T) {
	if !SlowCppTest {
		t.Skip("set SLOW_CPP_TEST=1 to compile oracle Python output")
	}
	if !oracleTestEnabled() {
		t.Skip("set ORACLE_TEST=1 to compile Python ivy_to_cpp output")
	}
	statuses := readOracleStatuses(t)
	var outcomes []oracleOutcome
	for _, fixture := range oracleFixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			status := statuses[fixture]
			if status == oracleStatusSkip {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "skip"})
				t.Skip("fixture marked SKIP in STATUS.md")
			}
			err := compilePythonOracleFixture(t, fixture, "impl")
			recordOracleExpectation(t, fixture, status, err, &outcomes)
		})
	}
	t.Cleanup(func() {
		t.Logf("oracle Python compile summary: %s", formatOracleOutcomes(outcomes))
	})
}

func TestOracleCompilePythonTargetTest(t *testing.T) {
	if !SlowCppTest {
		t.Skip("set SLOW_CPP_TEST=1 to compile oracle Python test output")
	}
	if !oracleTestEnabled() {
		t.Skip("set ORACLE_TEST=1 to compile Python ivy_to_cpp test output")
	}
	statuses := readOracleStatuses(t)
	var outcomes []oracleOutcome
	for _, fixture := range oracleFixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			status := statuses[fixture]
			if status == oracleStatusSkip {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "skip"})
				t.Skip("fixture marked SKIP in STATUS.md")
			}
			if _, err := readOracleTesterArgs(fixture); errors.Is(err, errOracleTesterSkipped) {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "tester skip"})
				t.Skip(err.Error())
			}
			err := compilePythonOracleFixture(t, fixture, "test")
			recordOracleExpectation(t, fixture, status, err, &outcomes)
		})
	}
	t.Cleanup(func() {
		t.Logf("oracle Python target=test compile summary: %s", formatOracleOutcomes(outcomes))
	})
}

func TestOracleSemanticEquivalence(t *testing.T) {
	if !SlowCppTest {
		t.Skip("set SLOW_CPP_TEST=1 to run oracle transcript equivalence")
	}
	if !oracleTestEnabled() {
		t.Skip("set ORACLE_TEST=1 to run Python ivy_to_cpp transcript equivalence")
	}
	statuses := readOracleStatuses(t)
	var outcomes []oracleOutcome
	for _, fixture := range oracleFixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			status := statuses[fixture]
			if status == oracleStatusSkip {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "skip"})
				t.Skip("fixture marked SKIP in STATUS.md")
			}
			err := runOracleTranscript(t, fixture)
			if err == errOracleTranscriptMissing {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "no transcript"})
				t.Skip("no .in transcript for this fixture")
			}
			recordOracleExpectation(t, fixture, status, err, &outcomes)
		})
	}
	t.Cleanup(func() {
		t.Logf("oracle semantic summary: %s", formatOracleOutcomes(outcomes))
	})
}

func TestOracleSemanticEquivalenceGeneratedTester(t *testing.T) {
	if !SlowCppTest {
		t.Skip("set SLOW_CPP_TEST=1 to run generated tester transcript equivalence")
	}
	if !oracleTestEnabled() {
		t.Skip("set ORACLE_TEST=1 to run Python ivy_to_cpp tester equivalence")
	}
	statuses := readOracleStatuses(t)
	var outcomes []oracleOutcome
	for _, fixture := range oracleFixtures {
		fixture := fixture
		t.Run(fixture, func(t *testing.T) {
			status := statuses[fixture]
			if status == oracleStatusSkip {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "skip"})
				t.Skip("fixture marked SKIP in STATUS.md")
			}
			err := runOracleTesterTranscript(t, fixture)
			if err == errOracleTesterArgsMissing {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "no tester args"})
				t.Skip("no .test.args transcript arguments for this fixture")
			}
			if errors.Is(err, errOracleTesterSkipped) {
				outcomes = append(outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "tester skip"})
				t.Skip(err.Error())
			}
			recordOracleExpectation(t, fixture, status, err, &outcomes)
		})
	}
	t.Cleanup(func() {
		t.Logf("oracle generated tester semantic summary: %s", formatOracleOutcomes(outcomes))
	})
}

type oracleOutcome struct {
	Fixture string
	Status  string
	Outcome string
}

func oracleDir() string {
	return filepath.Join("test_vec", "oracle")
}

func oracleFixturePath(fixture string) string {
	return filepath.Join(oracleDir(), fixture)
}

func oracleTesterArgsPath(fixture string) string {
	return strings.TrimSuffix(oracleFixturePath(fixture), ".ivy") + ".test.args"
}

// this should always return true now.
func oracleTestEnabled() bool {
	return true
	// was weak sauce:
	//return os.Getenv("ORACLE_TEST") != ""
}

func oracleTargetTestEnabled() bool {
	return os.Getenv("ORACLE_TEST_TARGET_TEST") != ""
}

func hermesRMWO3OracleFixturePath() string {
	if path := strings.TrimSpace(os.Getenv("HERMES_RMW_O3_TESTING_IVY")); path != "" {
		return path
	}
	return filepath.Join(os.Getenv("HOME"), "ivy", "ivy-lang-examples", "jea", "hermes_rmw_o3_testing.ivy")
}

func assertDiffWEEqual(t *testing.T, left, right string) {
	t.Helper()
	if err := diffWEEqual(left, right); err != nil {
		t.Fatal(err)
	}
	//vv("left='%v' right='%v' diff is: ''", left, right)
}

func diffWEEqual(left, right string) error {
	same, err := bytesEqualsIgnoreWhitespace(left, right)
	if err != nil {
		return err
	}
	if same {
		return nil
	}
	cmd := exec.Command("diff", "-w", "-B", "-b", left, right)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		out := buf.String()
		if len(out) > 12000 {
			out = out[:12000] + "\n... diff truncated ..."
		}
		return fmt.Errorf("diff -w -B -b %s %s failed: %w\n%s", left, right, err, out)
	}
	//vv("left='%v' right='%v' diff is: '%v'", left, right, buf.String())
	return nil
}

// bytesEqualsIgnoreWhitespace returns true if both slices match after removing all whitespace.
func bytesEqualsIgnoreWhitespace(pathA, pathB string) (bool, error) {
	a, err := os.ReadFile(pathA)
	if err != nil {
		return false, err
	}
	b, err := os.ReadFile(pathB)
	if err != nil {
		return false, err
	}
	// bytes.Fields splits on all whitespace (\t, \n, \r, ' ') and strips empty elements
	aFields := bytes.Fields(a)
	bFields := bytes.Fields(b)

	return bytes.Equal(bytes.Join(aFields, nil), bytes.Join(bFields, nil)), nil
}

func readOracleStatuses(t *testing.T) map[string]string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(oracleDir(), "STATUS.md"))
	if err != nil {
		t.Fatalf("read oracle STATUS.md: %v", err)
	}
	statuses := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "|") {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		fixture := strings.TrimSpace(parts[1])
		status := strings.TrimSpace(parts[2])
		if fixture == "Fixture" || strings.HasPrefix(fixture, "---") {
			continue
		}
		statuses[fixture] = status
	}
	return statuses
}

func compareOracleFixture(t *testing.T, fixture string, tc oracleCase) error {
	t.Helper()
	goDir := t.TempDir()
	pyDir := t.TempDir()
	return compareOracleFixtureAtPaths(t, oracleFixturePath(fixture), tc, goDir, pyDir, false)
}

func compareOracleFixtureAtPaths(t *testing.T, fixture string, tc oracleCase, goDir, pyDir string, diffWE bool) error {
	t.Helper()
	if _, err := generateGoOracleFixtureAtPath(fixture, goDir, tc); err != nil {
		return fmt.Errorf("Go generate: %w", err)
	}
	if err := runPythonIvyToCPP(fixture, pyDir, tc); err != nil {
		return fmt.Errorf("Python generate: %w", err)
	}
	if diffWE {
		return compareGeneratedCPPFilesDiffWE(goDir, pyDir)
	}
	goFiles, err := readGeneratedCPPFiles(goDir)
	if err != nil {
		return fmt.Errorf("read Go output: %w", err)
	}
	pyFiles, err := readGeneratedCPPFiles(pyDir)
	if err != nil {
		return fmt.Errorf("read Python output: %w", err)
	}
	return compareGeneratedCPPFiles(goFiles, pyFiles)
}

func generateGoOracleFixture(fixture, outDir string, tc oracleCase) (*BatchOutput, error) {
	absFixture, err := filepath.Abs(oracleFixturePath(fixture))
	if err != nil {
		return nil, err
	}
	return generateGoOracleFixtureAtPath(absFixture, outDir, tc)
}

func generateGoOracleFixtureAtPath(absFixture, outDir string, tc oracleCase) (*BatchOutput, error) {
	var err error
	absFixture, err = filepath.Abs(absFixture)
	if err != nil {
		return nil, err
	}
	params := map[string]string{
		"target": tc.Target,
		"outdir": outDir,
	}
	if className := tc.className(absFixture); className != "" {
		params["classname"] = className
	}
	if tc.Build {
		params["build"] = "true"
	}
	batch, err := CompileAndGenerateAll(absFixture, params, Config{})
	if err != nil {
		return nil, err
	}
	if err := WriteBatchOutput(batch, outDir); err != nil {
		return nil, err
	}
	if tc.Build {
		for _, out := range batch.Outputs {
			if _, err := BuildOutput(out, outDir); err != nil {
				return nil, err
			}
		}
	}
	return batch, nil
}

func runPythonIvyToCPP(fixture, outDir string, tc oracleCase) error {
	tool, err := pythonIvyToCPPPath()
	if err != nil {
		return err
	}
	absFixture, err := filepath.Abs(fixture)
	if err != nil {
		return err
	}
	cmdFixture := filepath.Base(absFixture)
	var args []string
	absOutDir, err := filepath.Abs(outDir)
	if err != nil {
		return err
	}
	cmdDir := absOutDir
	raw, err := os.ReadFile(absFixture)
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cmdDir, cmdFixture), raw, 0o644); err != nil {
		return err
	}
	if tc.Build {
		args = []string{"target=" + tc.Target, "build=true"}
	} else {
		args = []string{"target=" + tc.Target, "outdir=" + absOutDir}
	}
	if className := tc.className(absFixture); className != "" {
		args = append(args, "classname="+className)
	}
	args = append(args, cmdFixture)
	//vv("debug runPythonIvyToCPP() calling: tool: '%v' args='%#v'", tool, args)
	cmd := exec.Command(tool, args...)
	cmd.Dir = cmdDir
	cmd.Env = append(os.Environ(), "PYTHONHASHSEED=0")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w\n%s", tool, strings.Join(args, " "), err, buf.String())
	}
	return nil
}

func pythonIvyToCPPPath() (string, error) {
	for _, env := range []string{"IVY_TO_CPP", "PY_IVY_TO_CPP"} {
		if path := strings.TrimSpace(os.Getenv(env)); path != "" {
			if _, err := os.Stat(path); err != nil {
				return "", fmt.Errorf("%s=%s is not usable: %w", env, path, err)
			}
			return path, nil
		}
	}
	for _, candidate := range []string{
		"/Users/jaten/ivy/pyivy/goivy-venv/bin/ivy_to_cpp",
		filepath.Join(os.Getenv("HOME"), "ivy", "pyivy", "goivy-venv", "bin", "ivy_to_cpp"),
	} {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			//vv("using candidate = '%v'", candidate)
			return candidate, nil
		}
	}
	if path, err := exec.LookPath("ivy_to_cpp"); err == nil {
		//vv("using path = '%v'", path)
		return path, nil
	}
	return "", fmt.Errorf("Python ivy_to_cpp not found; set IVY_TO_CPP")
}

func readGeneratedCPPFiles(dir string) (map[string]string, error) {
	files := map[string]string{}
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".h" && ext != ".cpp" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(rel)] = string(raw)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func compareGeneratedCPPFiles(goFiles, pyFiles map[string]string) error {
	goNames := sortedMapKeys(goFiles)
	pyNames := sortedMapKeys(pyFiles)
	if strings.Join(goNames, "\n") != strings.Join(pyNames, "\n") {
		return fmt.Errorf("generated file set differs\nGo:\n%s\n\nPython:\n%s", strings.Join(goNames, "\n"), strings.Join(pyNames, "\n"))
	}
	for _, name := range goNames {
		diff := CompareCPPTokens("Go "+name, goFiles[name], "Python "+name, pyFiles[name])
		if !diff.Equal {
			return fmt.Errorf("%s: %s", name, diff.Error())
		}
	}
	return nil
}

func compareGeneratedCPPFilesDiffWE(goDir, pyDir string) error {
	goFiles, err := readGeneratedCPPFiles(goDir)
	if err != nil {
		return fmt.Errorf("read Go output: %w", err)
	}
	pyFiles, err := readGeneratedCPPFiles(pyDir)
	if err != nil {
		return fmt.Errorf("read Python output: %w", err)
	}
	goNames := sortedMapKeys(goFiles)
	pyNames := sortedMapKeys(pyFiles)
	if strings.Join(goNames, "\n") != strings.Join(pyNames, "\n") {
		return fmt.Errorf("generated file set differs\nGo:\n%s\n\nPython:\n%s", strings.Join(goNames, "\n"), strings.Join(pyNames, "\n"))
	}
	for _, name := range goNames {
		if err := diffWEEqual(filepath.Join(pyDir, name), filepath.Join(goDir, name)); err != nil {
			return err
		}
	}
	return nil
}

func sortedMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func compileGoOracleFixture(t *testing.T, fixture, target string) error {
	t.Helper()
	batch, err := generateGoOracleFixture(fixture, t.TempDir(), oracleCase{Target: target})
	if err != nil {
		return err
	}
	for _, out := range batch.Outputs {
		compiled := *out
		compiled.EmitMain = false
		if _, err := BuildOutput(&compiled, t.TempDir()); err != nil {
			if isMissingZ3ToolchainError(err) {
				t.Skip(err.Error())
			}
			return err
		}
	}
	return nil
}

func compilePythonOracleFixture(t *testing.T, fixture, target string) error {
	t.Helper()
	dir := t.TempDir()
	if err := runPythonIvyToCPP(oracleFixturePath(fixture), dir, oracleCase{Target: target}); err != nil {
		return err
	}
	files, err := cppSourcePaths(dir)
	if err != nil {
		return err
	}
	for _, cppPath := range files {
		if err := compileCPPTranslationUnit(cppPath, filepath.Join(dir, filepath.Base(cppPath)+".o")); err != nil {
			if isMissingZ3ToolchainError(err) {
				t.Skip(err.Error())
			}
			return err
		}
	}
	return nil
}

func cppSourcePaths(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".cpp" {
			return nil
		}
		files = append(files, path)
		return nil
	})
	sort.Strings(files)
	return files, err
}

func compileCPPTranslationUnit(cppPath, objPath string) error {
	compiler, err := cxxCompilerFor("default")
	if err != nil {
		return err
	}
	args := []string{"-std=c++11", "-Wno-parentheses-equality", "-g"}
	includeArgs, err := supportIncludeArgs()
	if err != nil {
		return err
	}
	args = append(args, includeArgs...)
	for _, dir := range pythonIvyIncludeDirs() {
		args = append(args, "-I", dir)
	}
	args = append(args, "-c", cppPath, "-o", objPath)
	cmd := exec.Command(compiler, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("compile %s: %w\n%s", cppPath, err, buf.String())
	}
	return nil
}

func pythonIvyIncludeDirs() []string {
	var dirs []string
	roots := []string{
		filepath.Join(os.Getenv("HOME"), "ivy", "pyivy", "ivy", "ivy", "include"),
		"/Users/jaten/ivy/pyivy/ivy/ivy/include",
	}
	seen := map[string]bool{}
	for _, root := range roots {
		if root == "" || seen[root] {
			continue
		}
		seen[root] = true
		if _, err := os.Stat(root); err == nil {
			dirs = append(dirs, root)
		}
		for _, version := range []string{"1.8", "1.7", "1.6"} {
			dir := filepath.Join(root, version)
			if seen[dir] {
				continue
			}
			seen[dir] = true
			if _, err := os.Stat(dir); err == nil {
				dirs = append(dirs, dir)
			}
		}
	}
	return dirs
}

var errOracleTranscriptMissing = fmt.Errorf("oracle transcript missing")
var errOracleTesterArgsMissing = fmt.Errorf("oracle tester args missing")
var errOracleTesterSkipped = fmt.Errorf("oracle tester skipped")

func runOracleTranscript(t *testing.T, fixture string) error {
	t.Helper()
	transcript := strings.TrimSuffix(oracleFixturePath(fixture), ".ivy") + ".in"
	input, err := os.ReadFile(transcript)
	if err != nil {
		if os.IsNotExist(err) {
			return errOracleTranscriptMissing
		}
		return err
	}
	goExe, err := buildGoOracleExecutable(t, fixture)
	if err != nil {
		return fmt.Errorf("Go executable: %w", err)
	}
	pyExe, err := buildPythonOracleExecutable(t, fixture)
	if err != nil {
		return fmt.Errorf("Python executable: %w", err)
	}
	goOut, err := runWithInput(goExe, input)
	if err != nil {
		return fmt.Errorf("Go run: %w", err)
	}
	pyOut, err := runWithInput(pyExe, input)
	if err != nil {
		return fmt.Errorf("Python run: %w", err)
	}
	if string(goOut) != string(pyOut) {
		return fmt.Errorf("transcript stdout differs\nGo:\n%s\nPython:\n%s", goOut, pyOut)
	}
	return nil
}

func buildGoOracleExecutable(t *testing.T, fixture string) (string, error) {
	t.Helper()
	return buildGoOracleExecutableForTarget(t, fixture, "repl")
}

func buildPythonOracleExecutable(t *testing.T, fixture string) (string, error) {
	t.Helper()
	return buildPythonOracleExecutableForTarget(t, fixture, "repl")
}

func runOracleTesterTranscript(t *testing.T, fixture string) error {
	t.Helper()
	args, err := readOracleTesterArgs(fixture)
	if err != nil {
		return err
	}
	goExe, err := buildGoOracleExecutableForTarget(t, fixture, "test")
	if err != nil {
		return fmt.Errorf("Go tester executable: %w", err)
	}
	pyExe, err := buildPythonOracleExecutableForTarget(t, fixture, "test")
	if err != nil {
		return fmt.Errorf("Python tester executable: %w", err)
	}
	goOut, err := runExecutable(goExe, args, nil)
	if err != nil {
		return fmt.Errorf("Go tester run: %w", err)
	}
	pyOut, err := runExecutable(pyExe, args, nil)
	if err != nil {
		return fmt.Errorf("Python tester run: %w", err)
	}
	if string(goOut) != string(pyOut) {
		return fmt.Errorf("tester stdout differs\nargs: %s\nGo:\n%s\nPython:\n%s", strings.Join(args, " "), goOut, pyOut)
	}
	return nil
}

func readOracleTesterArgs(fixture string) ([]string, error) {
	path := oracleTesterArgsPath(fixture)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errOracleTesterArgsMissing
		}
		return nil, err
	}
	var args []string
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 && fields[0] == "SKIP" {
			return nil, fmt.Errorf("%w: %s", errOracleTesterSkipped, strings.Join(fields[1:], " "))
		}
		args = append(args, fields...)
	}
	return args, nil
}

func buildGoOracleExecutableForTarget(t *testing.T, fixture, target string) (string, error) {
	t.Helper()
	batch, err := generateGoOracleFixture(fixture, t.TempDir(), oracleCase{Target: target})
	if err != nil {
		return "", err
	}
	if len(batch.Outputs) != 1 {
		return "", fmt.Errorf("semantic transcript expects one Go output, got %d", len(batch.Outputs))
	}
	return BuildOutput(batch.Outputs[0], t.TempDir())
}

func buildPythonOracleExecutableForTarget(t *testing.T, fixture, target string) (string, error) {
	t.Helper()
	dir := t.TempDir()
	if err := runPythonIvyToCPP(oracleFixturePath(fixture), dir, oracleCase{Target: target}); err != nil {
		return "", err
	}
	cpps, err := cppSourcePaths(dir)
	if err != nil {
		return "", err
	}
	if len(cpps) != 1 {
		return "", fmt.Errorf("semantic transcript expects one Python .cpp output, got %d", len(cpps))
	}
	exe := filepath.Join(dir, strings.TrimSuffix(filepath.Base(cpps[0]), ".cpp"))
	if err := linkCPPExecutable(cpps[0], exe); err != nil {
		return "", err
	}
	return exe, nil
}

func linkCPPExecutable(cppPath, exePath string) error {
	compiler, err := cxxCompilerFor("default")
	if err != nil {
		return err
	}
	args := []string{"-std=c++11", "-Wno-parentheses-equality", "-g"}
	includeArgs, err := supportIncludeArgs()
	if err != nil {
		return err
	}
	args = append(args, includeArgs...)
	var linkArgs []string
	raw, err := os.ReadFile(cppPath)
	if err != nil {
		return err
	}
	text := string(raw)
	if strings.Contains(text, "z3++.h") ||
		strings.Contains(text, "z3::") ||
		strings.Contains(text, "ivy_z3_gen.hpp") ||
		strings.Contains(text, "ivy_z3_helpers.hpp") ||
		strings.Contains(text, "ivy_go_z3.hpp") {
		z3IncludeArgs, z3LinkArgs, err := z3BuildArgs()
		if err != nil {
			return err
		}
		args = append(args, z3IncludeArgs...)
		linkArgs = z3LinkArgs
	}
	for _, dir := range pythonIvyIncludeDirs() {
		args = append(args, "-I", dir)
	}
	args = append(args, cppPath, "-o", exePath)
	args = append(args, linkArgs...)
	args = append(args, "-pthread")
	cmd := exec.Command(compiler, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("link %s: %w\n%s", cppPath, err, buf.String())
	}
	return nil
}

func runWithInput(path string, input []byte) ([]byte, error) {
	return runExecutable(path, nil, input)
}

func runExecutable(path string, args []string, input []byte) ([]byte, error) {
	cmd := exec.Command(path, args...)
	cmd.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("%w\nstderr:\n%s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}

func recordOracleExpectation(t *testing.T, fixture, status string, err error, outcomes *[]oracleOutcome) {
	t.Helper()
	switch status {
	case oracleStatusPass:
		if err != nil {
			*outcomes = append(*outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "fail"})
			t.Fatalf("fixture marked PASS failed: %v", err)
		}
		*outcomes = append(*outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "pass"})
	case oracleStatusExpectedFail:
		*outcomes = append(*outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "bad status"})
		if err != nil {
			t.Fatalf("fixture %s is marked EXPECTED_FAIL; oracle gaps are no longer accepted: %v", fixture, err)
		}
		t.Fatalf("fixture %s is marked EXPECTED_FAIL; promote it to PASS or mark it SKIP with a reason", fixture)
	case oracleStatusSkip:
		*outcomes = append(*outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "skip"})
		t.Skip("fixture marked SKIP in STATUS.md")
	default:
		*outcomes = append(*outcomes, oracleOutcome{Fixture: fixture, Status: status, Outcome: "bad status"})
		t.Fatalf("unknown oracle status %q for %s", status, fixture)
	}
}

func oracleClassName(fixture string) string {
	base := filepath.Base(fixture)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

func formatOracleOutcomes(outcomes []oracleOutcome) string {
	if len(outcomes) == 0 {
		return "no fixtures ran"
	}
	counts := map[string]int{}
	for _, outcome := range outcomes {
		counts[outcome.Outcome]++
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, counts[key]))
	}
	return strings.Join(parts, ", ")
}
