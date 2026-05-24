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
