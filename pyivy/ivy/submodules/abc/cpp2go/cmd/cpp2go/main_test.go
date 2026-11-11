package main

import (
	"encoding/json"
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
			files, err := filepath.Glob(filepath.Join(out, "*.go.txt"))
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
	if err := run([]string{"-out", out, "bad-name/file.h"}); err == nil {
		t.Fatal("expected invalid package error")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("created output before validating package", err)
	}
}
