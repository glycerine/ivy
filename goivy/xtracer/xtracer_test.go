//go:build !xtrace_off && !js

package xtracer

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeLine_IvyExamples(t *testing.T) {
	home := os.Getenv("HOME")
	line := filepath.Join(home, "/go/src/github.com/glycerine/ivy/ivy-lang-examples/doc/examples/apple/ord_live.ivy")
	line2 := NormalizeLine(line)
	if line2 == line {
		//fmt.Printf("bad! same: %v\n", line2)
		t.Fatalf("IVY_EXAMPLES replacement failed!")
	} else {
		fmt.Printf("good: %v\n", line2)
	}
}

func TestNormalizeLinePrefersMountedRepoExamplesPath(t *testing.T) {
	line := filepath.Join(repo, "ivy-lang-examples/doc/examples/apple/ord_live.ivy")
	got := NormalizeLine(line)
	want := "<IVY_EXAMPLES>/doc/examples/apple/ord_live.ivy"
	if got != want {
		t.Fatalf("NormalizeLine(%q) = %q, want %q", line, got, want)
	}

	examplesDir := filepath.Join(repo, "ivy-lang-examples")
	realExamplesDir, err := filepath.EvalSymlinks(examplesDir)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", examplesDir, err)
	}
	realLine := filepath.Join(realExamplesDir, "doc/examples/apple/ord_live.ivy")
	got = NormalizeLine(realLine)
	if got != want {
		t.Fatalf("NormalizeLine(%q) = %q, want %q", realLine, got, want)
	}
}
