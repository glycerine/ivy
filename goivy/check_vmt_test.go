package goivy

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type pythonVMTActionToTROracle struct {
	StateVars  []string `json:"stvars"`
	TransNames []string `json:"trans_names"`
	ErrorNames []string `json:"error_names"`
}

func vmtExprSymbolNames(e Expr) []string {
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

func TestVMTActionToTRUsesRealActionUpdateLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_actions as act, ivy_logic as il, ivy_logic_utils as lu, ivy_module, ivy_vmt
from ivy.logic import Boolean

m = ivy_module.Module()
ivy_module.module = m
p = il.Symbol("p", Boolean)
action = act.AssignAction(p, il.And())
stvars, trans, error = ivy_vmt.action_to_tr(m, action, "")

print(json.dumps({
    "stvars": [s.name for s in stvars],
    "trans_names": sorted(s.name for s in lu.used_symbols_clauses(trans)),
    "error_names": sorted(s.name for s in lu.used_symbols_clauses(error)),
}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python ivy_vmt.action_to_tr oracle failed: %v\n%s", err, out)
	}
	var want pythonVMTActionToTROracle
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python ivy_vmt.action_to_tr oracle %q: %v", out, err)
	}

	mod := New()
	p := NewConst("p", Boolean)
	action := NewAssignAction(p, True)
	gotStateVars, gotTrans, gotError, err := actionToTR(mod, action, "")
	if err != nil {
		t.Fatalf("actionToTR failed: %v", err)
	}

	if !reflect.DeepEqual(gotStateVars, want.StateVars) {
		t.Fatalf("state vars differ from Python\nwant: %#v\ngot:  %#v", want.StateVars, gotStateVars)
	}
	if names := vmtExprSymbolNames(gotTrans); !reflect.DeepEqual(names, want.TransNames) {
		t.Fatalf("transition symbol names differ from Python\nwant: %#v\ngot:  %#v", want.TransNames, names)
	}
	if names := vmtExprSymbolNames(gotError); !reflect.DeepEqual(names, want.ErrorNames) {
		t.Fatalf("error symbol names differ from Python\nwant: %#v\ngot:  %#v", want.ErrorNames, names)
	}
}

func TestVMTCheckIsolateAppliesConjectureProofTacticsLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_ast, ivy_logic as il, ivy_module, ivy_proof, ivy_tactics
from ivy.logic import Boolean

m = ivy_module.Module()
p = il.Symbol("p", Boolean)
lf = ivy_ast.LabeledFormula(ivy_ast.Atom("c"), p)
proof = ivy_ast.TacticTactic(ivy_ast.Atom("sorry"), ivy_ast.NoneAST(), ivy_ast.NoneAST())
pc = ivy_proof.ProofChecker(m.labeled_axioms, m.definitions, m.schemata)
subgoals = pc.admit_proposition(lf, proof)

print(json.dumps({"nconjs": len(subgoals)}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python proof oracle failed: %v\n%s", err, out)
	}
	var want struct {
		NConjs int `json:"nconjs"`
	}
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python proof oracle %q: %v", out, err)
	}

	mod := New()
	RegisterTactics(mod.Cfg.ProofCfg, mod)
	acfg := mod.Cfg.AstCfg
	p := NewConst("p", Boolean)
	lf := acfg.NewLabeledFormula(NewConst("c", Boolean), p)
	proof := acfg.NewTacticTactic(acfg.NewAtom("sorry"), acfg.NewNoneAST(), acfg.NewNoneAST())
	mod.LabeledConjs = []*LabeledFormula{lf}
	mod.Proofs = []ProofEntry{{Formula: lf, Proof: proof}}

	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmp := t.TempDir()
	if err := os.Chdir(tmp); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(origWD); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}()

	if err := VMTCheckIsolate("mc", mod); err != nil {
		t.Fatalf("VMTCheckIsolate failed: %v", err)
	}
	vmt, err := os.ReadFile("ivy.vmt")
	if err != nil {
		t.Fatalf("read ivy.vmt: %v", err)
	}
	if got := strings.Count(string(vmt), ":invar-property"); got != want.NConjs {
		t.Fatalf("VMT invariant property count differs from Python proof-tactic result\nwant: %d\ngot:  %d\n%s", want.NConjs, got, vmt)
	}
}
