package goivy

import (
	"encoding/json"
	"reflect"
	"sort"
	"testing"
)

type pythonLocalDecomposeOracle struct {
	PreNames  []string `json:"pre_names"`
	PostNames []string `json:"post_names"`
	Acts      []string `json:"acts"`
}

func decomposeExprSymbolNames(e Expr) []string {
	if e == nil {
		return nil
	}
	seen := UsedSymbolsAst(e)
	names := make([]string, 0, seen.Len())
	for _, sym := range seen.All() {
		names = append(names, ExprName(sym))
	}
	sort.Strings(names)
	return names
}

func decomposeUpdateSymbolNames(u *Update) []string {
	if u == nil || u.TR == nil {
		return nil
	}
	seen := UsedSymbolsClauses(u.TR)
	names := make([]string, 0, seen.Len())
	for _, sym := range seen.All() {
		names = append(names, ExprName(sym))
	}
	sort.Strings(names)
	return names
}

func TestDecomposeWithStateLocalActionHidesStateLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_logic as il, ivy_logic_utils as lu, ivy_transrel as tr
from ivy.logic import Boolean

l = il.Symbol("l", Boolean)
pre = tr.pure_state(tr.formula_to_clauses(l))
post = tr.pure_state(tr.formula_to_clauses(l))
action = act.LocalAction(l, act.Sequence(act.AssumeAction(l)))
comp = action.decompose(pre, post)[0]

print(json.dumps({
    "pre_names": sorted(s.name for s in lu.used_symbols_clauses(comp[0][1])),
    "post_names": sorted(s.name for s in lu.used_symbols_clauses(comp[2][1])),
    "acts": [type(a).__name__ for a in comp[1]],
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python LocalAction.decompose oracle failed: %v\n%s", err, out)
	}
	var want pythonLocalDecomposeOracle
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python LocalAction.decompose oracle %q: %v", out, err)
	}

	cfg := NewActionsConfig()
	local := NewConst("l", Boolean)
	action := NewLocalActionOn(cfg, "test.local.decompose", local, NewSequence(NewAssumeAction(local)))
	got := DecomposeWithState(action, local, local, false)
	if len(got) != 1 {
		t.Fatalf("DecomposeWithState returned %d triples, want 1", len(got))
	}
	if len(got[0].Actions) != 1 || ActionTypeName(got[0].Actions[0]) != "AssumeAction" {
		t.Fatalf("decomposed actions = %#v, want one AssumeAction", got[0].Actions)
	}

	if names := decomposeExprSymbolNames(got[0].Pre); !reflect.DeepEqual(names, want.PreNames) {
		t.Fatalf("pre-state symbol names differ from Python\nwant: %#v\ngot:  %#v", want.PreNames, names)
	}
	if names := decomposeExprSymbolNames(got[0].Post); !reflect.DeepEqual(names, want.PostNames) {
		t.Fatalf("post-state symbol names differ from Python\nwant: %#v\ngot:  %#v", want.PostNames, names)
	}
}

type pythonCallDecomposeOracle struct {
	PreNames   []string `json:"pre_names"`
	PostNames  []string `json:"post_names"`
	Acts       []string `json:"acts"`
	HasFormals []bool   `json:"has_formals"`
	HasReturns []bool   `json:"has_returns"`
}

func TestDecomposeWithUpdateCallActionConformsToPython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_ast, ivy_logic as il, ivy_logic_utils as lu, ivy_module, ivy_transrel as tr
from ivy.logic import Boolean

m = ivy_module.Module()
ivy_module.module = m
x = il.Symbol("x", Boolean)
y = il.Symbol("y", Boolean)
a = il.Symbol("a", Boolean)
r = il.Symbol("r", Boolean)
body = act.Sequence(act.AssumeAction(x))
body.formal_params = [x]
body.formal_returns = [y]
m.actions["foo"] = body
call = act.CallAction(ivy_ast.Atom("foo", a), r)
comp = call.decompose(tr.top_state(), tr.top_state())[0]

