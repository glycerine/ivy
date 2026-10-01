package ivy2golang

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTargetTestStructFieldNativeBVActionArgsRandomizeWideEpoch(t *testing.T) {
	mod := compileIvySource(t, `#lang ivy1.7
object epoch = {
    type this
    implementation {
        interpret this -> bv[63]
    }
}

type ts = struct {
    version : epoch,
    cid     : epoch
}

import action report(a:ts)

action send_ts(a:ts, b:ts) = {
    call report(a);
    call report(b)
}
export send_ts
`)
	out, err := Generate(mod, Config{Target: "test", ClassName: "native_bv_ts", TestIters: "1", Build: true})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	body := bodyAfterMarker(out.Source, "func (gen *Native_bv_ts_send_ts_generator) generate()")
	if body == "" {
		t.Fatalf("missing send_ts generator body:\n%s", out.Source)
	}
	for _, want := range []string{
		"version: ivyBVRandom(63)",
		"cid: ivyBVRandom(63)",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("native bv timestamp action generator missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "___ivy_randomize(0,") {
		t.Fatalf("native bv timestamp fields should not randomize with zero range:\n%s", body)
	}
}

func TestTargetTestSolverRandomizesStructFieldNativeBVActionArgs(t *testing.T) {
	src := `#lang ivy1.7
object epoch = {
    type this
    relation lt(X:this, Y:this)

    specification {
        axiom lt(X,Y) & lt(Y,Z) -> lt(X,Z)
        axiom ~lt(X,X)
        axiom lt(X,Y) | X = Y | lt(Y,X)
    }

    implementation {
        interpret this -> bv[1]
        definition lt(X,Y) = X < Y
    }
}

type ts = struct {
    version : epoch,
    cid     : epoch
}

function ts_gt(X:ts, Y:ts) = epoch.lt(Y.version, X.version) | ((X.version = Y.version) & epoch.lt(Y.cid, X.cid))

isolate test_timestamp = {
    action send_ts(a:ts)

    relation seen(X:ts)

    implementation {
        implement send_ts(a:ts) {
            require ~seen(a);
            seen(a) := true;
            call report_ts_a(a)
        }
    }
}

import action report_ts_a(a:ts)
export test_timestamp.send_ts
`
	path := filepath.Join(t.TempDir(), "native_bv_ts_require.ivy")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatalf("write test ivy: %v", err)
	}
	out, err := CompileAndGenerate(path, map[string]string{"isolate": "test_timestamp"}, Config{
		Target:    "test",
		ClassName: "native_bv_ts_require",
		TestIters: "1",
		Build:     true,
	})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	for _, want := range []string{
		`ivy.ext__test_timestamp__send_ts(__arg0)`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("native bv timestamp test harness should use isolated ext wrapper %q:\n%s", want, out.Source)
		}
	}
	body := bodyAfterMarker(out.Source, "func (gen *Native_bv_ts_require_test_timestamp__send_ts_generator) __ivy_generate_with_solver() bool")
	if body == "" {
		t.Fatalf("missing send_ts solver generator body:\n%s", out.Source)
	}
	for _, want := range []string{
		`goivy.NewConst("test_timestamp.seen"`,
		`gen.__ivy_solver_failure = "action generator precondition cannot be satisfied"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("native bv timestamp solver generator missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, `baseSMT := "(assert true)"`) {
		t.Fatalf("native bv timestamp solver must assert the require/seen precondition, not solve true:\n%s", body)
	}
	for _, bad := range []string{
		"___ivy_randomize(0,",
	} {
		if strings.Contains(body, bad) {
			t.Fatalf("native bv timestamp solver fields should not use tiny randomization %q:\n%s", bad, body)
		}
	}
	genBody := bodyAfterMarker(out.Source, "func (gen *Native_bv_ts_require_test_timestamp__send_ts_generator) generate() bool")
	if genBody == "" {
		t.Fatalf("missing send_ts public generator body:\n%s", out.Source)
	}
	for _, want := range []string{
		`__ivy_solver_failure string`,
		`__ivy_solver_failure_label string`,
		`ivyFailureEvent("assumption_unsatisfied", __ivy_solver_msg)`,
		`fmt.Fprintf(os.Stderr, "%s: error: %s\n", __ivy_solver_failure_label, __ivy_solver_failure)`,
		`os.Exit(1)`,
	} {
		if !strings.Contains(out.Source, want) {
			t.Fatalf("native bv timestamp generator missing unsatisfied precondition report %q:\n%s", want, out.Source)
		}
	}
	if !strings.Contains(out.Source, `gen.__ivy_solver_failure_label = "native_bv_ts_require.ivy: line 32"`) {
		t.Fatalf("native bv timestamp generator should report the unsatisfied require source line:\n%s", out.Source)
	}
}

func TestTargetTestSolverReportsExhaustedStructNativeBVRequireAtRuntime(t *testing.T) {
	src := `#lang ivy1.7
object epoch = {
    type this
    implementation {
        interpret this -> bv[1]
    }
}

type ts = struct {
    version : epoch,
    cid     : epoch
}

isolate test_timestamp = {
    action send_ts(a:ts)

    relation seen(X:ts)

    implementation {
        implement send_ts(a:ts) {
            require ~seen(a);
            seen(a) := true;
            call report_ts_a(a)
        }
    }
}

import action report_ts_a(a:ts)
export test_timestamp.send_ts
`
	path := filepath.Join(t.TempDir(), "native_bv_ts_runtime_require.ivy")
	if err := os.WriteFile(path, []byte(src), 0644); err != nil {
		t.Fatalf("write test ivy: %v", err)
	}
	out, err := CompileAndGenerate(path, map[string]string{"isolate": "test_timestamp"}, Config{
		Target:    "test",
		ClassName: "native_bv_ts_runtime_require",
		TestIters: "5",
		Build:     true,
	})
	if err != nil {
		t.Fatalf("Generate: %v\n%s", err, outSource(out))
	}
	bin, err := buildOutputForTest(t, out, t.TempDir())
	if err != nil {
		t.Fatalf("build generated Go: %v\n%s", err, out.Source)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "iters=5", "runs=1", "seed=1", "wait=0", "delay=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		t.Fatalf("generated timestamp test hung instead of reporting exhausted require\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
	}
	if err == nil {
		t.Fatalf("generated timestamp test should fail once all struct values have been seen\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
	}
	combined := stdout.String() + stderr.String()
	for _, want := range []string{
		"native_bv_ts_runtime_require.ivy: line 21",
		"action generator precondition cannot be satisfied",
		"assumption_unsatisfied",
	} {
		if !strings.Contains(combined, want) {
			t.Fatalf("generated timestamp test missing %q\nstdout:\n%s\nstderr:\n%s", want, stdout.String(), stderr.String())
		}
	}
	if strings.Contains(combined, "assumption_failed") {
		t.Fatalf("generated timestamp test should report exhausted generator precondition, not action-body assumption failure\nstdout:\n%s\nstderr:\n%s", stdout.String(), stderr.String())
	}
}
