package goivy

import (
	"os"
	"path/filepath"
	"testing"
)

// Python check_isolate_completeness finds every assertion of
// test/crashpong1.ivy checked by some isolate, so check_module goes on
// to check each isolate. goivy reported missing assertions and stopped
// with "some assertions are not checked".
func TestCheckIsolateCompletenessCrashpong1(t *testing.T) {
	fn := filepath.Join(examplesDir(), "test", "crashpong1.ivy")
	if _, err := os.Stat(fn); err != nil {
		t.Skip(err)
	}
	mod, err := IvyLoadFile(fn)
	if err != nil {
		t.Fatal(err)
	}
	if missing := CheckIsolateCompleteness(mod); len(missing) != 0 {
		t.Fatalf("missing %+v", missing)
	}
}
