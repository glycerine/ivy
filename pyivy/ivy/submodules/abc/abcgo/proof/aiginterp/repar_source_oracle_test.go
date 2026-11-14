package aiginterp

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestReparameterizationSourceInitialFailures(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("fixture uses Linux crash-report suppression and GNU section collection")
	}
	cc, err := exec.LookPath("gcc")
	if err != nil {
		t.Skip("gcc unavailable")
	}
	src := filepath.Clean("../../../src")
	source, err := os.ReadFile(filepath.Join(src, "aig/aig/aigRepar.c"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	extract := func(name string) string {
		start := strings.Index(text, "static inline void "+name)
		if start < 0 {
			t.Fatal("missing original helper", name)
		}
		open := start + strings.IndexByte(text[start:], '{')
		end, depth := open+1, 1
		for depth > 0 && end < len(text) {
			if text[end] == '{' {
				depth++
			}
			if text[end] == '}' {
				depth--
			}
			end++
		}
		if depth != 0 {
			t.Fatal("unterminated original helper", name)
		}
		return text[start:end]
	}
	driver := strings.Index(text, "Aig_Man_t * Aig_ManInterRepar(")
	if driver < 0 {
		t.Fatal("missing original driver")
	}
	start := strings.Index(text[driver:], "    // start the solver")
	if start < 0 {
		t.Fatal("missing solver setup")
	}
	start += driver
	end := strings.Index(text[start:], "    // start the interpolant")
	if end < 0 {
		t.Fatal("missing solver boundary")
	}
	end += start
	template, err := os.ReadFile("test_vectors/repar_initial.c.in")
	if err != nil {
		t.Fatal(err)
	}
	program := strings.ReplaceAll(string(template), "/*HELPERS*/", extract("Aig_ManInterAddBuffer")+"\n"+extract("Aig_ManInterAddXor"))
	program = strings.ReplaceAll(program, "/*PREFIX*/", text[start:end])
	dir := t.TempDir()
	fixture, binary := filepath.Join(dir, "repar.c"), filepath.Join(dir, "repar")
	if err := os.WriteFile(fixture, []byte(program), 0600); err != nil {
		t.Fatal(err)
	}
	args := []string{"-DABC_USE_STDINT_H", "-ffunction-sections", "-fdata-sections", "-I", src, fixture}
	for _, file := range []string{"sat/bsat/satSolver.c", "sat/bsat/satSolver2.c", "sat/bsat/satMem.c", "sat/bsat/satProof.c", "sat/bsat/satStore.c", "misc/util/utilSort.c"} {
		args = append(args, filepath.Join(src, file))
	}
	args = append(args, "-Wl,--gc-sections", "-lm", "-o", binary)
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile source oracle: %v: %s", err, out)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, binary).CombinedOutput()
	if err == nil || ctx.Err() != nil || !strings.Contains(string(out), "Id != 0") {
		t.Fatalf("expected original solver2 clause-ID assertion, err=%v context=%v output=%q", err, ctx.Err(), out)
	}
	// Repair only clause IDs in the fixture, leaving all source constraints
	// unchanged. The source then reaches SAT, contradicting its UNSAT assertion.
	args = append([]string{"-DREPAIR_CLAUSE_IDS"}, args...)
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile ID-repaired probe: %v: %s", err, out)
	}
	fixedCtx, fixedCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer fixedCancel()
	out, err = exec.CommandContext(fixedCtx, binary).CombinedOutput()
	if err != nil || strings.TrimSpace(string(out)) != "status=1" {
		t.Fatalf("ID-repaired original constraints: err=%v output=%q", err, out)
	}

}
