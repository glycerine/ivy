package goivy

import (
	"encoding/json"
	"os"
	"path/filepath"
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

func TestVMTExportFramesIndependentActions(t *testing.T) {
	const source = `#lang ivy1.8
type item
individual sample : item
relation stable(X:item)
axiom [fixed] stable(sample)
relation p
relation q
after init { p := false; q := false }
export action set_p = { p := true }
export action set_q = { q := true }
invariant [p_safe] p | ~p
invariant [q_safe] q | ~q
`
	mod := New()
	if err := SourceString("framed.ivy", source, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	if err := CreateIsolate("this", mod); err != nil {
		t.Fatal(err)
	}
	exitTheory := mod.TheoryContext()
	defer exitTheory()
	mod.Name = filepath.Join(t.TempDir(), "framed")
	if err := VMTCheckIsolate("vmt", mod); err != nil {
		t.Fatal(err)
	}
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
from ivy import z3
from pathlib import Path
import sys

model = Path(sys.argv[1]).read_text()
for constraint, want in [
    ("(and $trans (not p) (not q) new_p (not new_q))", z3.sat),
    ("(and $trans (not p) (not q) new_p new_q)", z3.unsat),
    ("(and $init (or p q))", z3.unsat),
    ("(not (stable sample))", z3.unsat),
]:
    solver = z3.Solver()
    solver.add(z3.parse_smt2_string(model + "\n(assert " + constraint + ")"))
    got = solver.check()
    if got != want:
        raise AssertionError((constraint, str(got), str(want)))
`, mod.Name+".vmt")
	if out, err := cmd.CombinedOutput(); err != nil {
		data, _ := os.ReadFile(mod.Name + ".vmt")
		t.Fatalf("VMT transition validation failed: %v\n%s\n%s", err, out, data)
	}
}

func TestPyIvyVMTExportUsesDeclaredInitBeforeInitializer(t *testing.T) {
	const source = `#lang ivy1.6
relation p
relation q
init p
after init { q := p; p := false }
action step = { p := true; q := false }
export step
`
	path := exportPyIvyVMT(t, source)
	validate := pythonIvyCommandForTest(t, "-O", "-c", `
from pathlib import Path
import sys
from ivy import z3

vmt = Path(sys.argv[1]).read_text()
for condition, expected in [('init', z3.sat), ('(and init (or p (not q)))', z3.unsat)]:
    solver = z3.Solver()
    solver.add(z3.parse_smt2_string(vmt + "\n(assert " + condition + ")"))
    actual = solver.check()
    if actual != expected:
        raise AssertionError((condition, actual, expected, vmt))
`, path)
	if out, err := validate.CombinedOutput(); err != nil {
		t.Fatalf("check PyIvy VMT initial states: %v\n%s", err, out)
	}
}

func TestPyIvyVMTExportKeepsStateAndMetadataNamesDistinct(t *testing.T) {
	const source = `#lang ivy1.8
relation trans
relation property_0
after init { trans := false; property_0 := true }
export action step = { trans := true; property_0 := false }
invariant [observe] ~trans & property_0
`
	path := exportPyIvyVMT(t, source)
	validate := pythonIvyCommandForTest(t, "-O", "-c", `
from pathlib import Path
import sys
from ivy import z3

vmt = Path(sys.argv[1]).read_text()
solver = z3.Solver()
solver.add(z3.parse_smt2_string(vmt + "\n(assert (and init (or trans (not property_0))))"))
if solver.check() != z3.unsat:
    raise AssertionError("source state names were not preserved:\n" + vmt)
`, path)
	if out, err := validate.CombinedOutput(); err != nil {
		t.Fatalf("check PyIvy VMT names: %v\n%s", err, out)
	}
}

func TestVMTExportDeclaresSortUsedOnlyByEnumConstructors(t *testing.T) {
	const source = `#lang ivy1.8
type color = {red, green}
invariant [colors_differ] red ~= green
`
	for _, engine := range []string{"GoIvy", "PyIvy"} {
		t.Run(engine, func(t *testing.T) {
			var path, property string
			switch engine {
			case "GoIvy":
				path, property = exportGoIvyVMT(t, source), "$property.0"
			case "PyIvy":
				path, property = exportPyIvyVMT(t, source), "property_0"
			}
			validate := pythonIvyCommandForTest(t, "-O", "-c", `
from pathlib import Path
import sys
from ivy import z3

model = Path(sys.argv[1]).read_text()
solver = z3.Solver()
solver.add(z3.parse_smt2_string(model + "\n(assert (not " + sys.argv[2] + "))"))
if solver.check() != z3.unsat:
    raise AssertionError("different enum constructors must be distinct:\n" + model)
`, path, property)
			if out, err := validate.CombinedOutput(); err != nil {
				t.Fatalf("check VMT enumeration: %v\n%s", err, out)
			}
		})
	}
}

func TestVMTExportHonorsWireAxiomInPreState(t *testing.T) {
	const source = `#lang ivy1.8
wire p : bool
axiom [valid_state] p
export action step = { }
invariant [safe] p
`
	path := exportGoIvyVMT(t, source)
	validate := pythonIvyCommandForTest(t, "-O", "-c", `
from pathlib import Path
import sys
from ivy import z3

model = Path(sys.argv[1]).read_text()
conditions = [
    ('$init', z3.sat),
    ('(and $init (not p))', z3.unsat),
    ('(and $trans p new_p)', z3.sat),
    ('(and $trans (not p) new_p)', z3.unsat),
]
for condition, expected in conditions:
    solver = z3.Solver()
    solver.add(z3.parse_smt2_string(model + "\n(assert " + condition + ")"))
    actual = solver.check()
    if actual != expected:
        raise AssertionError((condition, actual, expected, model))
`, path)
	if out, err := validate.CombinedOutput(); err != nil {
		t.Fatalf("check VMT state axioms: %v\n%s", err, out)
	}
}

func TestPyIvyVMTExportDeclaresSortUsedOnlyByQuantifier(t *testing.T) {
	const source = `#lang ivy1.8
type token
invariant [reflexive] forall V:token. V = V
`
	path := exportPyIvyVMT(t, source)
	validate := pythonIvyCommandForTest(t, "-O", "-c", `
from pathlib import Path
import sys
from ivy import z3

model = Path(sys.argv[1]).read_text()
solver = z3.Solver()
solver.add(z3.parse_smt2_string(model + "\n(assert (not property_0))"))
if solver.check() != z3.unsat:
    raise AssertionError("expected quantified reflexivity to hold:\n" + model)
`, path)
	if out, err := validate.CombinedOutput(); err != nil {
		t.Fatalf("check VMT quantifier sort: %v\n%s", err, out)
	}
}

func exportGoIvyVMT(t *testing.T, source string) string {
	t.Helper()
	mod := New()
	if err := SourceString("model.ivy", source, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	if err := CreateIsolate("this", mod); err != nil {
		t.Fatal(err)
	}
	endTheory := mod.TheoryContext()
	defer endTheory()
	mod.Name = filepath.Join(t.TempDir(), "model")
	if err := VMTCheckIsolate("vmt", mod); err != nil {
		t.Fatal(err)
	}
	return mod.Name + ".vmt"
}

func exportPyIvyVMT(t *testing.T, source string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "initial.ivy"), []byte(source), 0o600); err != nil {
		t.Fatal(err)
	}
	export := pythonIvyCommandForTest(t, "-O", "-c", `
from ivy import ivy_init, ivy_isolate, ivy_module, ivy_vmt

with ivy_module.Module():
    with open("initial.ivy") as source:
        ivy_init.source_file("initial.ivy", source, create_isolate=False)
    ivy_isolate.create_isolate(None)
    with ivy_module.module.theory_context():
        ivy_vmt.check_isolate()
`)
	export.Dir = dir
	if out, err := export.CombinedOutput(); err != nil {
		t.Fatalf("export PyIvy VMT: %v\n%s", err, out)
	}
	return filepath.Join(dir, "initial.vmt")
}

func TestVMTExportIndexedUpdates(t *testing.T) {
	const source = `#lang ivy1.8
type index
interpret index -> bv[4]
function live(X:index) : index
function saved(X:index) : index
after init { live(X) := X + 3; saved(X) := 0 }
export action save = { saved(X) := live(X) }
export action step(x:index) = { live(x) := 7 }
invariant [same] live(0) = saved(0) | live(0) ~= saved(0)
`
	mod := New()
	if err := SourceString("vmt_array.ivy", source, mod, mod.Sig, map[string]interface{}{"create_isolate": false}); err != nil {
		t.Fatal(err)
	}
	if err := CreateIsolate("this", mod); err != nil {
		t.Fatal(err)
	}
	leaveTheory := mod.TheoryContext()
	defer leaveTheory()
	mod.Name = filepath.Join(t.TempDir(), "vmt_array")
	if err := VMTCheckIsolate("vmt", mod); err != nil {
		t.Fatal(err)
	}
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
from ivy import z3
from pathlib import Path
import sys

model = Path(sys.argv[1]).read_text()
step = next(token for token in model.split() if token.startswith('|ivy.action.') and token.endswith('.step|'))
save = next(token for token in model.split() if token.startswith('|ivy.action.') and token.endswith('.save|'))
for constraint, expected in [
    ('(and $init (not (= (select live #x2) #x5)))', z3.unsat),
    ('(and $init (not (= (select saved #x2) #x0)))', z3.unsat),
    ('(and $trans ' + save + ' (not (= (select live #x2) (select new_saved #x2))))', z3.unsat),
    ('(and $trans ' + save + ' (not (= live new_live)))', z3.unsat),
    ('(and $trans ' + step + ' (= |__fml:x| #x2) (not (= (select new_live #x2) #x7)))', z3.unsat),
    ('(and $trans ' + step + ' (= |__fml:x| #x2) (not (= (select new_live #x3) (select live #x3))))', z3.unsat),
    ('(and $init $trans ' + save + ')', z3.sat),
    ('(and $init $trans ' + step + ')', z3.sat),
]:
    solver = z3.Then('simplify', 'qe', 'smt').solver()
    solver.add(z3.parse_smt2_string(model + '\n(assert ' + constraint + ')'))
    got = solver.check()
    if got != expected:
        raise AssertionError((constraint, got, expected))
`, mod.Name+".vmt")
	if out, err := cmd.CombinedOutput(); err != nil {
		model, _ := os.ReadFile(mod.Name + ".vmt")
		t.Fatalf("indexed VMT export failed: %v\n%s\n%s", err, out, model)
	}
}
