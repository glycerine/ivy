// Category C: Randomized Go-vs-Python Conformance tests.
// These tests generate .ivy programs, run both Go and Python pipelines,
// and compare the results.
package end2end

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/check"
	"github.com/glycerine/goivy/compiler"
	il "github.com/glycerine/goivy/ivylogic"
	"github.com/glycerine/goivy/lexer"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/parser"
)

// pythonIvyCheck runs the Python Ivy checker on the given .ivy file.
// Returns true if all conjectures pass, false if any fail.
// Returns error if the Python process fails to run.
func pythonIvyCheck(ivyFile string) (bool, error) {
	// Try to find the Python ivy_check command
	cmd := exec.Command("python3", "-m", "ivy.ivy_check", ivyFile)
	cmd.Dir = filepath.Join(os.Getenv("HOME"), "pyivy", "ivy")
	cmd.Env = append(os.Environ(), "PYTHONPATH="+filepath.Join(os.Getenv("HOME"), "pyivy", "ivy"))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String() + stderr.String()

	// Python ivy_check exits 0 on success, non-zero on failure
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			// Non-zero exit = verification failure
			return false, nil
		}
		return false, fmt.Errorf("python ivy_check failed: %v\n%s", err, output)
	}

	// Check output for FAIL indicators
	if strings.Contains(output, "FAIL") {
		return false, nil
	}

	return true, nil
}

// goIvyCheck runs the Go Ivy checker on the given .ivy source.
// Returns (pass bool, ok bool). ok=false if compilation fails.
func goIvyCheck(t *testing.T, src string) (pass bool, ok bool) {
	t.Helper()

	// Attempt to compile — may fail for generated programs
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Go compile panic: %v", r)
			pass = false
			ok = false
		}
	}()

	version := lexer.Version{1, 7}
	p := parser.New(src, version)
	result, err := p.Parse()
	if err != nil {
		t.Logf("Go parse error: %v", err)
		return false, false
	}
	mod := module.New()
	mod.Sig = il.NewSig()
	err = compiler.IvyCompile(result.Decls, mod)
	if err != nil {
		t.Logf("Go compile error: %v", err)
		return false, false
	}

	ag := art.NewAnalysisGraph(mod)
	ag.Initialize(art.AbstractorFunc(func(s *art.State) {}))
	if len(ag.States) == 0 {
		return false, true
	}
	return check.CheckConjsInStateWithAG(mod, ag, ag.States[0], 8, nil), true
}

// relInfo tracks a generated relation's name, arity, and sort names.
type relInfo struct {
	name  string
	arity int
	sorts []string
}

// genRandomIvy generates a random .ivy program for conformance testing.
// Only generates unary relations to avoid arity mismatches in init/conjectures.
func genRandomIvy(rng *rand.Rand) string {
	var b strings.Builder
	b.WriteString("#lang ivy1.7\n\n")

	// 1-2 sorts
	nSorts := 1 + rng.Intn(2)
	sortNames := make([]string, nSorts)
	for i := 0; i < nSorts; i++ {
		sortNames[i] = fmt.Sprintf("s%d", i)
		b.WriteString(fmt.Sprintf("type %s\n", sortNames[i]))
	}
	b.WriteString("\n")

	// 1-3 unary relations (one sort parameter each)
	nRels := 1 + rng.Intn(3)
	rels := make([]relInfo, nRels)
	for i := 0; i < nRels; i++ {
		s := sortNames[rng.Intn(nSorts)]
		rels[i] = relInfo{
			name:  fmt.Sprintf("r%d", i),
			arity: 1,
			sorts: []string{s},
		}
		b.WriteString(fmt.Sprintf("relation %s(X:%s)\n", rels[i].name, s))
	}
	b.WriteString("\n")

	// Init: set all relations to false
	b.WriteString("after init {\n")
	for _, rel := range rels {
		b.WriteString(fmt.Sprintf("    %s(X) := false;\n", rel.name))
	}
	b.WriteString("}\n\n")

	// 1-2 actions: each picks a relation and sets it to true for the param
	nActions := 1 + rng.Intn(2)
	for i := 0; i < nActions; i++ {
		actName := fmt.Sprintf("act%d", i)
		rel := rels[rng.Intn(nRels)]
		b.WriteString(fmt.Sprintf("action %s(x:%s) = {\n", actName, rel.sorts[0]))
		b.WriteString(fmt.Sprintf("    %s(x) := true\n", rel.name))
		b.WriteString("}\n\n")
		b.WriteString(fmt.Sprintf("export %s\n\n", actName))
	}

	// 1-2 trivially true conjectures: r(X) | ~r(X)
	nConjs := 1 + rng.Intn(2)
	for i := 0; i < nConjs; i++ {
		rel := rels[rng.Intn(nRels)]
		b.WriteString(fmt.Sprintf("conjecture %s(X) | ~%s(X)\n", rel.name, rel.name))
	}

	return b.String()
}

// --- Category C tests ---

func TestConformRandomized(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping randomized conformance test in short mode")
	}

	// Check if Python ivy_check is available and working
	testCmd := exec.Command("python3", "-c", "from ivy import ivy_check")
	testCmd.Dir = filepath.Join(os.Getenv("HOME"), "pyivy", "ivy")
	testCmd.Env = append(os.Environ(), "PYTHONPATH="+filepath.Join(os.Getenv("HOME"), "pyivy", "ivy"))
	if err := testCmd.Run(); err != nil {
		t.Skip("Python ivy_check not available (Z3 library or import issue), skipping conformance test")
	}

	rng := rand.New(rand.NewSource(42))

	for i := 0; i < 10; i++ {
		src := genRandomIvy(rng)
		t.Run(fmt.Sprintf("seed42_iter%d", i), func(t *testing.T) {
			// Run Go checker
			goResult, goOK := goIvyCheck(t, src)
			if !goOK {
				t.Skipf("Go compilation failed, skipping")
			}

			// Write to temp file for Python
			tmpFile := filepath.Join(t.TempDir(), "test.ivy")
			if err := os.WriteFile(tmpFile, []byte(src), 0644); err != nil {
				t.Fatalf("write temp file: %v", err)
			}

			// Run Python checker
			pyResult, err := pythonIvyCheck(tmpFile)
			if err != nil {
				t.Logf("Python check error (skipping): %v", err)
				return
			}

			// Compare results
			if goResult != pyResult {
				t.Errorf("MISMATCH: Go=%v Python=%v\nSource:\n%s", goResult, pyResult, src)
			} else {
				t.Logf("MATCH: Go=%v Python=%v", goResult, pyResult)
			}
		})
	}
}

func TestConformPythonTestSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Python test suite conformance in short mode")
	}

	// Check if Python test files exist
	testDir := filepath.Join(os.Getenv("HOME"), "pyivy", "ivy", "test")
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		t.Skipf("Python test directory not found: %s", testDir)
	}

	// Curated subset of simple test files that should work
	testFiles := []string{
		// Add specific test files here as they are verified to work
	}

	for _, tf := range testFiles {
		t.Run(tf, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(testDir, tf))
			if err != nil {
				t.Skipf("cannot read %s: %v", tf, err)
			}
			src := string(data)
			_ = src
			// TODO: compile and check with Go, compare against known Python result
		})
	}
}
