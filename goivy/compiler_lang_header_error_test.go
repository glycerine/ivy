package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Python read_module locates the missing-#lang error at line 1 of the
// file (test/asgeir2.ivy).
func TestMissingLangHeaderErrorHasLocation(t *testing.T) {
	fn := filepath.Join(t.TempDir(), "nolang.ivy")
	// A blank first line is not an empty-file error in Python either
	// (test/callback.ivy).
	for _, src := range []string{"type t\n", "\n#lang ivy1.7\n", ""} {
		if err := os.WriteFile(fn, []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
		_, err := ReadModule(fn, false, New().Cfg)
		if err == nil {
			t.Fatal("no error")
		}
		if got, want := pyErrorString(err), `nolang.ivy: line 1: error: file must begin with "#lang ivyN.N"`; !strings.HasSuffix(got, want) {
			t.Fatalf("%q: %q", src, got)
		}
	}
}
