package goivy

// Batch F — Tests for:
//   §6.2 #5:  CheckDefinitions — separate defs from props, check redefinition, detect cycles
//   §6.2 #7:  CreateConjActions — determine which actions must preserve each conjecture
//   §6.3 #29: ConjSetup pass 2 — implement real conjecture setup

import (
	"encoding/json"
	"strings"
	"testing"
)

// ============================================================================
// §6.2 #5: CheckDefinitions
// ============================================================================

// helper: make a lg.Definition wrapping symbol names, usable as LabeledFormula.Formula
func makeLogicDef(defSym string, rhsSyms ...string) *LogicDefinition {
	lhs := NewConst(defSym, Boolean)
	var rhs Expr
	if len(rhsSyms) == 0 {
		rhs = NewConst("true_const", Boolean)
	} else if len(rhsSyms) == 1 {
		rhs = NewConst(rhsSyms[0], Boolean)
	} else {
		// Build an Apply chain so UsedConstantsList can find all rhs symbols.
		// Use nested Eq nodes: (rhsSyms[0] = rhsSyms[1]) for simplicity.
		// For > 2, just use the first two — tests only need specific symbols visible.
		s0 := NewConst(rhsSyms[0], Boolean)
		s1 := NewConst(rhsSyms[1], Boolean)
		rhs = NewDefinition(s0, s1) // nested def as expression — carries both constants
	}
	return NewDefinition(lhs, rhs)
}

// helper: wrap a lg.Definition in a LabeledFormula with a label
func makeLabeledDef(cfg *AstConfig, label string, def *LogicDefinition) *LabeledFormula {
	return cfg.NewLabeledFormula(cfg.NewAtom(label), def)
}

// helper: wrap an arbitrary formula in a LabeledFormula
func makeLabeledFormula(cfg *AstConfig, label string, formula Node) *LabeledFormula {
	return cfg.NewLabeledFormula(cfg.NewAtom(label), formula)
}

