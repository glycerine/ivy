package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStandardLibraryIncludeSourceSelectsVersionedDirectory(t *testing.T) {
	base := t.TempDir()
	writeStdlibTestFile(t, base, "1.6", "marker.ivy", "#lang ivy\ntype from16\n")
	writeStdlibTestFile(t, base, "1.7", "marker.ivy", "#lang ivy\ntype from17\n")
	writeStdlibTestFile(t, base, "1.8", "marker.ivy", "#lang ivy\ntype from18\n")

	lib, err := loadStandardLibrary(base)
	if err != nil {
		t.Fatalf("loadStandardLibrary: %v", err)
	}

	for _, tc := range []struct {
		languageVersion string
		wantVersion     string
		wantSource      string
	}{
		{languageVersion: "1.5", wantVersion: "1.6", wantSource: "type from16"},
		{languageVersion: "1.6", wantVersion: "1.6", wantSource: "type from16"},
		{languageVersion: "1.7", wantVersion: "1.7", wantSource: "type from17"},
		{languageVersion: "1.8", wantVersion: "1.8", wantSource: "type from18"},
	} {
		t.Run(tc.languageVersion, func(t *testing.T) {
			filename, source, ok := lib.includeSource(tc.languageVersion, "marker")
			if !ok {
				t.Fatalf("includeSource(%q, marker) ok = false", tc.languageVersion)
			}
			if !strings.Contains(filename, filepath.Join(tc.wantVersion, "marker.ivy")) {
				t.Fatalf("includeSource(%q, marker) filename = %q, want version %s", tc.languageVersion, filename, tc.wantVersion)
			}
			if !strings.Contains(source, tc.wantSource) {
				t.Fatalf("includeSource(%q, marker) source = %q, want %q", tc.languageVersion, source, tc.wantSource)
			}
		})
	}

	if _, _, ok := lib.includeSource("1.9", "marker"); ok {
		t.Fatal("includeSource(1.9, marker) ok = true, want false with no matching future stdlib")
	}
}

func writeStdlibTestFile(t *testing.T, base, version, name, source string) {
	t.Helper()
	dir := filepath.Join(base, version)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", dir, err)
	}
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}
