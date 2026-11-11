package main

import (
	"bytes"
	"encoding/json"
	startingpoint "github.com/glycerine/abc/starting-point"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSharedOutputPackage(t *testing.T) {
	for _, override := range []string{"", "satoko"} {
		t.Run("override="+override, func(t *testing.T) {
			root := t.TempDir()
			var sources []string
			for _, dir := range []string{"vec", "utils"} {
				path := filepath.Join(root, dir, "input.h")
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("static inline int f() { return 1; }"), 0644); err != nil {
					t.Fatal(err)
				}
				sources = append(sources, path)
			}
			out := filepath.Join(root, "unrelated_destination")
			args := []string{"-out", out}
			if override != "" {
				args = append(args, "-package", override)
			}
			args = append(args, sources...)
			if err := run(args); err != nil {
				t.Fatal(err)
			}
			want := override
			if want == "" {
				want = "vec"
			}
			files, err := filepath.Glob(filepath.Join(out, "*.go"))
			if err != nil || len(files) != 2 {
				t.Fatal(files, err)
			}
			for _, path := range files {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(string(data), "package "+want+"\n") {
					t.Fatalf("%s: %s", path, data)
				}
			}
			data, err := os.ReadFile(filepath.Join(out, "run.json"))
			if err != nil {
				t.Fatal(err)
			}
			var report struct{ Package string }
			if err := json.Unmarshal(data, &report); err != nil {
				t.Fatal(err)
			}
			if report.Package != want {
				t.Fatal(report)
			}
		})
	}
}

func TestInvalidPackageBeforeOutputCreation(t *testing.T) {
	out := filepath.Join(t.TempDir(), "output")
	if err := run([]string{"-out", out, "-package", "bad-name", "file.h"}); err == nil {
		t.Fatal("expected invalid package error")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("created output before validating package", err)
	}
}

func TestDirectoryCommand(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "abc")
	out := filepath.Join(root, "out")
	for _, dir := range []string{"vec", "utils"} {
		path := filepath.Join(source, dir, "input.h")
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("static inline int f() { return 1; }"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := run([]string{"-out", out, source}); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"vec", "utils"} {
		data, err := os.ReadFile(filepath.Join(out, dir, "input.go"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.HasPrefix(string(data), "package "+dir+"\n") {
			t.Fatal(string(data))
		}
	}
}

func TestSourceArgumentCheckoutName(t *testing.T) {
	parent := t.TempDir()
	checkout := filepath.Join(parent, "abc")
	if err := os.Mkdir(checkout, 0755); err != nil {
		t.Fatal(err)
	}
	// Reproduce ABC's same-named executable inside its source checkout.
	if err := os.WriteFile(filepath.Join(checkout, "abc"), []byte("executable"), 0755); err != nil {
		t.Fatal(err)
	}
	for _, cwd := range []string{parent, checkout} {
		source, isDir, err := resolveSourceArgument("abc", cwd)
		if err != nil || !isDir || source != checkout {
			t.Fatal(source, isDir, err)
		}
	}
	if _, _, err := resolveSourceArgument("missing", parent); err == nil || !strings.Contains(err.Error(), "cannot access source") {
		t.Fatal(err)
	}
	file, isDir, err := resolveSourceArgument(filepath.Join(checkout, "abc"), parent)
	if err != nil || isDir || file != filepath.Join(checkout, "abc") {
		t.Fatal(file, isDir, err)
	}
}

func TestFallbackDetailsPrinted(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "abc")
	out := filepath.Join(root, "out")
	if err := os.Mkdir(source, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(source, "broken.c"), []byte("/* unfinished"), 0644); err != nil {
		t.Fatal(err)
	}
	report, err := startingpoint.TranslateTree(source, out, startingpoint.Options{})
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	printTreeReport(&output, report)
	for _, want := range []string{"1 source-preserving fallbacks", "commented original source; analysis failed", "broken.c -> broken.go: unterminated comment at source line 1, column 1 (byte offset 0)"} {
		if !strings.Contains(output.String(), want) {
			t.Fatal(output.String())
		}
	}
	data, err := os.ReadFile(filepath.Join(out, "run.json"))
	if err != nil {
		t.Fatal(err)
	}
	var saved startingpoint.TreeReport
	if err = json.Unmarshal(data, &saved); err != nil {
		t.Fatal(err)
	}
	if saved.Files[0].FallbackLocation == nil || saved.Files[0].FallbackLocation.Line != 1 || saved.Files[0].FallbackLocation.Column != 1 || saved.Files[0].FallbackLocation.ByteOffset != 0 {
		t.Fatal(saved)
	}
	if saved.Files[0].FallbackReason != "unterminated comment at source line 1, column 1 (byte offset 0)" {
		t.Fatal(saved)
	}
	output.Reset()
	printTreeReport(&output, startingpoint.TreeReport{})
	if strings.Contains(output.String(), "Fallbacks (") {
		t.Fatal("unexpected fallback listing")
	}
}
