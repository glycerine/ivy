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
