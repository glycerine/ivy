package ivy2golang

import (
	"strings"
	"testing"

	"github.com/glycerine/ivy/goivy"
)

func TestPingPongLeftPlayerTargetTestCallsImportSoEnvironmentCanRun(t *testing.T) {
	out := generatePingPongLeftPlayerTargetTest(t)

	leftHitBody := bodyAfterMarker(out.Source, "func (ivy *pingpong) ext__left_player__hit")
	if leftHitBody == "" {
		t.Fatalf("left_player.hit body not emitted:\n%s", out.Source)
	}
	if !strings.Contains(leftHitBody, "ivy.intf__ping()") {
		t.Fatalf("left_player.hit should call intf.ping so intf.pong can become enabled:\n%s", leftHitBody)
	}
	if !strings.Contains(out.Source, `fmt.Fprintln(__ivy_out, "> intf.pong")`) {
		t.Fatalf("target=test loop should include the environment-side intf.pong action:\n%s", out.Source)
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
	goIface := snapshotGoInterface(isoMod)
	if err := goivy.CreateIsolate("left_player", isoMod); err != nil {
		t.Fatalf("create left_player isolate: %v", err)
	}
	pruneStateStoresToSignature(isoMod)
	restoreGoInterface(isoMod, goIface)
	prepareModuleForGo(isoMod, cfg)

	out, err := Generate(isoMod, cfg)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return out
}
