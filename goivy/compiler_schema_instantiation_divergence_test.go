package goivy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const topLevelSchemaInstantiationSource = `#lang ivy1.7

type t
individual a : t
axiom a = a
schema reflex = a = a
instantiate reflex
`

const topLevelSchemaAxiomCollectionSource = `#lang ivy1.7

type t
individual a : t
relation p(X:t)
schema adds_p = p(a)
instantiate adds_p
`

const actionSchemaInstantiationSource = `#lang ivy1.7

type t
relation p(X:t)
individual a : t
individual b : t
schema use_p(X) = p(X)
action act = {
    instantiate use_p(a)
}
`

type pythonSchemaInstantiationResult struct {
	LabeledAxioms        int                 `json:"labeled_axioms"`
	ModuleInstantiations int                 `json:"module_instantiations"`
	SchemaInstances      map[string]int      `json:"schema_instances"`
	SchemaInstanceTexts  map[string][]string `json:"schema_instance_texts"`
}

type pythonActionInstantiationResult struct {
	TRFmlas []string `json:"tr_fmlas"`
}

func TestTopLevelSchemaInstantiationMatchesPython(t *testing.T) {
	py := runPythonTopLevelSchemaInstantiation(t)

	result, err := Parse(topLevelSchemaInstantiationSource, Version{1, 7})
	if err != nil {
		t.Fatalf("parse Go Ivy source: %v", err)
	}

	mod := New()
	mod.Cfg = NewConfig()
	if err := CheckInstantiations(mod, result.Decls); err != nil {
		t.Fatalf("Go CheckInstantiations: %v", err)
	}

	c := NewFromModule(mod)
	c.TopCtx = CollectActions(result.Decls)
	domain := NewDomainSetup(c)
	if err := domain.ProcessDecls(result.Decls); err != nil {
		t.Fatalf("Go DomainSetup: %v", err)
	}

	schemaNode, ok := mod.Schemata.Get2("reflex")
	if !ok {
		t.Fatalf("Go did not register schema reflex")
	}
	goSchema, ok := schemaNode.(*AstSchema)
	if !ok {
		t.Fatalf("Go schema reflex has type %T, want *ast.Schema", schemaNode)
	}

	goInstanceTexts := make([]string, len(goSchema.Instances))
	for i, inst := range goSchema.Instances {
		goInstanceTexts[i] = fmt.Sprint(inst)
	}
	if got, want := strings.Join(goInstanceTexts, "\n"), strings.Join(py.SchemaInstanceTexts["reflex"], "\n"); got != want {
		t.Fatalf("top-level schema instantiation mismatch for schema %q:\nGo instances:\n%s\nPython instances:\n%s\nGo module instantiations=%d, Python module instantiations=%d",
			"reflex", got, want, len(mod.Instantiations), py.ModuleInstantiations)
	}
}

func TestTopLevelSchemaInstancesAppearInGetAxioms(t *testing.T) {
	result, err := Parse(topLevelSchemaAxiomCollectionSource, Version{1, 7})
	if err != nil {
		t.Fatalf("parse Go Ivy source: %v", err)
	}

	mod := New()
	mod.Cfg = NewConfig()
	if err := CheckInstantiations(mod, result.Decls); err != nil {
		t.Fatalf("Go CheckInstantiations: %v", err)
	}

	c := NewFromModule(mod)
	c.TopCtx = CollectActions(result.Decls)
	if err := NewDomainSetup(c).ProcessDecls(result.Decls); err != nil {
		t.Fatalf("Go DomainSetup: %v", err)
	}

	axioms := mod.GetAxioms()
	var got []string
	for _, ax := range axioms {
		got = append(got, fmt.Sprint(ax))
	}
	if !containsString(got, "p(a)") {
		t.Fatalf("Module.GetAxioms() did not include top-level schema instance p(a); got %v", got)
	}
}

