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

func TestTraceTrackingWorksAcrossCheckers(t *testing.T) {
	dir := t.TempDir()
	source, err := filepath.Abs("test_vectors/trace_failures.ivy")
	if err != nil {
		t.Fatal(err)
	}
	pyTraces := filepath.Join(dir, "python")
	check := pythonIvyCommandForTest(t, "-O", "-m", "ivy.ivy", "check", "trace_dir="+pyTraces, source)
	if out, err := check.CombinedOutput(); err == nil || !strings.Contains(string(out), "failed checks:") {
		t.Fatalf("PyIvy did not report the expected failures: %v\n%s", err, out)
	}
	converted := filepath.Join(dir, "portable.a2g")
	convert := pythonIvyCommandForTest(t, "-O", "-m", "ivy.ivy", "replay", "export_trace="+converted, filepath.Join(pyTraces, "keep_p.a2g"))
	if out, err := convert.CombinedOutput(); err != nil {
		t.Fatalf("PyIvy could not convert its trace: %v\n%s", err, out)
	}
	var goReplay strings.Builder
	if err := ReplayTrace(converted, &goReplay, true, "p & q"); err != nil {
		t.Fatal(err)
	}
	goCfg := NewConfig()
	goCfg.TraceDir = filepath.Join(dir, "go")
	if err := Start([]string{source}, goCfg); err == nil || !strings.Contains(err.Error(), "failed checks:") {
		t.Fatalf("GoIvy did not report the expected failures: %v", err)
	}
	pyReplay := pythonIvyCommandForTest(t, "-O", "-m", "ivy.ivy", "replay", "track=p & q", filepath.Join(goCfg.TraceDir, "keep_p.a2g"))
	out, err := pyReplay.CombinedOutput()
	if err != nil {
		t.Fatalf("PyIvy could not replay the Go archive: %v\n%s", err, out)
	}
	for _, replay := range []string{goReplay.String(), string(out)} {
		for _, want := range []string{"(p & q) = true", "(p & q) = false"} {
			if !strings.Contains(replay, want) {
				t.Errorf("cross-checker replay is missing %q:\n%s", want, replay)
			}
		}
	}
}

func TestPyIvyReplaysPortableTraceWithLeadingWhitespace(t *testing.T) {
	path := filepath.Join(t.TempDir(), "formatted.a2g")
	const archive = `{"format":"ivy-trace-1","detailed":"detected counterexample","brief":"brief trace"}`
	contents := strings.Repeat(" \n\t", 8192) + archive
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	replay := pythonIvyCommandForTest(t, "-O", "-m", "ivy.ivy", "replay", "detailed=true", path)
	if out, err := replay.CombinedOutput(); err != nil || string(out) != "detected counterexample" {
		t.Fatalf("could not replay the formatted portable trace: %v\n%s", err, out)
	}
}
