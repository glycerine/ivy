package goivy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIvyVersionSupportedIncludesV16(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "v16.ivy")
	if err := os.WriteFile(path, []byte("#lang ivy1.6\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ok, err := IvyVersionSupported(path)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("ivy1.6 should be supported")
	}
}

func TestIvyVersionSupportedStillRejectsBeforeV16(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "v15.ivy")
	if err := os.WriteFile(path, []byte("#lang ivy1.5\n"), 0644); err != nil {
		t.Fatal(err)
	}

	ok, err := IvyVersionSupported(path)
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("ivy1.5 should not be reported as supported")
	}
}
