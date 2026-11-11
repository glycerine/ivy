package startingpoint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func putSource(t *testing.T, root, name, text string) {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestTranslateTree(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "abc")
	out := filepath.Join(root, "starting-point-cpp2go-abc")
	for _, name := range []string{"top.h", "src/vec/x.c", "src/vec/x.h", "src/logic/x.cpp", "src/bad-name/y.h"} {
		putSource(t, source, name, "static inline int f() { return 1; }")
	}
	putSource(t, source, "src/broken/bad.c", "/* unterminated")
	for _, name := range []string{".git/ignored.c", "starting-point-cpp2go-old/ignored.h", "src/readme.txt"} {
		putSource(t, source, name, "ignored")
	}
	report, err := TranslateTree(source, out, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"top.go", "src/vec/x.c.go", "src/vec/x.h.go", "src/logic/x.go", "src/bad-name/y.go", "src/broken/bad.go"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatal(err)
		}
	}
	if len(report.Files) != 6 || report.Fallbacks != 1 {
		t.Fatalf("%+v", report)
	}
	for _, f := range report.Files {
		bytes, err := os.ReadFile(filepath.Join(out, f.Output))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(bytes), "package "+f.Package+"\n") {
			t.Fatal(f)
		}
		if f.Output == "top.go" && f.Package != "starting_point_cpp2go_abc" {
			t.Fatal(f)
		}
		if strings.HasPrefix(f.Output, "src/vec/") && f.Package != "vec" {
			t.Fatal(f)
		}
		if f.Output == "src/bad-name/y.go" && f.Package != "bad_name" {
			t.Fatal(f)
		}
		if f.Fallback && !strings.Contains(string(bytes), "// /* unterminated") {
			t.Fatal("source lost")
		}
	}
	if _, err := TranslateTree(source, out, Options{}); err == nil {
		t.Fatal("overwrote output")
	}
	override := filepath.Join(root, "explicit")
	report, err = TranslateTree(source, override, Options{Package: "shared"})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range report.Files {
		if f.Package != "shared" {
			t.Fatal(f)
		}
	}
}

func TestTreeDestinationInsideSource(t *testing.T) {
	source := filepath.Join(t.TempDir(), "abc")
	putSource(t, source, "one.h", "static inline void f() {}")
	out := filepath.Join(source, "output")
	report, err := TranslateTree(source, out, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(report.Files) != 1 {
		t.Fatal(report)
	}
	name, err := DefaultTreeOutput(source)
	if err != nil {
		t.Fatal(err)
	}
	if name != "starting-point-cpp2go-abc" {
		t.Fatal(name)
	}
}
