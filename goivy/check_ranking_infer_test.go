package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRankingInferenceExportsTypedReaction(t *testing.T) {
	source, err := os.ReadFile("../pyivy/ivy/test/ranking_infer1.ivy")
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ranking.ivy")
	if err := os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Start([]string{path}, NewConfig()); err != nil {
		t.Fatal(err)
	}
	modelPath := strings.TrimSuffix(path, ".ivy") + ".vmt"
	model, err := os.ReadFile(modelPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, tag := range []string{":react_p true", ":react_q true", ":react_r true", ":init true", ":trans true"} {
		if !strings.Contains(string(model), tag) {
			t.Errorf("VMT model is missing %q", tag)
		}
	}
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
from ivy import z3
from pathlib import Path
import sys

model = Path(sys.argv[1]).read_text()
z3.parse_smt2_string(model + "\n(assert $init)\n(assert $trans)")
`, modelPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("Z3 rejected ranking model: %v\n%s", err, out)
	}
}
