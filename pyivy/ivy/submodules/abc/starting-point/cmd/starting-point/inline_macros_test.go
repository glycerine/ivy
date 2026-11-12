package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInlineMacrosDefault(t *testing.T) {
	for _, tree := range []bool{false, true} {
		t.Run(fmt.Sprintf("tree=%v", tree), func(t *testing.T) {
			root := t.TempDir()
			src := filepath.Join(root, "src")
			if err := os.Mkdir(src, 0755); err != nil {
				t.Fatal(err)
			}
			source := filepath.Join(src, "sample.c")
			if err := os.WriteFile(source, []byte("#define PLUS(x) ((x)+1)\nint f() { return PLUS(2); }\n"), 0644); err != nil {
				t.Fatal(err)
			}
			cfg := filepath.Join(root, "config")
			if err := os.WriteFile(cfg, nil, 0644); err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(root, "out")
			args := []string{"-cfg", cfg, "-out", out}
			if tree {
				args = append(args, src)
			} else {
				args = append(args, source)
			}
			if err := run(args); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(filepath.Join(out, "sample.go"))
			if err != nil {
				t.Fatal(err)
			}
			text := string(data)
			if strings.Contains(text, "func PLUS(") || !strings.Contains(text, "return ((2)+1);") {
				t.Fatal(text)
			}
		})
	}
}

func TestInlineMacrosFlagRemoved(t *testing.T) {
	if err := run([]string{"-inlinemacros"}); err == nil || !strings.Contains(err.Error(), "flag provided but not defined") {
		t.Fatal(err)
	}
}
