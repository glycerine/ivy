//go:build !xtracer_off && wasip1

package xtracer

import "testing"

func TestNormalizeLineWasip1PassThrough(t *testing.T) {
	line := "/browser/virtual/path/example.ivy"
	if got := NormalizeLine(line); got != line {
		t.Fatalf("NormalizeLine() = %q, want %q", got, line)
	}
}