print(json.dumps({
    "pre_names": sorted(s.name for s in lu.used_symbols_clauses(comp[0][1])),
    "post_names": sorted(s.name for s in lu.used_symbols_clauses(comp[2][1])),
    "acts": [type(a).__name__ for a in comp[1]],
    "has_formals": [hasattr(a, "formal_params") for a in comp[1]],
    "has_returns": [hasattr(a, "formal_returns") for a in comp[1]],
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python CallAction.decompose oracle failed: %v\n%s", err, out)
	}
	var want pythonCallDecomposeOracle
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python CallAction.decompose oracle %q: %v", out, err)
	}

	cfg := NewActionsConfig()
	mod := NewModule()
	mod.Cfg.ActCfg = cfg
	x := NewConst("x", Boolean)
	y := NewConst("y", Boolean)
	actual := NewConst("a", Boolean)
	ret := NewConst("r", Boolean)
	body := NewSequence(NewAssumeAction(x))
	body.SetFormalParams([]*Const{x})
	body.SetFormalReturns([]*Const{y})
	mod.Actions.Set("foo", body)
	ctx := &UpdateContext{
		Domain: mod,
		ActCfg: cfg,
		GetAction: func(name string) ActionsAction {
			if act, ok := mod.Actions.Get2(name); ok {
				return act
			}
			return nil
		},
	}

	call := NewCallActionOn(cfg, MustApply(NewConst("foo", TopS), actual), ret)
	got := DecomposeWithUpdate(ctx, call, TopState(), TopState(), false)
	if len(got) != 1 {
		t.Fatalf("DecomposeWithUpdate returned %d triples, want 1", len(got))
	}
	if acts := []string{ActionTypeName(got[0].Actions[0])}; !reflect.DeepEqual(acts, want.Acts) {
		t.Fatalf("decomposed action types differ from Python\nwant: %#v\ngot:  %#v", want.Acts, acts)
	}
	if names := decomposeUpdateSymbolNames(got[0].Pre); !reflect.DeepEqual(names, want.PreNames) {
		t.Fatalf("pre-state symbol names differ from Python\nwant: %#v\ngot:  %#v", want.PreNames, names)
	}
	if names := decomposeUpdateSymbolNames(got[0].Post); !reflect.DeepEqual(names, want.PostNames) {
		t.Fatalf("post-state symbol names differ from Python\nwant: %#v\ngot:  %#v", want.PostNames, names)
	}
	if len(got[0].Actions[0].GetFormalParams()) != 0 || len(got[0].Actions[0].GetFormalReturns()) != 0 {
		t.Fatalf("call decomposition callee retained formals; Python clone drops them")
	}
}

func TestDecomposeWithUpdateWhileActionExpandsLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_logic as il, ivy_module, ivy_transrel as tr
from ivy.logic import Boolean

m = ivy_module.Module()
ivy_module.module = m
p = il.Symbol("p", Boolean)
action = act.WhileAction(p, act.Sequence(act.AssumeAction(p)))
comp = action.decompose(tr.top_state(), tr.top_state())[0]
print(json.dumps({"acts": [type(a).__name__ for a in comp[1]]}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python WhileAction.decompose oracle failed: %v\n%s", err, out)
	}
	var want struct {
		Acts []string `json:"acts"`
	}
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python WhileAction.decompose oracle %q: %v", out, err)
	}

	cfg := NewActionsConfig()
	mod := NewModule()
	mod.Cfg.ActCfg = cfg
	ctx := &UpdateContext{Domain: mod, ActCfg: cfg}
	p := NewConst("p", Boolean)
	action := NewWhileAction(p, NewSequence(NewAssumeAction(p)))

	got := DecomposeWithUpdate(ctx, action, TopState(), TopState(), false)
	if len(got) != 1 {
		t.Fatalf("DecomposeWithUpdate returned %d triples, want 1", len(got))
	}
	acts := make([]string, len(got[0].Actions))
	for i, act := range got[0].Actions {
		acts[i] = ActionTypeName(act)
	}
	if !reflect.DeepEqual(acts, want.Acts) {
		t.Fatalf("while decomposition action types differ from Python\nwant: %#v\ngot:  %#v", want.Acts, acts)
	}
}