// Test 1: Definitions without proofs move to mod.Definitions;
// non-definitions stay in LabeledProps; definitions with proofs stay in LabeledProps.
func TestCheckDefinitions_SeparatesDefsFromProps(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	// defA: definition of f (no proof) — should move to Definitions
	defA := makeLabeledDef(cfg, "defA", makeLogicDef("f"))

	// propB: non-definition property — should stay in LabeledProps
	propB := makeLabeledFormula(cfg, "propB", cfg.NewAtom("some_prop"))

	// defC: definition of g WITH a proof — should stay in LabeledProps (stale path)
	defC := makeLabeledDef(cfg, "defC", makeLogicDef("g"))
	mod.Proofs = append(mod.Proofs, ProofEntry{Formula: defC, Proof: cfg.NewAtom("proof_body")})

	mod.LabeledProps = []*LabeledFormula{defA, propB, defC}

	err := CheckDefinitions(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// defA should be in Definitions
	if len(mod.Definitions) != 1 {
		t.Errorf("expected 1 definition, got %d", len(mod.Definitions))
	} else if mod.Definitions[0] != defA {
		t.Errorf("expected defA in Definitions")
	}

	// propB and defC should remain in LabeledProps
	if len(mod.LabeledProps) != 2 {
		t.Errorf("expected 2 labeled props, got %d", len(mod.LabeledProps))
	}
}

// Test 2: Definition C uses symbol g which is stale (defined by B which has proof).
// C should NOT move to Definitions — it stays in LabeledProps.
func TestCheckDefinitions_StaleSymbols(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	// defA: f = true_const (no proof, no stale deps)
	defA := makeLabeledDef(cfg, "defA", makeLogicDef("f"))

	// defB: g = f (has proof → g becomes stale)
	defB := makeLabeledDef(cfg, "defB", makeLogicDef("g", "f"))
	mod.Proofs = append(mod.Proofs, ProofEntry{Formula: defB, Proof: cfg.NewAtom("pf")})

	// defC: h = g (g is stale → C should NOT move to Definitions)
	defC := makeLabeledDef(cfg, "defC", makeLogicDef("h", "g"))

	mod.LabeledProps = []*LabeledFormula{defA, defB, defC}

	err := CheckDefinitions(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// defA should move to Definitions (clean)
	found := false
	for _, d := range mod.Definitions {
		if d == defA {
			found = true
		}
	}
	if !found {
		t.Error("defA should be in Definitions")
	}

	// defC should stay in LabeledProps (uses stale g)
	foundC := false
	for _, p := range mod.LabeledProps {
		if p == defC {
			foundC = true
		}
	}
	if !foundC {
		t.Error("defC should remain in LabeledProps because it uses stale symbol g")
	}
}

// Test 3: Two definitions both define 'f' — should return error mentioning "redefinition".
func TestCheckDefinitions_RedefinitionError(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	def1 := makeLabeledDef(cfg, "def1", makeLogicDef("f"))
	def2 := makeLabeledDef(cfg, "def2", makeLogicDef("f"))

	mod.LabeledProps = []*LabeledFormula{def1, def2}

	err := CheckDefinitions(mod)
	if err == nil {
		t.Fatal("expected error for redefinition of 'f', got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "redefinition") {
		t.Errorf("error should mention 'redefinition', got: %s", err.Error())
	}
}

// Test 4: Definition in Definitions + NativeDefinitions with same symbol → error.
func TestCheckDefinitions_NativeDefinitionRedefinitionError(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	// Put one definition of 'f' as a LabeledProp (will move to Definitions)
	def1 := makeLabeledDef(cfg, "def1", makeLogicDef("f"))
	mod.LabeledProps = []*LabeledFormula{def1}

	// Put another definition of 'f' in NativeDefinitions
	nativeDef := makeLabeledDef(cfg, "native_f", makeLogicDef("f"))
	mod.NativeDefinitions = append(mod.NativeDefinitions, nativeDef)

	err := CheckDefinitions(mod)
	if err == nil {
		t.Fatal("expected error for redefinition of 'f' (native + regular), got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "redefinition") {
		t.Errorf("error should mention 'redefinition', got: %s", err.Error())
	}
}

// Test 5: Definition of 'f' and a Named entry with same symbol → error.
func TestCheckDefinitions_NamedRedefinitionError(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	def1 := makeLabeledDef(cfg, "def1", makeLogicDef("f"))
	mod.LabeledProps = []*LabeledFormula{def1}

	// Named entry with symbol 'f'
	fSym := NewConst("f", Boolean)
	namedLF := makeLabeledFormula(cfg, "named_f", cfg.NewAtom("something"))
	mod.Named = append(mod.Named, NamedEntry{Formula: namedLF, Name: fSym})

	err := CheckDefinitions(mod)
	if err == nil {
		t.Fatal("expected error for redefinition of 'f' (def + named), got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "redefinition") {
		t.Errorf("error should mention 'redefinition', got: %s", err.Error())
	}
}

// Test 6: Two definitions forming a cycle: f→g and g→f. Should return error.
func TestCheckDefinitions_CycleDetection(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	// f = g
	defF := makeLabeledDef(cfg, "defF", makeLogicDef("f", "g"))
	// g = f
	defG := makeLabeledDef(cfg, "defG", makeLogicDef("g", "f"))

	mod.LabeledProps = []*LabeledFormula{defF, defG}

	err := CheckDefinitions(mod)
	if err == nil {
		t.Fatal("expected error for dependency cycle between f and g, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "cycle") {
		t.Errorf("error should mention 'cycle', got: %s", err.Error())
	}
}

// Test 7: Self-recursive definition (f uses f) without proof → error "recursion schema".
func TestCheckDefinitions_SelfLoopRequiresProof(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	// f = f (self-loop)
	defF := makeLabeledDef(cfg, "defF", makeLogicDef("f", "f"))
	mod.LabeledProps = []*LabeledFormula{defF}

	err := CheckDefinitions(mod)
	if err == nil {
		t.Fatal("expected error 'requires a recursion schema' for self-recursive def without proof, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "recursion") {
		t.Errorf("error should mention 'recursion', got: %s", err.Error())
	}
}

// Test 8: Self-recursive definition WITH matching proof → should NOT error.
func TestCheckDefinitions_SelfLoopWithProofAccepted(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	// f = f (self-loop) but has a proof
	defF := makeLabeledDef(cfg, "defF", makeLogicDef("f", "f"))
	mod.Proofs = append(mod.Proofs, ProofEntry{Formula: defF, Proof: cfg.NewAtom("rec_proof")})
	mod.LabeledProps = []*LabeledFormula{defF}

	err := CheckDefinitions(mod)
	if err != nil {
		t.Fatalf("self-recursive def with proof should be accepted, got error: %v", err)
	}
	// The def should remain in LabeledProps (has proof → stale → stays)
	// OR be accepted via admit_definition. Either way, no error.
}

// Test 9: Action modifies symbol used in axiom → error for version >= 1.7.
func TestCheckDefinitions_ActionInterference_ModifiesAxiomSymbol(t *testing.T) {
	mod := New()
	oldVer := mod.Cfg.IuCfg.GetStringVersion()
	defer SetStringVersionOn(mod.Cfg.IuCfg, oldVer)
	SetStringVersionOn(mod.Cfg.IuCfg, "1.7")

	// Axiom uses symbol 'f' — use compiled lg.Const so structural keys match
	cfg := mod.Cfg.AstCfg
	fSym := NewConst("f", Boolean)
	axiomLF := makeLabeledFormula(cfg, "ax1", fSym)
	mod.LabeledAxioms = append(mod.LabeledAxioms, axiomLF)

	// Action that assigns to 'f' — uses real AssignAction so actions.Modifies finds it
	assignAction := NewAssignAction(fSym, NewConst("true_val", Boolean))
	mod.Actions.Set("act1", assignAction)
	err := CheckDefinitions(mod)
	if err == nil {
		t.Fatal("expected error for action modifying axiom symbol 'f', got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "immutable") ||
		!strings.Contains(strings.ToLower(err.Error()), "assigned") {
		t.Errorf("error should mention immutable symbol assigned, got: %s", err.Error())
	}
}

// Test 10: Action assigns to a defined symbol → error.
func TestCheckDefinitions_ActionInterference_ModifiesDefinedSymbol(t *testing.T) {
	mod := New()
	oldVer := mod.Cfg.IuCfg.GetStringVersion()
	defer SetStringVersionOn(mod.Cfg.IuCfg, oldVer)
	SetStringVersionOn(mod.Cfg.IuCfg, "1.7")

	cfg := mod.Cfg.AstCfg

	// Definition of 'f'
	defF := makeLabeledDef(cfg, "defF", makeLogicDef("f"))
	mod.LabeledProps = []*LabeledFormula{defF}

	// Action that assigns to 'f'
	fAssign := NewConst("f", Boolean)
	mod.Actions.Set("act1", NewAssignAction(fAssign, NewConst("true_val", Boolean)))

	err := CheckDefinitions(mod)
	if err == nil {
		t.Fatal("expected error for action modifying defined symbol 'f', got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "assigned") {
		t.Errorf("error should mention symbol being assigned, got: %s", err.Error())
	}
}

// Test 11: Clean definitions — no cycles, no redefinition, no stale. Should pass.
func TestCheckDefinitions_NoErrorOnCleanDefinitions(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	defA := makeLabeledDef(cfg, "defA", makeLogicDef("a"))
	defB := makeLabeledDef(cfg, "defB", makeLogicDef("b"))
	defC := makeLabeledDef(cfg, "defC", makeLogicDef("c"))

	mod.LabeledProps = []*LabeledFormula{defA, defB, defC}

	err := CheckDefinitions(mod)
	if err != nil {
		t.Fatalf("expected no error for clean definitions, got: %v", err)
	}
	if len(mod.Definitions) != 3 {
		t.Errorf("expected 3 definitions, got %d", len(mod.Definitions))
	}
}

// ============================================================================
// §6.2 #7: CreateConjActions
// ============================================================================

// Test 12: Version <= 1.6 → CreateConjActions should be a no-op.
func TestCreateConjActions_VersionGate(t *testing.T) {
	mod := New()
	oldVer := mod.Cfg.IuCfg.GetStringVersion()
	defer SetStringVersionOn(mod.Cfg.IuCfg, oldVer)
	SetStringVersionOn(mod.Cfg.IuCfg, "1.6")

	cfg := mod.Cfg.AstCfg

	// Add a conjecture and an export
	conjLF := makeLabeledFormula(cfg, "this.inv1", cfg.NewAtom("conj_body"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	expDef := cfg.NewExportDef(cfg.NewAtom("act1"), nil)
	mod.Exports = append(mod.Exports, expDef)

	CreateConjActions(mod)

	if len(mod.ConjActions) != 0 {
		t.Errorf("for version 1.6, ConjActions should be empty, got %d entries", len(mod.ConjActions))
	}
}

// Test 13: Top-level conjecture with no isolates → all exports.
func TestCreateConjActions_TopLevelConj_AllExports(t *testing.T) {
	mod := New()
	oldVer := mod.Cfg.IuCfg.GetStringVersion()
	defer SetStringVersionOn(mod.Cfg.IuCfg, oldVer)
	SetStringVersionOn(mod.Cfg.IuCfg, "1.7")

	cfg := mod.Cfg.AstCfg

	conjLF := makeLabeledFormula(cfg, "this.inv1", cfg.NewAtom("conj_body"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	exp1 := cfg.NewExportDef(cfg.NewAtom("act1"), nil)
	exp2 := cfg.NewExportDef(cfg.NewAtom("act2"), nil)
	mod.Exports = append(mod.Exports, exp1, exp2)

	CreateConjActions(mod)

	acts, ok := mod.ConjActions["this.inv1"]
	if !ok {
		t.Fatal("expected ConjActions entry for 'this.inv1'")
	}
	if len(acts) != 2 {
		t.Errorf("expected 2 actions for top-level conj, got %d", len(acts))
	}
}

// Test 14: Isolate-scoped conjecture → only that isolate's exports.
func TestCreateConjActions_IsolateScoping(t *testing.T) {
	mod := New()
	oldVer := mod.Cfg.IuCfg.GetStringVersion()
	defer SetStringVersionOn(mod.Cfg.IuCfg, oldVer)
	SetStringVersionOn(mod.Cfg.IuCfg, "1.7")

	cfg := mod.Cfg.AstCfg

	// Two conjectures in different objects
	conjA := makeLabeledFormula(cfg, "obj_a.inv", cfg.NewAtom("inv_a"))
	conjB := makeLabeledFormula(cfg, "obj_b.inv", cfg.NewAtom("inv_b"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjA, conjB)

	// Hierarchy: obj_a has child act1, obj_b has child act2.
	// IterIsolate walks verified names through mod.Hierarchy to find actions.
	hierA := NewInsMap[string, bool]()
	hierA.Set("act1", true)
	mod.Hierarchy.Set("obj_a", hierA)
	hierB := NewInsMap[string, bool]()
	hierB.Set("act2", true)
	mod.Hierarchy.Set("obj_b", hierB)

	// Actions registered with composed names (obj_a.act1, obj_b.act2)
	mod.Actions.Set("obj_a.act1", NewSequence()) // placeholder action body
	mod.Actions.Set("obj_b.act2", NewSequence())

	// Exports use the composed action names
	exp1 := cfg.NewExportDef(cfg.NewAtom("obj_a.act1"), nil)
	exp2 := cfg.NewExportDef(cfg.NewAtom("obj_b.act2"), nil)
	mod.Exports = append(mod.Exports, exp1, exp2)

	// Isolate iso_a verifies obj_a, iso_b verifies obj_b
	mod.Isolates["iso_a"] = cfg.NewIsolateDef(
		[]Node{cfg.NewAtom("iso_a"), cfg.NewAtom("obj_a")}, 0,
	)
	mod.Isolates["iso_b"] = cfg.NewIsolateDef(
		[]Node{cfg.NewAtom("iso_b"), cfg.NewAtom("obj_b")}, 0,
	)

	CreateConjActions(mod)

	actsA, ok := mod.ConjActions["obj_a.inv"]
	if !ok {
		t.Fatal("expected ConjActions entry for 'obj_a.inv'")
	}
	// obj_a.inv should only map to obj_a.act1 (from iso_a), not obj_b.act2
	if len(actsA) != 1 {
		t.Errorf("obj_a.inv should map to 1 action, got %d: %v", len(actsA), actsA)
	} else if actsA[0] != "obj_a.act1" {
		t.Errorf("obj_a.inv should map to obj_a.act1, got %s", actsA[0])
	}

	actsB, ok := mod.ConjActions["obj_b.inv"]
	if !ok {
		t.Fatal("expected ConjActions entry for 'obj_b.inv'")
	}
	if len(actsB) != 1 {
		t.Errorf("obj_b.inv should map to 1 action, got %d: %v", len(actsB), actsB)
	} else if actsB[0] != "obj_b.act2" {
		t.Errorf("obj_b.inv should map to obj_b.act2, got %s", actsB[0])
	}
}

// Test 15: Nested object walk — conj "obj.sub.inv" walks up to find "obj".
func TestCreateConjActions_NestedObjectWalk(t *testing.T) {
	mod := New()
	oldVer := mod.Cfg.IuCfg.GetStringVersion()
	defer SetStringVersionOn(mod.Cfg.IuCfg, oldVer)
	SetStringVersionOn(mod.Cfg.IuCfg, "1.7")

	cfg := mod.Cfg.AstCfg

	conjLF := makeLabeledFormula(cfg, "obj.sub.inv", cfg.NewAtom("nested_inv"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	// Hierarchy: obj has child act1
	hierObj := NewInsMap[string, bool]()
	hierObj.Set("act1", true)
	mod.Hierarchy.Set("obj", hierObj)
	mod.Actions.Set("obj.act1", NewSequence())

	// Two exports — only obj.act1 belongs to iso1's isolate
	exp1 := cfg.NewExportDef(cfg.NewAtom("obj.act1"), nil)
	exp2 := cfg.NewExportDef(cfg.NewAtom("other.act2"), nil)
	mod.Exports = append(mod.Exports, exp1, exp2)

	// iso1 verifies "obj"
	mod.Isolates["iso1"] = cfg.NewIsolateDef(
		[]Node{cfg.NewAtom("iso1"), cfg.NewAtom("obj")}, 0,
	)

	CreateConjActions(mod)

	acts, ok := mod.ConjActions["obj.sub.inv"]
	if !ok {
		t.Fatal("expected ConjActions entry for 'obj.sub.inv'")
	}
	// Should walk up: obj.sub.inv → obj.sub (not found) → obj (found in iso1)
	// → use iso1's exports only (obj.act1)
	if len(acts) != 1 {
		t.Errorf("obj.sub.inv should map to 1 action via hierarchy walk, got %d: %v", len(acts), acts)
	} else if acts[0] != "obj.act1" {
		t.Errorf("expected obj.act1, got %s", acts[0])
	}
}

// Test 16: Conjecture with nil label → skipped, no ConjActions entry.
func TestCreateConjActions_NoLabelSkipped(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	conjLF := cfg.NewLabeledFormula(nil, cfg.NewAtom("unlabeled"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	exp1 := cfg.NewExportDef(cfg.NewAtom("act1"), nil)
	mod.Exports = append(mod.Exports, exp1)

	CreateConjActions(mod)

	if len(mod.ConjActions) != 0 {
		t.Errorf("conjecture with nil label should be skipped, got %d entries", len(mod.ConjActions))
	}
}

const createConjActionsInterferenceSource = `#lang ivy1.7
individual p : bool
object x = {
    action victim = { call o.a }
}
object o = {
    action a = { call x.victim }
    invariant [inv] p
}
object y = {
    action env = { call x.victim }
}
export y.env
isolate i = x with o
isolate j = o
`

type pythonCompileOracle struct {
	OK  bool   `json:"ok"`
	Err string `json:"err"`
}

func pythonCompileSourceOracle(t *testing.T, src string) pythonCompileOracle {
	t.Helper()
	cmd := pythonIvyCommandForTest(t, "-O", "-c", `
import io
import json
from ivy import ivy_compiler as ic, ivy_module as im

src = `+pythonTripleQuote(src)+`
try:
    with im.Module():
        decls = ic.read_module(io.StringIO(src))
        ic.ivy_compile(decls, im.module, create_isolate=False)
        print(json.dumps({"ok": True, "err": ""}))
except Exception as e:
    print(json.dumps({"ok": False, "err": str(e)}))
`)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("python compile oracle failed: %v\n%s", err, out)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	var got pythonCompileOracle
	if err := json.Unmarshal([]byte(lines[len(lines)-1]), &got); err != nil {
		t.Fatalf("decode python compile oracle %q: %v\nfull output:\n%s", lines[len(lines)-1], err, out)
	}
	return got
}

// Test 17: Interference detection — cross-isolate invariant violations.
func TestCreateConjActions_InterferenceDetection(t *testing.T) {
	want := pythonCompileSourceOracle(t, createConjActionsInterferenceSource)
	if want.OK || !strings.Contains(want.Err, "depends on invariant o.inv") {
		t.Fatalf("python oracle sanity check failed: %+v", want)
	}

	result, err := Parse(createConjActionsInterferenceSource, Version{1, 7})
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	mod := New()
	mod.Cfg = NewConfig()
	err = IvyCompile(result.Decls, mod, false)
	if err == nil {
		t.Fatalf("IvyCompile accepted object-invariant interference; Python rejected it with: %s", want.Err)
	}
	if !strings.Contains(err.Error(), "depends on invariant o.inv") {
		t.Fatalf("Go compile error differs from Python\nwant substring: %q\ngot: %v", "depends on invariant o.inv", err)
	}
}

// ============================================================================
// §6.3 #29: ConjSetup pass 2
// ============================================================================

// Test 18: ConjectureDecl → appended to LabeledConjs with compiled formula and label.
func TestConjSetup_ConjectureAddedToLabeledConjs(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	// Build a conjecture decl containing a labeled formula
	conjBody := cfg.NewAtom("true") // simple body
	conjLF := cfg.NewLabeledFormula(cfg.NewAtom("inv1"), conjBody)
	conjDecl := cfg.NewConjectureDecl(conjLF)

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]Node{conjDecl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.LabeledConjs) != 1 {
		t.Fatalf("expected 1 labeled conjecture, got %d", len(c.Module.LabeledConjs))
	}

	lc := c.Module.LabeledConjs[0]
	if lc.Label == nil {
		t.Error("conjecture label should be preserved, got nil")
	} else {
		if a, ok := lc.Label.(*Atom); !ok || a.Relname() != "inv1" {
			t.Errorf("expected label 'inv1', got %v", lc.Label)
		}
	}
	if lc.Formula == nil {
		t.Error("conjecture formula should not be nil")
	}
}

// Test 19: Conjecture then unlabeled proof → proof attaches to conjecture.
func TestConjSetup_ProofAttachesToConjecture(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	conjBody := cfg.NewAtom("true")
	conjLF := cfg.NewLabeledFormula(cfg.NewAtom("inv1"), conjBody)
	conjDecl := cfg.NewConjectureDecl(conjLF)

	proofBody := cfg.NewAtom("proof_step") // unlabeled proof body
	proofDecl := cfg.NewProofDecl(proofBody)

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]Node{conjDecl, proofDecl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.Proofs) != 1 {
		t.Fatalf("expected 1 proof entry, got %d", len(c.Module.Proofs))
	}

	pe := c.Module.Proofs[0]
	if pe.Formula == nil {
		t.Fatal("proof entry formula should not be nil")
	}
	// The proof's formula should be the conjecture's LabeledFormula
	if pe.Formula != c.Module.LabeledConjs[0] {
		t.Error("proof should attach to the conjecture's labeled formula")
	}
}

// Test 20: Property then proof → proof NOT attached (property clears lastFact).
func TestConjSetup_ProofAfterPropertyNotAttached(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	propDecl := cfg.NewPropertyDecl(cfg.NewAtom("prop_body"))
	proofBody := cfg.NewAtom("proof_step")
	proofDecl := cfg.NewProofDecl(proofBody)

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]Node{propDecl, proofDecl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.Proofs) != 0 {
		t.Errorf("expected 0 proof entries (property clears lastFact), got %d", len(c.Module.Proofs))
	}
}

// Test 21: Conjecture then labeled proof → NOT attached.
func TestConjSetup_LabeledProofSkipped(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	conjBody := cfg.NewAtom("true")
	conjLF := cfg.NewLabeledFormula(cfg.NewAtom("inv1"), conjBody)
	conjDecl := cfg.NewConjectureDecl(conjLF)

	// Labeled proof: the DeclArg is itself a LabeledFormula
	labeledProofBody := cfg.NewLabeledFormula(cfg.NewAtom("pf_label"), cfg.NewAtom("proof_body"))
	proofDecl := cfg.NewProofDecl(labeledProofBody)

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]Node{conjDecl, proofDecl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.Proofs) != 0 {
		t.Errorf("labeled proof should be skipped, got %d proof entries", len(c.Module.Proofs))
	}
}

// Test 22: conj1, conj2, proof → proof attaches to conj2 (most recent).
func TestConjSetup_MultipleConjecturesLastFactTracking(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	conj1 := cfg.NewConjectureDecl(cfg.NewLabeledFormula(cfg.NewAtom("inv1"), cfg.NewAtom("true")))
	conj2 := cfg.NewConjectureDecl(cfg.NewLabeledFormula(cfg.NewAtom("inv2"), cfg.NewAtom("true")))
	proof := cfg.NewProofDecl(cfg.NewAtom("proof_step"))

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]Node{conj1, conj2, proof})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.LabeledConjs) != 2 {
		t.Fatalf("expected 2 conjectures, got %d", len(c.Module.LabeledConjs))
	}
	if len(c.Module.Proofs) != 1 {
		t.Fatalf("expected 1 proof, got %d", len(c.Module.Proofs))
	}

	// Proof should attach to conj2 (the second conjecture)
	pe := c.Module.Proofs[0]
	if pe.Formula != c.Module.LabeledConjs[1] {
		t.Error("proof should attach to the second (most recent) conjecture")
	}
}

// Test 23: conjecture, definition, proof → proof NOT attached (definition clears lastFact).
func TestConjSetup_DefinitionClearsLastFact(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	conjDecl := cfg.NewConjectureDecl(cfg.NewLabeledFormula(cfg.NewAtom("inv1"), cfg.NewAtom("body")))
	defDecl := cfg.NewDefinitionDecl(cfg.NewAtom("def_body"))
	proofDecl := cfg.NewProofDecl(cfg.NewAtom("proof_step"))

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]Node{conjDecl, defDecl, proofDecl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.Proofs) != 0 {
		t.Errorf("definition should clear lastFact; expected 0 proofs, got %d", len(c.Module.Proofs))
	}
}

// Test 24: conjecture, theorem, proof → proof NOT attached (theorem clears lastFact).
func TestConjSetup_TheoremClearsLastFact(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	conjDecl := cfg.NewConjectureDecl(cfg.NewLabeledFormula(cfg.NewAtom("inv1"), cfg.NewAtom("body")))
	thDecl := cfg.NewTheoremDecl(cfg.NewAtom("th_body"))
	proofDecl := cfg.NewProofDecl(cfg.NewAtom("proof_step"))

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]Node{conjDecl, thDecl, proofDecl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.Proofs) != 0 {
		t.Errorf("theorem should clear lastFact; expected 0 proofs, got %d", len(c.Module.Proofs))
	}
}

// Test 25: Conjecture label like "obj.inv1" is preserved.
func TestConjSetup_ConjectureLabel_Preserved(t *testing.T) {
	cfg := NewAstConfig()
	c := newTestCompiler()

	conjBody := cfg.NewAtom("true")
	conjLF := cfg.NewLabeledFormula(cfg.NewAtom("obj.inv1"), conjBody)
	conjDecl := cfg.NewConjectureDecl(conjLF)

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]Node{conjDecl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.LabeledConjs) != 1 {
		t.Fatalf("expected 1 conjecture, got %d", len(c.Module.LabeledConjs))
	}

	lc := c.Module.LabeledConjs[0]
	if lc.Label == nil {
		t.Fatal("label should not be nil")
	}
	if a, ok := lc.Label.(*Atom); !ok {
		t.Errorf("label should be *ast.Atom, got %T", lc.Label)
	} else if a.Relname() != "obj.inv1" {
		t.Errorf("expected label 'obj.inv1', got '%s'", a.Relname())
	}
}

// ============================================================================
// New tests added during TDD refactor
// ============================================================================

// Test 26: Version comparison edge case — "1.10" is semantically > "1.7"
// but lexicographically < "1.7". After the VersionLE fix, the action
// interference check (v1.7+) should still trigger at version 1.10.
func TestCheckDefinitions_VersionComparisonSemantic(t *testing.T) {
	mod := New()
	oldVer := mod.Cfg.IuCfg.GetStringVersion()
	defer SetStringVersionOn(mod.Cfg.IuCfg, oldVer)
	SetStringVersionOn(mod.Cfg.IuCfg, "1.10") // > 1.7 semantically but < "1.7" lexicographically

	cfg := mod.Cfg.AstCfg

	// Axiom uses symbol 'f' — use compiled lg.Const so structural keys match
	fSym := NewConst("f", Boolean)
	axiomLF := makeLabeledFormula(cfg, "ax1", fSym)
	mod.LabeledAxioms = append(mod.LabeledAxioms, axiomLF)

	// Action assigns to 'f'
	mod.Actions.Set("act1", NewAssignAction(fSym, NewConst("true_val", Boolean)))

	err := CheckDefinitions(mod)
	if err == nil {
		t.Fatal("expected error for action modifying axiom symbol 'f' at version 1.10, got nil")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "immutable") {
		t.Errorf("error should mention immutable symbol, got: %s", err.Error())
	}
}

// Test 27: Definition of an interpreted symbol should be rejected.
func TestCheckDefinitions_InterpretedSymbolError(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg
	// Mark "myint" as interpreted in the signature
	mod.Sig.Interp["myint"] = &UninterpretedSort{Name: "int"}

	// Definition of "myint" should be rejected
	defMyint := makeLabeledDef(cfg, "def_myint", makeLogicDef("myint"))
	mod.LabeledProps = []*LabeledFormula{defMyint}

	err := CheckDefinitions(mod)
	if err == nil {
		t.Fatal("expected error for definition of interpreted symbol 'myint'")
	}
	if !strings.Contains(err.Error(), "interpreted symbol") {
		t.Errorf("error should mention 'interpreted symbol', got: %s", err.Error())
	}
}

// Test 28: Definition ordering — multiple clean definitions maintain their order.
func TestCheckDefinitions_OrderPreserved(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	defA := makeLabeledDef(cfg, "defA", makeLogicDef("a"))
	defB := makeLabeledDef(cfg, "defB", makeLogicDef("b"))
	defC := makeLabeledDef(cfg, "defC", makeLogicDef("c"))

	mod.LabeledProps = []*LabeledFormula{defA, defB, defC}

	err := CheckDefinitions(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mod.Definitions) != 3 {
		t.Fatalf("expected 3 definitions, got %d", len(mod.Definitions))
	}
	if mod.Definitions[0] != defA || mod.Definitions[1] != defB || mod.Definitions[2] != defC {
		t.Error("definitions should be in insertion order: a, b, c")
	}
}

// Test 29: Stale-symbol transitivity — definition h = f(g(x)) where g is stale
// means h should stay in LabeledProps even though f is clean.
func TestCheckDefinitions_StaleSymbolTransitive(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	// defF: f = true_const (clean, no proof, no stale deps)
	defF := makeLabeledDef(cfg, "defF", makeLogicDef("f"))

	// defG: g = f (has proof → g becomes stale)
	defG := makeLabeledDef(cfg, "defG", makeLogicDef("g", "f"))
	mod.Proofs = append(mod.Proofs, ProofEntry{Formula: defG, Proof: cfg.NewAtom("pf")})

	// defH: h = g (uses stale g → h should NOT move to Definitions, and h becomes stale)
	defH := makeLabeledDef(cfg, "defH", makeLogicDef("h", "g"))

	// defK: k = h (uses stale h → k should also NOT move to Definitions)
	defK := makeLabeledDef(cfg, "defK", makeLogicDef("k", "h"))

	mod.LabeledProps = []*LabeledFormula{defF, defG, defH, defK}

	err := CheckDefinitions(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// defF should be in Definitions (clean)
	foundF := false
	for _, d := range mod.Definitions {
		if d == defF {
			foundF = true
		}
	}
	if !foundF {
		t.Error("defF should be in Definitions")
	}

	// defH and defK should stay in LabeledProps (stale dependency chain)
	foundH, foundK := false, false
	for _, p := range mod.LabeledProps {
		if p == defH {
			foundH = true
		}
		if p == defK {
			foundK = true
		}
	}
	if !foundH {
		t.Error("defH should remain in LabeledProps (uses stale g)")
	}
	if !foundK {
		t.Error("defK should remain in LabeledProps (uses stale h transitively)")
	}
}

// Test 30: VersionLE in CreateConjActions — version "1.10" should NOT skip (> 1.6).
func TestCreateConjActions_VersionSemantic(t *testing.T) {
	mod := New()
	oldVer := mod.Cfg.IuCfg.GetStringVersion()
	defer SetStringVersionOn(mod.Cfg.IuCfg, oldVer)
	SetStringVersionOn(mod.Cfg.IuCfg, "1.10") // > 1.6 semantically

	cfg := mod.Cfg.AstCfg
	conjLF := makeLabeledFormula(cfg, "this.inv1", cfg.NewAtom("conj_body"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	exp1 := cfg.NewExportDef(cfg.NewAtom("act1"), nil)
	mod.Exports = append(mod.Exports, exp1)

	CreateConjActions(mod)

	if len(mod.ConjActions) == 0 {
		t.Error("version 1.10 should NOT be treated as <= 1.6; ConjActions should have entries")
	}
}

// Test 31: definesName with Apply LHS — f(x) = body should extract "f", not "f(x)".
func TestCheckDefinitions_ApplyLHSExtractsName(t *testing.T) {
	mod := New()
	cfg := mod.Cfg.AstCfg

	// Build definition with Apply LHS: f(x) = true_const
	fSym := NewConst("f", &LogicFunctionSort{Sorts: []Sort{Boolean, Boolean}})
	xSym := NewConst("x", Boolean)
	lhs := MustApply(fSym, xSym)
	rhs := NewConst("true_const", Boolean)
	def := &LogicDefinition{Lhs: lhs, Rhs: rhs}

	defLF := cfg.NewLabeledFormula(cfg.NewAtom("def_f"), def)
	mod.LabeledProps = []*LabeledFormula{defLF}

	err := CheckDefinitions(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have moved to Definitions
	if len(mod.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(mod.Definitions))
	}

	// Verify name extraction works correctly via redefinition
	mod2 := New()
	def1 := cfg.NewLabeledFormula(cfg.NewAtom("def1"), &LogicDefinition{Lhs: lhs, Rhs: rhs})
	def2 := cfg.NewLabeledFormula(cfg.NewAtom("def2"), &LogicDefinition{Lhs: fSym, Rhs: rhs})
	mod2.LabeledProps = []*LabeledFormula{def1, def2}

	err = CheckDefinitions(mod2)
	if err == nil {
		t.Fatal("expected redefinition error for 'f' (Apply LHS + Symbol LHS)")
	}
	if !strings.Contains(err.Error(), "redefinition") {
		t.Errorf("error should mention redefinition, got: %s", err.Error())
	}
}

// Test 32: opt_mutax guard — when OptMutax is true, axiom interference is allowed.
func TestCheckDefinitions_OptMutaxAllowsAxiomInterference(t *testing.T) {
	mod := New()
	oldVer := mod.Cfg.IuCfg.GetStringVersion()
	defer SetStringVersionOn(mod.Cfg.IuCfg, oldVer)
	SetStringVersionOn(mod.Cfg.IuCfg, "1.7")

	// Save and restore OptMutax
	oldMutax := OptMutax.GetBool()
	defer func() {
		if oldMutax {
			OptMutax.Set("true")
		} else {
			OptMutax.Set("false")
		}
	}()
	OptMutax.Set("true")

	cfg := mod.Cfg.AstCfg

	// Axiom uses symbol 'f'
	axiomLF := makeLabeledFormula(cfg, "ax1", cfg.NewAtom("f"))
	mod.LabeledAxioms = append(mod.LabeledAxioms, axiomLF)

	// Action that assigns to 'f'
	fSym := NewConst("f", Boolean)
	mod.Actions.Set("act1", NewAssignAction(fSym, NewConst("true_val", Boolean)))

	err := CheckDefinitions(mod)
	if err != nil {
		t.Fatalf("with opt_mutax=true, axiom interference should be allowed, got: %v", err)
	}
}

// Test 33: opt_mutax guard — definition LHS check is NOT skipped even with opt_mutax=true.
func TestCheckDefinitions_OptMutaxStillChecksDefinitionLHS(t *testing.T) {
	mod := New()
	oldVer := mod.Cfg.IuCfg.GetStringVersion()
	defer SetStringVersionOn(mod.Cfg.IuCfg, oldVer)
	SetStringVersionOn(mod.Cfg.IuCfg, "1.7")

	// Save and restore OptMutax
	oldMutax := OptMutax.GetBool()
	defer func() {
		if oldMutax {
			OptMutax.Set("true")
		} else {
			OptMutax.Set("false")
		}
	}()
	OptMutax.Set("true")

	cfg := mod.Cfg.AstCfg

	// Definition of 'f'
	defF := makeLabeledDef(cfg, "defF", makeLogicDef("f"))
	mod.LabeledProps = []*LabeledFormula{defF}

	// Action assigns to 'f'
	fSym := NewConst("f", Boolean)
	mod.Actions.Set("act1", NewAssignAction(fSym, NewConst("true_val", Boolean)))

	err := CheckDefinitions(mod)
	if err == nil {
		t.Fatal("even with opt_mutax=true, definition LHS should still be immutable")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "assigned") {
		t.Errorf("error should mention symbol being assigned, got: %s", err.Error())
	}
}
