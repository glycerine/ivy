package goivy

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestListIsolatesHandlesSorryTacticDuringLoad(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve test file path")
	}
	path := filepath.Join(filepath.Dir(file), "..", "ivy-lang-examples", "doc", "examples", "cav2024", "examp1_numeric.ivy")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("example not available at %s: %v", path, err)
	}

	isolates, err := ListIsolates(path)
	if err != nil {
		t.Fatalf("ListIsolates failed on source with tactic sorry: %v", err)
	}
	if len(isolates) == 0 {
		t.Fatal("expected at least one isolate")
	}
}

func TestListIsolatesUnknownTypeReturnsError(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("unable to resolve test file path")
	}
	path := filepath.Join(filepath.Dir(file), "..", "ivy-lang-examples", "test", "cont2b.ivy")
	if _, err := os.Stat(path); err != nil {
		t.Skipf("example not available at %s: %v", path, err)
	}

	_, err := ListIsolates(path)
	if err == nil {
		t.Fatal("expected unknown type error, got nil")
	}
	if !strings.Contains(err.Error(), "unknown type: foo") {
		t.Fatalf("ListIsolates error = %v, want unknown type: foo", err)
	}
}
