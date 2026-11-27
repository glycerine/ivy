package cudd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// testcudd prints its lap and total times with util_print_time ("%ld.%02ld
// sec" of milliseconds); the port printed Go durations ("total time = 3ms").
func TestTestCuddTimesLikeUtilPrintTime(t *testing.T) {
	dir := t.TempDir()
	matrix := filepath.Join(dir, "m.mat")
	if err := os.WriteFile(matrix, []byte("2 2\n0 0 1\n1 1 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout, err := os.Create(filepath.Join(dir, "out"))
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := os.Create(filepath.Join(dir, "err"))
	if err != nil {
		t.Fatal(err)
	}
	runTestCudd([]string{"-p", "1", matrix}, stdout, stderr)
	stdout.Close()
	stderr.Close()
	out, err := os.ReadFile(filepath.Join(dir, "out"))
	if err != nil {
		t.Fatal(err)
	}
	// (this small run stops before "total time =", which uses the same timer)
	re := `(?m):2: time to read the matrix = [0-9]+\.[0-9][0-9] sec$`
	if !regexp.MustCompile(re).Match(out) {
		t.Errorf("no match for %s in:\n%s", re, strings.TrimSpace(string(out)))
	}
}
