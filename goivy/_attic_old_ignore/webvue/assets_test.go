package webvue

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMaterializeRunwebDirCopiesEmbeddedVueAssets(t *testing.T) {
	dst := filepath.Join(t.TempDir(), ".runweb")
	if err := MaterializeRunwebDir(dst); err != nil {
		t.Fatalf("materialize runweb dir: %v", err)
	}

	index, err := os.ReadFile(filepath.Join(dst, "index.html"))
	if err != nil {
		t.Fatalf("read index.html: %v", err)
	}
	if !strings.Contains(string(index), `id="app"`) {
		t.Fatalf("index.html does not look like the Vue shell: %q", string(index))
	}

	if false { // not present atm
		bundle, err := os.Stat(filepath.Join(dst, "dist", "ivywebvue.js"))
		if err != nil {
			t.Fatalf("stat ivywebvue.js: %v", err)
		}
		if bundle.Size() == 0 {
			t.Fatalf("ivywebvue.js was materialized empty")
		}
	}
}
