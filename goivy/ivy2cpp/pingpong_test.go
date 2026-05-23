package ivy2cpp

import (
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

	if SlowCppTest {
		compileGeneratedCPP(t, out)
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
	isoMod.Cfg.IsolateCfg.CompileWithInvariants = languageVersionAtLeast(isoMod, "1.7")
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
