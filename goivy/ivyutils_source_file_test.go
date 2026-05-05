package goivy

import (
	"os"
	"testing"
)

func TestWithSourceFile(t *testing.T) {
	cfg := NewIvyUtilsConfig()
	cfg.Filename = "original.ivy"
	cfg.WithSourceFile("temp.ivy", func() {
		if cfg.Filename != "temp.ivy" {
			t.Errorf("inside WithSourceFile: Filename = %q, want 'temp.ivy'", cfg.Filename)
		}
	})
	if cfg.Filename != "original.ivy" {
		t.Errorf("after WithSourceFile: Filename = %q, want 'original.ivy'", cfg.Filename)
	}
}

func TestWithWorkingDir(t *testing.T) {
	origDir, _ := os.Getwd()
	// Create a temp dir and resolve symlinks (macOS /var -> /private/var)
	tmpDir, err := os.MkdirTemp("", "ivyutils_test")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = WithWorkingDir(tmpDir, func() {
		// Just verify we changed away from origDir
		cwd, _ := os.Getwd()
		if cwd == origDir {
			t.Error("inside WithWorkingDir: cwd should differ from original")
		}
	})
	if err != nil {
		t.Fatalf("WithWorkingDir error: %v", err)
	}
	cwd, _ := os.Getwd()
	if cwd != origDir {
		t.Errorf("after WithWorkingDir: cwd = %q, want %q", cwd, origDir)
	}
}

func TestWithWorkingDirBadDir(t *testing.T) {
	err := WithWorkingDir("/nonexistent_dir_12345", func() {
		t.Error("should not reach here")
	})
	if err == nil {
		t.Error("WithWorkingDir should return error for nonexistent directory")
	}
}
