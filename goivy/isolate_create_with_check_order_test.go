package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Python create_isolate runs check_with_parameters before
// apply_present_conjectures, so a bad "with" name fails first
// (test/badwith1.ivy).
func TestCreateIsolateChecksWithBeforePresentConjectures(t *testing.T) {
	src := "#lang ivy1.7\n\nisolate foo = {\n} with xxx\n"
	fn := filepath.Join(t.TempDir(), "badwith.ivy")
	if err := os.WriteFile(fn, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}
	mod, err := IvyLoadFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	var cerr error
	out := captureActionUpdateStdout(t, func() { cerr = CreateIsolate("foo", mod) })
	if cerr == nil || !strings.Contains(cerr.Error(), "xxx is not an object") {
		t.Fatalf("err %v", cerr)
	}
	if !strings.Contains(out, "check.CreateIsolate ENTER") {
		t.Fatalf("traces not captured:\n%s", out)
	}
	if strings.Contains(out, "ApplyPresentConjectures") {
		t.Fatalf("present conjectures applied before the with check:\n%s", out)
	}
}
