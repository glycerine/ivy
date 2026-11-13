package fra

import (
	"encoding/json"
	"os"
	"os/exec"
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
		pars := &Fra_Ssw_t{nFramesK: 1}
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
	source, err := os.ReadFile(sourcePath)
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
	cmd := exec.Command("go", "test", "-race", "-count=1", "-overlay", overlayPath, "-run", "^TestInductionStalledRefinementReturnsResult$", "-v", ".")
	cmd.Env = append(os.Environ(), "FRA_TEST_STALLED_REFINEMENT=1")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("induction overlay failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), "Internal error. The result may not verify.") {
		t.Fatalf("stalled branch did not run:\n%s", output)
	}
}

