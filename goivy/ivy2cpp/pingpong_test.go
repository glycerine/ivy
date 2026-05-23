package ivy2cpp

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

const pingPongIvySource = `#lang ivy1.7

object intf = {
    action ping
    action pong
}

type side_t = {left,right}

specification {
    individual side : side_t
    after init {
        side := left
    }

    before intf.ping {
        require side = left;
        side := right
    }

    before intf.pong {
        require side = right;
        side := left
    }
}

implementation {
    isolate left_player = {
        individual ball : bool
        after init {
            ball := true
        }

        action hit = {
            if ball {
                call intf.ping;
                ball := false
            }
        }

        implement intf.pong {
            ball := true
        }

        invariant ball -> side = left
    } with this

    isolate right_player = {
        individual ball : bool
        after init {
            ball := false
        }

        action hit = {
            if ball {
                call intf.pong;
                ball := false
            }
        }

        implement intf.ping {
            ball := true
        }

        invariant ball -> side = right
    } with this
}

export left_player.hit
export right_player.hit
`

func TestPingPongLeftPlayerTargetTestStateConstraintUsesSolver(t *testing.T) {
	out := generatePingPongLeftPlayerTargetTest(t)
	impl := compactCPPForPingPongTest(out.Impl)

	// Python ivy_to_cpp emits slvr.add(...) here. Go must not route this
	// z3::expr through gen::add(...), whose Python-style signature accepts
	// SMT-LIB strings.
	want := `slvr.add(__to_solver(*this,apply("side"),obj.side));`
	if !strings.Contains(impl, want) {
		t.Errorf("ping-pong target=test state constraint should use %q; generated impl:\n%s", want, out.Impl)
	}

	intfPingBody := bodyAfterMarker(out.Impl, "pingpong::intf__ping")
	if intfPingBody == "" {
		t.Fatalf("intf__ping body not emitted:\n%s", out.Impl)
	}
	if !strings.Contains(intfPingBody, `__ivy_out << "< intf.ping"`) {
		t.Errorf("intf__ping should trace the imported call like Python:\n%s", intfPingBody)
	}

	impPingBody := bodyAfterMarker(out.Impl, "pingpong::imp__intf__ping")
	if impPingBody == "" {
		t.Fatalf("imp__intf__ping body not emitted:\n%s", out.Impl)
	}
	if strings.Contains(impPingBody, `__ivy_out << "< imp__intf.ping"`) {
		t.Errorf("implementation stub should not carry the imported-call trace:\n%s", impPingBody)
	}

	invariantExpr := `ivy_assert((!(left_player__ball) || ((side == left)))`
	for _, marker := range []string{
		"pingpong::__init",
		"pingpong::ext__intf__pong",
		"pingpong::ext__left_player__hit",
	} {
		body := bodyAfterMarker(out.Impl, marker)
		if body == "" {
			t.Fatalf("%s body not emitted:\n%s", marker, out.Impl)
		}
		if strings.Contains(body, invariantExpr) {
			t.Errorf("%s should match Python ivy1.7 target=test output without appended invariant checks:\n%s", marker, body)
		}
	}

	if SlowCppTest {
		runGeneratedPingPongTestSlow(t, out)
	}
}

func generatePingPongLeftPlayerTargetTest(t *testing.T) *Output {
	t.Helper()
	cfg, _, err := normalizeConfig(Config{
		Target:    "test",
		ClassName: "pingpong",
		TestIters: "30",
		TestRuns:  "1",
	})
	if err != nil {
		t.Fatalf("normalize config: %v", err)
	}

	mod := goivy.New()
	mod.Cfg = goivy.NewConfig()
	mod.Cfg.Isolate = "left_player"
	applySessionParameters(mod, cfg)
	sig := goivy.NewSigOn(mod.Cfg.IuCfg)
	if err := goivy.SourceString("pingpong.ivy", pingPongIvySource, mod, sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatalf("compile ping-pong source: %v", err)
	}

	isoMod := mod.Copy()
	if isoMod.Cfg == nil {
		isoMod.Cfg = goivy.NewConfig()
	}
	isoMod.Cfg.Isolate = "left_player"
	isoMod.Cfg.IsolateCfg.CompileWithInvariants = languageVersionAfter(isoMod, "1.7")
	cppIface := snapshotCPPInterface(isoMod)
	if err := goivy.CreateIsolate("left_player", isoMod); err != nil {
		t.Fatalf("create left_player isolate: %v", err)
	}
	pruneStateStoresToSignature(isoMod)
	restoreCPPInterface(isoMod, cppIface)
	prepareModuleForCPP(isoMod, cfg)

	out, err := Generate(isoMod, cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return out
}

func compactCPPForPingPongTest(s string) string {
	replacer := strings.NewReplacer(" ", "", "\t", "", "\n", "", "\r", "")
	return replacer.Replace(s)
}

func runGeneratedPingPongTestSlow(t *testing.T, out *Output) {
	t.Helper()
	path, err := BuildOutput(out, t.TempDir())
	if err != nil {
		if isMissingZ3ToolchainError(err) {
			t.Skip(err.Error())
		}
		t.Fatalf("build generated ping-pong test: %v", err)
	}
	cmd := exec.Command(path, "iters=30", "runs=1", "seed=1")
	data, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("run generated ping-pong test: %v\n%s", err, data)
	}
	got := string(data)
	if !strings.Contains(got, "< intf.ping") {
		t.Fatalf("generated ping-pong run should include imported-call trace; output:\n%s", got)
	}
	if !strings.Contains(got, "test_completed") {
		t.Fatalf("generated ping-pong run should finish; output:\n%s", got)
	}
}
