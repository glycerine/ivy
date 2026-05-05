package goivy

import (
	"encoding/json"
	"strings"
	"testing"
)

type pythonARGSetupActionErrorResult struct {
	Errored bool     `json:"errored"`
	Type    string   `json:"type"`
	Message string   `json:"message"`
	Actions []string `json:"actions"`
}

func TestARGSetupActionCompileErrorsDoNotRegisterFallbackLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act
from ivy import ivy_ast as ia
from ivy import ivy_compiler as ic
from ivy import ivy_module as im
from ivy import ivy_utils as iu

m = im.Module()
loc = iu.LocationTuple(('test.ivy', 1))
body = act.CallAction(ia.Atom('missing'))
body.lineno = loc
adef = ia.ActionDef(ia.Atom('bad'), body)
adef.lineno = loc

try:
    with m:
        with ic.TopContext({}):
            ic.IvyARGSetup(m).action(adef)
    res = {"errored": False, "type": "", "message": "", "actions": sorted(m.actions.keys())}
except Exception as err:
    res = {"errored": True, "type": type(err).__name__, "message": str(err), "actions": sorted(m.actions.keys())}

print(json.dumps(res, sort_keys=True))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python ARGSetup action-error oracle failed: %v\n%s", err, out)
	}
	var want pythonARGSetupActionErrorResult
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python ARGSetup action-error oracle %q: %v", out, err)
	}
	if !want.Errored || want.Type != "IvyError" || !strings.Contains(want.Message, "unknown symbol: missing") || len(want.Actions) != 0 {
		t.Fatalf("python ARGSetup action-error oracle changed: %+v", want)
	}

	cfg := NewAstConfig()
	c := newTestCompiler()
	body := cfg.NewCallAction(cfg.NewAtom("missing"))
	body.SetLineno(Location{Filename: "test.ivy", Line: 1})
	adef := cfg.NewActionDef(cfg.NewAtom("bad"), body, nil, nil)
	adef.SetLineno(Location{Filename: "test.ivy", Line: 1})
	decl := cfg.NewActionDecl(adef)

	c.TopCtx = &TopContext{Actions: map[string]*ActionInfo{}}
	err = NewARGSetup(c).ProcessDecls([]Node{decl})
	if err == nil {
		t.Fatal("ARGSetup.ProcessDecls swallowed action compile error; Python raised IvyError")
	}
	if !strings.Contains(err.Error(), "unknown symbol: missing") {
		t.Fatalf("ARGSetup.ProcessDecls propagated wrong error\nwant substring %q\ngot: %v", "unknown symbol: missing", err)
	}
	if _, ok := c.Module.Actions.Get2("bad"); ok {
		t.Fatal("ARGSetup registered fallback action bad after compile failure; Python leaves module.actions empty")
	}
}
