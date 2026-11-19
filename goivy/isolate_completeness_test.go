package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Python check_isolate_completeness finds every assertion of
// test/crashpong1.ivy checked by some isolate, so check_module goes on
// to check each isolate. goivy reported missing assertions and stopped
// with "some assertions are not checked".
// test/paramiso.ivy likewise.
func TestCheckIsolateCompletenessMatchesPython(t *testing.T) {
	for _, name := range []string{"crashpong1", "paramiso"} {
		fn := filepath.Join(examplesDir(), "test", name+".ivy")
		if _, err := os.Stat(fn); err != nil {
			t.Skip(err)
		}
		mod, err := IvyLoadFile(fn)
		if err != nil {
			t.Fatal(name, err)
		}
		if missing := CheckIsolateCompleteness(mod); len(missing) != 0 {
			t.Errorf("%s: missing %+v", name, missing)
		}
	}
}

// Python check_isolate_completeness prints each unchecked assertion,
// its mixee action and caller before check_module raises
// (test/this2.ivy).
func TestCheckIsolateCompletenessPrintsLikePython(t *testing.T) {
	fn := filepath.Join(examplesDir(), "test", "this2.ivy")
	if _, err := os.Stat(fn); err != nil {
		t.Skip(err)
	}
	mod, err := IvyLoadFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	var missing []IsolateError
	out := captureActionUpdateStdout(t, func() { missing = CheckIsolateCompleteness(mod) })
	if len(missing) == 0 {
		t.Fatal("no missing assertions")
	}
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		if l != "" && !strings.HasPrefix(l, "XTRACE") {
			lines = append(lines, l)
		}
	}
	want := []string{
		"this2.ivy: line 20: error: assertion is not checked",
		"this2.ivy: line 4: error: ...in action foo",
		"error: ...when called from the environment",
	}
	if len(lines) != len(want) {
		t.Fatalf("got %q", lines)
	}
	for i, w := range want {
		if !strings.HasSuffix(lines[i], w) {
			t.Fatalf("line %d: got %q, want suffix %q", i, lines[i], w)
		}
	}
}