func TestActionLevelSchemaInstantiationMatchesPython(t *testing.T) {
	wantTRFmlas := []string{"p(X)"}

	result, err := Parse(actionSchemaInstantiationSource, Version{1, 7})
	if err != nil {
		t.Fatalf("parse Go Ivy source: %v", err)
	}

	mod := New()
	mod.Cfg = NewConfig()
	if err := CheckInstantiations(mod, result.Decls); err != nil {
		t.Fatalf("Go CheckInstantiations: %v", err)
	}

	c := NewFromModule(mod)
	c.TopCtx = CollectActions(result.Decls)
	if err := NewDomainSetup(c).ProcessDecls(result.Decls); err != nil {
		t.Fatalf("Go DomainSetup: %v", err)
	}

	var update *Update
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("Go action-level schema instantiation panicked; Python TR formulas=%v; panic=%v", wantTRFmlas, r)
			}
		}()
		inst := mod.Cfg.AstCfg.NewAtom("use_p", mod.Cfg.AstCfg.NewApp(mod.Cfg.AstCfg.NewSymbol("a", nil)))
		act := NewInstantiateAction(nil)
		act.AstInst = inst
		update = IntUpdate(act, &UpdateContext{
			Domain:                   mod,
			CompileWithSortInference: c.CompileWithSortInference,
		})
	}()
	if update == nil || update.TR == nil {
		t.Fatalf("Go action-level schema instantiation returned nil update/TR; Python TR formulas=%v", wantTRFmlas)
	}

	goTRFmlas := make([]string, len(update.TR.Fmlas))
	for i, fmla := range update.TR.Fmlas {
		goTRFmlas[i] = fmt.Sprint(fmla)
	}
	if got, want := strings.Join(goTRFmlas, "\n"), strings.Join(wantTRFmlas, "\n"); got != want {
		t.Fatalf("action-level schema instantiation mismatch:\nGo TR formulas:\n%s\nPython TR formulas:\n%s", got, want)
	}
}

func runPythonTopLevelSchemaInstantiation(t *testing.T) pythonSchemaInstantiationResult {
	t.Helper()

	repoRoot := testRepoRoot(t)
	python := filepath.Join(repoRoot, "pyivy", "goivy-venv", "bin", "python3")
	if _, err := os.Stat(python); err != nil {
		if _, lookErr := exec.LookPath("python3"); lookErr != nil {
			t.Skip("python3 not available")
		}
		python = "python3"
	}

	script := `
import io
import json
import sys

from ivy import ivy_compiler as ic
from ivy import ivy_module as im

src = ` + pythonTripleQuote(topLevelSchemaInstantiationSource) + `

with im.Module():
    decls = ic.read_module(io.StringIO(src))
    ic.check_instantiations(im.module, decls)
    ic.IvyDomainSetup(im.module)(decls)
    schema_instances = {
        name: len(getattr(schema, "instances", []))
        for name, schema in im.module.schemata.items()
    }
    schema_instance_texts = {
        name: [str(instance) for instance in getattr(schema, "instances", [])]
        for name, schema in im.module.schemata.items()
    }
    print("PYRESULT " + json.dumps({
        "labeled_axioms": len(im.module.labeled_axioms),
        "module_instantiations": len(im.module.instantiations),
        "schema_instances": schema_instances,
        "schema_instance_texts": schema_instance_texts,
    }, sort_keys=True))
`

	cmd := exec.Command(python, "-c", script)
	cmd.Dir = filepath.Join(repoRoot, "pyivy", "ivy")
	cmd.Env = append(os.Environ(), "PYTHONPATH="+filepath.Join(repoRoot, "pyivy", "ivy"))
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python schema oracle failed: %v\n%s", err, out)
	}

	const marker = "PYRESULT "
	idx := bytes.LastIndex(out, []byte(marker))
	if idx < 0 {
		t.Fatalf("python schema oracle did not emit %q; output:\n%s", marker, out)
	}
	line := string(out[idx+len(marker):])
	if nl := strings.IndexByte(line, '\n'); nl >= 0 {
		line = line[:nl]
	}

	var res pythonSchemaInstantiationResult
	if err := json.Unmarshal([]byte(line), &res); err != nil {
		t.Fatalf("decode python schema oracle result %q: %v", line, err)
	}
	if res.SchemaInstances["reflex"] != 1 {
		t.Fatalf("python oracle sanity check failed: reflex instances=%d, want 1", res.SchemaInstances["reflex"])
	}
	if len(res.SchemaInstanceTexts["reflex"]) != 1 || res.SchemaInstanceTexts["reflex"][0] == "" {
		t.Fatalf("python oracle sanity check failed: reflex instance texts=%v", res.SchemaInstanceTexts["reflex"])
	}
	return res
}

func containsString(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

func testRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	root, err := filepath.Abs(filepath.Join(wd, ".."))
	if err != nil {
		t.Fatalf("resolve repo root: %v", err)
	}
	return root
}

func pythonTripleQuote(s string) string {
	return `"""` + strings.ReplaceAll(s, `"""`, `\"\"\"`) + `"""`
}
