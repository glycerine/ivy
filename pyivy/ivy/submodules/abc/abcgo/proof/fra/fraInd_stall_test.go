package fra

import (
	"os/exec"
	"encoding/json"
	"github.com/glycerine/abcgo/internal/testoracle"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glycerine/abcgo/aig/aig"
)

// ABC fraInd.c warns and breaks on stalled refinement, then constructs a
// result. The same fixture with the condition injected into compiled C returns
// a graph with zero iterations. Force this internal error in a Go overlay,
// without adding a fault-injection hook to the production implementation.
func TestInductionStalledRefinementReturnsResult(t *testing.T) {
	if os.Getenv("FRA_TEST_STALLED_REFINEMENT") == "1" {
		g := aig.NewGraph()
		in, a, b := g.NewInput(), g.NewInput(), g.NewInput()
		g.AddOutput(g.XOR(a, b))
		g.AddOutput(in)
		g.AddOutput(in)
		aig.Aig_ManSetRegNum(g, 2)
		pars := &Fra_Ssw_t{NFramesK: 1}
		result := Fra_FraigInduction(g, pars)
		if result == nil {
			t.Fatal("stalled induction must return the current reduction")
		}
		if pars.nIters != 0 {
			t.Fatalf("stalled induction iterations = %d, want 0", pars.nIters)
		}
		return
	}
	sourcePath, err := filepath.Abs("fraInd.go")
	if err != nil {
		t.Fatal(err)
	}
	source, err := testoracle.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	needle := "\t\tif p.PCla.FRefinement != 0 && oldLits == Fra_ClassesCountLits(p.PCla) && oldImps == imps && oldHots == hots {"
	if strings.Count(string(source), needle) != 1 {
		t.Fatal("induction stall condition changed; update test overlay")
	}
	// Set the exact condition checked in ABC, after the real sweep, while
	// preserving all later production control flow and result construction.
	injection := "\t\tp.PCla.FRefinement = 1\n\t\toldLits = Fra_ClassesCountLits(p.PCla)\n\t\toldImps, oldHots = imps, hots\n"
	dir := t.TempDir()
	replacement := filepath.Join(dir, "fraInd.go")
	if err := os.WriteFile(replacement, []byte(strings.Replace(string(source), needle, injection+needle, 1)), 0600); err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(struct{ Replace map[string]string }{map[string]string{sourcePath: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(dir, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", fraInductionOverlayArgs(overlayPath, fraTestRaceEnabled)...)
	cmd.Env = append(os.Environ(), "FRA_TEST_STALLED_REFINEMENT=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("induction overlay failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "Internal error. The result may not verify.") {
		t.Fatalf("stalled branch did not run:\n%s", output)
	}
}

// Match the parent test's instrumentation; ordinary tests must not require
// a C toolchain just because this regression uses a child Go test.
func fraInductionOverlayArgs(path string, race bool) []string {
	args := []string{"test"}
	if race {
		args = append(args, "-race")
	}
	return append(args, "-count=1", "-overlay", path, "-run", "^TestInductionStalledRefinementReturnsResult$", "-v", ".")
}

func TestInductionOverlayPreservesRaceMode(t *testing.T) {
	for _, race := range []bool{false, true} {
		args := fraInductionOverlayArgs("overlay.json", race)
		found := false
		for _, arg := range args {
			if arg == "-race" {
				found = true
			}
		}
		if found != race {
			t.Fatalf("race=%t: child args %q", race, args)
		}
		if !strings.Contains(strings.Join(args, " "), "-overlay overlay.json -run ^TestInductionStalledRefinementReturnsResult$ -v .") {
			t.Fatalf("missing overlay or selected test: %q", args)
		}
	}
}
