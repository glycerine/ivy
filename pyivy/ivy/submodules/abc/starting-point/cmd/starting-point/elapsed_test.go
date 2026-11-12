package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestElapsedFormat(t *testing.T) {
	for _, tc := range []struct {
		duration time.Duration
		want     string
	}{
		{0, "elapsed: 0m 0sec\n"},
		{999 * time.Millisecond, "elapsed: 0m 0sec\n"},
		{59 * time.Second, "elapsed: 0m 59sec\n"},
		{time.Minute, "elapsed: 1m 0sec\n"},
		{123 * time.Second, "elapsed: 2m 3sec\n"},
		{time.Hour + time.Second, "elapsed: 60m 1sec\n"},
	} {
		var output strings.Builder
		printElapsed(&output, tc.duration)
		if output.String() != tc.want {
			t.Fatalf("got %q, want %q", output.String(), tc.want)
		}
	}
}

func TestElapsedIsLastCLIOutput(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(root, "starting-point")
	// TestMain runs in a temporary directory; locate this package from this file.
	_, file, _, _ := runtime.Caller(0)
	build := exec.Command("go", "build", "-o", binary, ".")
	build.Dir = filepath.Dir(file)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("%v: %s", err, out)
	}
	source := filepath.Join(root, "src")
	if err := os.Mkdir(source, 0755); err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(source, "x.c")
	if err := os.WriteFile(input, []byte("#if UNKNOWN\nint x;\n#endif\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := filepath.Join(root, "cfg")
	if err := os.WriteFile(cfg, nil, 0644); err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"tree", "flat", "error"} {
		t.Run(mode, func(t *testing.T) {
			args := []string{"-pre", "-cfg", cfg, "-out", filepath.Join(root, mode)}
			switch mode {
			case "tree":
				args = append(args, source)
			case "flat":
				args = append(args, input)
			case "error":
				args = []string{"-not-a-flag"}
			}
			cmd := exec.Command(binary, args...)
			cmd.Dir = root
			output, err := cmd.CombinedOutput()
			if (err != nil) != (mode == "error") {
				t.Fatalf("%v: %s", err, output)
			}
			text := string(output)
			if !regexp.MustCompile(`(?m)^elapsed: [0-9]+m [0-9]+sec\n\z`).MatchString(text) || strings.Count(text, "elapsed:") != 1 {
				t.Fatal(text)
			}
			if mode != "error" && !strings.Contains(text, "preserved #if UNKNOWN") {
				t.Fatal("missing preceding report", text)
			}
		})
	}
}
