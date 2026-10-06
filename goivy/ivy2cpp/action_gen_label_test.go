package ivy2cpp

import (
	"github.com/glycerine/ivy/goivy"
	"testing"
)

func TestActionGenPreconditionFailureLabelPreservesSourceLocationsFast(t *testing.T) {
	g := &Generator{Mod: goivy.New()}
	act := goivy.NewSequence()
	act.SetLineno(goivy.Location{Filename: "probe.ivy", Line: 10})
	plan := &actionGenPlan{name: "ext:step", act: act, origAct: act}
	if got := g.actionGenPreconditionFailureLabel(plan); got != "probe.ivy: line 10" {
		t.Fatalf("unlocated precondition label = %q, want action source location", got)
	}
	plan.act = nil
	if got := g.actionGenPreconditionFailureLabel(plan); got != "probe.ivy: line 10" {
		t.Fatalf("original action label = %q, want original source location", got)
	}
	plan.act = goivy.NewSequence()
	plan.origAct = nil
	if got := g.actionGenPreconditionFailureLabel(plan); got != "step" {
		t.Fatalf("unlocated action label = %q, want action name", got)
	}
	require := goivy.NewRequiresAction(goivy.NewConst("enabled", goivy.Boolean))
	require.SetLineno(goivy.Location{Filename: "probe.ivy", Line: 12})
	plan.act = goivy.NewSequence(require)
	if got := g.actionGenPreconditionFailureLabel(plan); got != "probe.ivy: line 12" {
		t.Fatalf("require label = %q, want require source location", got)
	}
	pre := goivy.NewConst("external", goivy.Boolean)
	pre.SetLineno(goivy.Location{Filename: "probe.ivy", Line: 5})
	g.Mod.ExtPreconds = map[string]goivy.Expr{plan.name: pre}
	if got := g.actionGenPreconditionFailureLabel(plan); got != "probe.ivy: line 5" {
		t.Fatalf("external precondition label = %q, want external source location", got)
	}
}
