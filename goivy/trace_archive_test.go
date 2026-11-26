package goivy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTraceDirectorySavesIndependentFailures(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "traces")
	cfg := NewConfig()
	if err := ApplyIvyCheckParams(cfg, map[string]string{"trace_dir": dir}); err != nil {
		t.Fatal(err)
	}
	args := []string{"test_vectors/trace_failures.ivy"}
	if err := Start(args, cfg); err == nil || !strings.Contains(err.Error(), "failed checks:") {
		t.Fatalf("verification error = %v, want failed checks", err)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 4 {
		t.Fatalf("wrote %d traces, want four: %v", len(files), files)
	}
	for _, tc := range []struct{ name, witness, forbidden string }{
		{"keep_p", "p = false", "q = false"},
		{"keep_q", "q = false", "p = false"},
		{"requires_p", "p = false", ""},
		{"requires_not_p", "assert [requires_not_p]", "p = false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var view strings.Builder
			if err := ReplayTrace(filepath.Join(dir, tc.name+".a2g"), &view, true, ""); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(view.String(), tc.witness) || (tc.forbidden != "" && strings.Contains(view.String(), tc.forbidden)) {
				t.Fatalf("trace should contain %q and omit %q:\n%s", tc.witness, tc.forbidden, view.String())
			}
		})
	}
	var tracked strings.Builder
	if err := ReplayTrace(filepath.Join(dir, "keep_p.a2g"), &tracked, true, "p & q"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(tracked.String(), "(p & q) = true") || !strings.Contains(tracked.String(), "(p & q) = false") {
		t.Fatalf("trace does not show both tracked values:\n%s", tracked.String())
	}
	if err := Start(args, cfg); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("existing trace directory error = %v", err)
	}
}

func TestTraceDirectorySavesEachInitializerFailure(t *testing.T) {
	const source = `#lang ivy1.8
relation p
after init {
    if p {
        assert [initial_when_true] false
    } else {
        assert [initial_when_false] false
    }
}
`
	cfg := NewConfig()
	cfg.TraceDir = filepath.Join(t.TempDir(), "initializers")
	if err := StartSourceWithConfig("initializers.ivy", source, cfg); err == nil || !strings.Contains(err.Error(), "failed checks: 2") {
		t.Fatalf("verification error = %v, want two failed checks", err)
	}
	for name, value := range map[string]string{"initial_when_true": "p = true", "initial_when_false": "p = false"} {
		var view strings.Builder
		if err := ReplayTrace(filepath.Join(cfg.TraceDir, name+".a2g"), &view, true, "p"); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(view.String(), value) {
			t.Errorf("%s did not record %s:\n%s", name, value, view.String())
		}
	}
}
