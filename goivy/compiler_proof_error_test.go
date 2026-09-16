package goivy

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

type pythonCheckPropertiesProofErrorResult struct {
	Errored bool   `json:"errored"`
	Type    string `json:"type"`
	Message string `json:"message"`
}

func TestCheckPropertiesPropagatesProofAdmissionErrorsLikePython(t *testing.T) {
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import json
from ivy import ivy_ast as ia
from ivy import ivy_compiler as ic
from ivy import ivy_logic as il
from ivy import ivy_module as im

m = im.Module()
prop = ia.LabeledFormula(ia.Atom('badprop'), il.And())
proof = ia.SchemaInstantiation(ia.Atom('missing'), None)
m.labeled_props = [prop]
m.proofs = [(prop, proof)]

try:
    with m:
        ic.check_properties(m)
    res = {"errored": False, "type": "", "message": ""}
except Exception as err:
    res = {"errored": True, "type": type(err).__name__, "message": str(err)}

print(json.dumps(res, sort_keys=True))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python check_properties proof-error oracle failed: %v\n%s", err, out)
	}
	var want pythonCheckPropertiesProofErrorResult
	if err := json.Unmarshal(out, &want); err != nil {
		t.Fatalf("decode python check_properties proof-error oracle %q: %v", out, err)
	}
	if !want.Errored || want.Type != "ProofError" || !strings.Contains(want.Message, "No property missing") {
		t.Fatalf("python check_properties proof-error oracle changed: %+v", want)
	}

	mod := New()
	cfg := mod.Cfg.AstCfg
	prop := cfg.NewLabeledFormula(cfg.NewAtom("badprop"), True)
	proof := cfg.NewSchemaInstantiation(cfg.NewAtom("missing"), nil)
	mod.LabeledProps = []*LabeledFormula{prop}
	mod.Proofs = []ProofEntry{{Formula: prop, Proof: proof}}

	err = CheckProperties(mod)
	if err == nil {
		t.Fatalf("CheckProperties swallowed proof admission error; Python raised %s: %s", want.Type, want.Message)
	}
	if !strings.Contains(err.Error(), "No property missing") {
		t.Fatalf("CheckProperties propagated wrong proof error\nwant substring %q\ngot: %v", "No property missing", err)
	}
}

func TestCheckPropertiesGivesEmptyProofToSchemaBodyTheorems(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg
	prop := cfg.NewLabeledFormula(cfg.NewAtom("schema"), cfg.NewSchemaBody(True))
	mod.LabeledProps = []*LabeledFormula{prop}

	out := captureActionUpdateStdout(t, func() {
		if err := CheckProperties(mod); err != nil {
			t.Fatalf("CheckProperties: %v", err)
		}
	})

	wantID := fmt.Sprintf("compiler.CheckProperties.classify label=schema id=%d", prop.ID)
	if !strings.Contains(out, wantID) || !strings.Contains(out, "hasPf=True") {
		t.Fatalf("SchemaBody theorem should receive an empty proof before classification.\nwant trace containing: %s ... hasPf=True\ngot:\n%s", wantID, out)
	}
	if _, found := mod.Schemata.Get2("schema"); !found {
		t.Fatalf("proved SchemaBody theorem should be registered as a schema")
	}
}

func TestCheckPropertiesRelabelsGeneratedSubgoalsWithPreservingClone(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg
	prem := cfg.NewLabeledFormula(cfg.NewAtom("prem"), True)
	prop := cfg.NewLabeledFormula(cfg.NewAtom("schema"), cfg.NewSchemaBody(prem, True))
	mod.LabeledProps = []*LabeledFormula{prop}

	out := captureActionUpdateStdout(t, func() {
		if err := CheckProperties(mod); err != nil {
			t.Fatalf("CheckProperties: %v", err)
		}
	})

	if !strings.Contains(out, "compiler.CheckProperties.classify label=schema -> props (proved, 1 subgoals)") {
		t.Fatalf("expected schema theorem to produce one generated subgoal; got:\n%s", out)
	}
	if !strings.Contains(out, "ast.LF.clone PRESERVE") {
		t.Fatalf("generated subgoal relabeling should preserve the theorem_to_property LF id; got:\n%s", out)
	}
	if len(mod.LabeledProps) == 0 || mod.LabeledProps[0].ID >= cfg.LfCounter {
		t.Fatalf("relabelled generated subgoal should preserve an existing LF id; id=%d counter=%d", mod.LabeledProps[0].ID, cfg.LfCounter)
	}
}

func TestCheckPropertiesNamedTransDropsOneElementAndLikePython(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg
	s := &UninterpretedSort{Name: "node"}
	mod.Sig.AddSort(s)
	pSort, err := NewFunctionSort(s, Boolean)
	if err != nil {
		t.Fatalf("NewFunctionSort: %v", err)
	}
	p := NewConst("p", pSort)
	a := NewConst("a", s)
	mod.Sig.AddSymbol(p.Name, p.CSort)
	mod.Sig.AddSymbol(a.Name, a.CSort)

	x, err := NewVariable("X", s)
	if err != nil {
		t.Fatalf("NewVariable: %v", err)
	}
	px, err := NewApply(p, x)
	if err != nil {
		t.Fatalf("NewApply(p, X): %v", err)
	}
	formula := &LogicAnd{Terms: []Expr{&LogicExists{Variables: []*LogicVariable{x}, Body: px}}}
	prop := cfg.NewLabeledFormula(cfg.NewAtom("named_prop"), formula)
	mod.LabeledProps = []*LabeledFormula{prop}
	mod.Named = []NamedEntry{{Formula: prop, Name: a}}

	if err := CheckProperties(mod); err != nil {
		t.Fatalf("CheckProperties: %v", err)
	}
	if len(mod.LabeledProps) < 2 {
		t.Fatalf("CheckProperties should add original and named-transformed props, got %d", len(mod.LabeledProps))
	}
	got, ok := mod.LabeledProps[1].Formula.(*Apply)
	if !ok {
		t.Fatalf("named transformed formula = %T, want *Apply p(a)", mod.LabeledProps[1].Formula)
	}
	if got.Func != p || len(got.Terms) != 1 || got.Terms[0] != a {
		t.Fatalf("named transformed formula = %#v, want p(a)", got)
	}
}
