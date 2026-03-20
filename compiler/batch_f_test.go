package compiler

// Batch F — Tests for:
//   §6.2 #5:  CheckDefinitions — separate defs from props, check redefinition, detect cycles
//   §6.2 #7:  CreateConjActions — determine which actions must preserve each conjecture
//   §6.3 #29: ConjSetup pass 2 — implement real conjecture setup

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/ast"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// ============================================================================
// §6.2 #5: CheckDefinitions
// ============================================================================

// helper: make a lg.Definition wrapping symbol names, usable as LabeledFormula.Formula
func makeLogicDef(defSym string, rhsSyms ...string) *lg.Definition {
	lhs := lg.NewSymbol(defSym, lg.Boolean)
	var rhs lg.Expr
	if len(rhsSyms) == 0 {
		rhs = lg.NewSymbol("true_const", lg.Boolean)
	} else if len(rhsSyms) == 1 {
		rhs = lg.NewSymbol(rhsSyms[0], lg.Boolean)
	} else {
		// Build an Apply chain so UsedConstantsList can find all rhs symbols.
		// Use nested Eq nodes: (rhsSyms[0] = rhsSyms[1]) for simplicity.
		// For > 2, just use the first two — tests only need specific symbols visible.
		s0 := lg.NewSymbol(rhsSyms[0], lg.Boolean)
		s1 := lg.NewSymbol(rhsSyms[1], lg.Boolean)
		rhs = lg.NewDefinition(s0, s1) // nested def as expression — carries both constants
	}
	return lg.NewDefinition(lhs, rhs)
}

// helper: wrap a lg.Definition in a LabeledFormula with a label
func makeLabeledDef(label string, def *lg.Definition) *ast.LabeledFormula {
	lf := ast.NewLabeledFormula(ast.NewAtom(label), def)
	return lf
}

// helper: wrap an arbitrary formula in a LabeledFormula
func makeLabeledFormula(label string, formula ast.Node) *ast.LabeledFormula {
	return ast.NewLabeledFormula(ast.NewAtom(label), formula)
}

// Test 1: Definitions without proofs move to mod.Definitions;
// non-definitions stay in LabeledProps; definitions with proofs stay in LabeledProps.
func TestCheckDefinitions_SeparatesDefsFromProps(t *testing.T) {
	mod := module.New()

	// defA: definition of f (no proof) — should move to Definitions
	defA := makeLabeledDef("defA", makeLogicDef("f"))

	// propB: non-definition property — should stay in LabeledProps
	propB := makeLabeledFormula("propB", ast.NewAtom("some_prop"))

	// defC: definition of g WITH a proof — should stay in LabeledProps (stale path)
	defC := makeLabeledDef("defC", makeLogicDef("g"))
	mod.Proofs = append(mod.Proofs, module.ProofEntry{Formula: defC, Proof: ast.NewAtom("proof_body")})

	mod.LabeledProps = []*ast.LabeledFormula{defA, propB, defC}

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
	mod := module.New()

	// defA: f = true_const (no proof, no stale deps)
	defA := makeLabeledDef("defA", makeLogicDef("f"))

	// defB: g = f (has proof → g becomes stale)
	defB := makeLabeledDef("defB", makeLogicDef("g", "f"))
	mod.Proofs = append(mod.Proofs, module.ProofEntry{Formula: defB, Proof: ast.NewAtom("pf")})

	// defC: h = g (g is stale → C should NOT move to Definitions)
	defC := makeLabeledDef("defC", makeLogicDef("h", "g"))

	mod.LabeledProps = []*ast.LabeledFormula{defA, defB, defC}

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
	mod := module.New()

	def1 := makeLabeledDef("def1", makeLogicDef("f"))
	def2 := makeLabeledDef("def2", makeLogicDef("f"))

	mod.LabeledProps = []*ast.LabeledFormula{def1, def2}

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
	mod := module.New()

	// Put one definition of 'f' as a LabeledProp (will move to Definitions)
	def1 := makeLabeledDef("def1", makeLogicDef("f"))
	mod.LabeledProps = []*ast.LabeledFormula{def1}

	// Put another definition of 'f' in NativeDefinitions
	nativeDef := makeLabeledDef("native_f", makeLogicDef("f"))
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
	mod := module.New()

	def1 := makeLabeledDef("def1", makeLogicDef("f"))
	mod.LabeledProps = []*ast.LabeledFormula{def1}

	// Named entry with symbol 'f'
	fSym := lg.NewSymbol("f", lg.Boolean)
	namedLF := makeLabeledFormula("named_f", ast.NewAtom("something"))
	mod.Named = append(mod.Named, module.NamedEntry{Formula: namedLF, Name: fSym})

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
	mod := module.New()

	// f = g
	defF := makeLabeledDef("defF", makeLogicDef("f", "g"))
	// g = f
	defG := makeLabeledDef("defG", makeLogicDef("g", "f"))

	mod.LabeledProps = []*ast.LabeledFormula{defF, defG}

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
	mod := module.New()

	// f = f (self-loop)
	defF := makeLabeledDef("defF", makeLogicDef("f", "f"))
	mod.LabeledProps = []*ast.LabeledFormula{defF}

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
	mod := module.New()

	// f = f (self-loop) but has a proof
	defF := makeLabeledDef("defF", makeLogicDef("f", "f"))
	mod.Proofs = append(mod.Proofs, module.ProofEntry{Formula: defF, Proof: ast.NewAtom("rec_proof")})
	mod.LabeledProps = []*ast.LabeledFormula{defF}

	err := CheckDefinitions(mod)
	if err != nil {
		t.Fatalf("self-recursive def with proof should be accepted, got error: %v", err)
	}
	// The def should remain in LabeledProps (has proof → stale → stays)
	// OR be accepted via admit_definition. Either way, no error.
}

// Test 9: Action modifies symbol used in axiom → error for version >= 1.7.
func TestCheckDefinitions_ActionInterference_ModifiesAxiomSymbol(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()

	// Axiom uses symbol 'f' — use compiled lg.Symbol so structural keys match
	fSym := lg.NewSymbol("f", lg.Boolean)
	axiomLF := makeLabeledFormula("ax1", fSym)
	mod.LabeledAxioms = append(mod.LabeledAxioms, axiomLF)

	// Action that assigns to 'f' — uses real AssignAction so actions.Modifies finds it
	assignAction := actions.NewAssignAction(fSym, lg.NewSymbol("true_val", lg.Boolean))
	mod.Actions["act1"] = assignAction
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
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()

	// Definition of 'f'
	defF := makeLabeledDef("defF", makeLogicDef("f"))
	mod.LabeledProps = []*ast.LabeledFormula{defF}

	// Action that assigns to 'f'
	fAssign := lg.NewSymbol("f", lg.Boolean)
	mod.Actions["act1"] = actions.NewAssignAction(fAssign, lg.NewSymbol("true_val", lg.Boolean))

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
	mod := module.New()

	defA := makeLabeledDef("defA", makeLogicDef("a"))
	defB := makeLabeledDef("defB", makeLogicDef("b"))
	defC := makeLabeledDef("defC", makeLogicDef("c"))

	mod.LabeledProps = []*ast.LabeledFormula{defA, defB, defC}

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
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.6")

	mod := module.New()

	// Add a conjecture and an export
	conjLF := makeLabeledFormula("this.inv1", ast.NewAtom("conj_body"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	expDef := &ast.ExportDef{ExportedNode: ast.NewAtom("act1")}
	mod.Exports = append(mod.Exports, expDef)

	CreateConjActions(mod)

	if len(mod.ConjActions) != 0 {
		t.Errorf("for version 1.6, ConjActions should be empty, got %d entries", len(mod.ConjActions))
	}
}

// Test 13: Top-level conjecture with no isolates → all exports.
func TestCreateConjActions_TopLevelConj_AllExports(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()

	conjLF := makeLabeledFormula("this.inv1", ast.NewAtom("conj_body"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	exp1 := &ast.ExportDef{ExportedNode: ast.NewAtom("act1")}
	exp2 := &ast.ExportDef{ExportedNode: ast.NewAtom("act2")}
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
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()

	// Two conjectures in different objects
	conjA := makeLabeledFormula("obj_a.inv", ast.NewAtom("inv_a"))
	conjB := makeLabeledFormula("obj_b.inv", ast.NewAtom("inv_b"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjA, conjB)

	// Hierarchy: obj_a has child act1, obj_b has child act2.
	// IterIsolate walks verified names through mod.Hierarchy to find actions.
	mod.Hierarchy["obj_a"] = map[string]bool{"act1": true}
	mod.Hierarchy["obj_b"] = map[string]bool{"act2": true}

	// Actions registered with composed names (obj_a.act1, obj_b.act2)
	mod.Actions["obj_a.act1"] = ast.NewAtom("skip") // placeholder action body
	mod.Actions["obj_b.act2"] = ast.NewAtom("skip")

	// Exports use the composed action names
	exp1 := &ast.ExportDef{ExportedNode: ast.NewAtom("obj_a.act1")}
	exp2 := &ast.ExportDef{ExportedNode: ast.NewAtom("obj_b.act2")}
	mod.Exports = append(mod.Exports, exp1, exp2)

	// Isolate iso_a verifies obj_a, iso_b verifies obj_b
	mod.Isolates["iso_a"] = &ast.IsolateDef{
		Elems: []ast.Node{ast.NewAtom("iso_a"), ast.NewAtom("obj_a")}, WithArgs: 0,
	}
	mod.Isolates["iso_b"] = &ast.IsolateDef{
		Elems: []ast.Node{ast.NewAtom("iso_b"), ast.NewAtom("obj_b")}, WithArgs: 0,
	}

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
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()

	conjLF := makeLabeledFormula("obj.sub.inv", ast.NewAtom("nested_inv"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	// Hierarchy: obj has child act1
	mod.Hierarchy["obj"] = map[string]bool{"act1": true}
	mod.Actions["obj.act1"] = ast.NewAtom("skip")

	// Two exports — only obj.act1 belongs to iso1's isolate
	exp1 := &ast.ExportDef{ExportedNode: ast.NewAtom("obj.act1")}
	exp2 := &ast.ExportDef{ExportedNode: ast.NewAtom("other.act2")}
	mod.Exports = append(mod.Exports, exp1, exp2)

	// iso1 verifies "obj"
	mod.Isolates["iso1"] = &ast.IsolateDef{
		Elems: []ast.Node{ast.NewAtom("iso1"), ast.NewAtom("obj")}, WithArgs: 0,
	}

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
	mod := module.New()

	conjLF := &ast.LabeledFormula{Formula: ast.NewAtom("unlabeled")}
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	exp1 := &ast.ExportDef{ExportedNode: ast.NewAtom("act1")}
	mod.Exports = append(mod.Exports, exp1)

	CreateConjActions(mod)

	if len(mod.ConjActions) != 0 {
		t.Errorf("conjecture with nil label should be skipped, got %d entries", len(mod.ConjActions))
	}
}

// Test 17: Interference detection — cross-isolate invariant violations.
// TODO: Python's create_conj_actions has a do_check_interference section that
// detects when an isolate's action could invalidate another isolate's conjecture.
// This is not yet ported to Go. When ported, this test should verify that
// cross-isolate interference is detected and raises an error.
func TestCreateConjActions_InterferenceDetection(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()

	conjLF := makeLabeledFormula("obj1.inv", ast.NewAtom("inv1"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	exp1 := &ast.ExportDef{ExportedNode: ast.NewAtom("act1")}
	mod.Exports = append(mod.Exports, exp1)
	mod.Isolates["iso1"] = ast.NewAtom("iso1_def")
	mod.Isolates["iso2"] = ast.NewAtom("iso2_def")

	CreateConjActions(mod)

	acts := mod.ConjActions["obj1.inv"]
	t.Logf("TODO: interference detection not yet ported from Python; got %d actions for obj1.inv", len(acts))
}

// ============================================================================
// §6.3 #29: ConjSetup pass 2
// ============================================================================

// Test 18: ConjectureDecl → appended to LabeledConjs with compiled formula and label.
func TestConjSetup_ConjectureAddedToLabeledConjs(t *testing.T) {
	c := newTestCompiler()

	// Build a conjecture decl containing a labeled formula
	conjBody := ast.NewAtom("true") // simple body
	conjLF := ast.NewLabeledFormula(ast.NewAtom("inv1"), conjBody)
	conjDecl := ast.NewConjectureDecl(conjLF)

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]ast.Node{conjDecl})
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
		if a, ok := lc.Label.(*ast.Atom); !ok || a.Relname() != "inv1" {
			t.Errorf("expected label 'inv1', got %v", lc.Label)
		}
	}
	if lc.Formula == nil {
		t.Error("conjecture formula should not be nil")
	}
}

// Test 19: Conjecture then unlabeled proof → proof attaches to conjecture.
func TestConjSetup_ProofAttachesToConjecture(t *testing.T) {
	c := newTestCompiler()

	conjBody := ast.NewAtom("true")
	conjLF := ast.NewLabeledFormula(ast.NewAtom("inv1"), conjBody)
	conjDecl := ast.NewConjectureDecl(conjLF)

	proofBody := ast.NewAtom("proof_step") // unlabeled proof body
	proofDecl := ast.NewProofDecl(proofBody)

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]ast.Node{conjDecl, proofDecl})
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
	c := newTestCompiler()

	propDecl := ast.NewPropertyDecl(ast.NewAtom("prop_body"))
	proofBody := ast.NewAtom("proof_step")
	proofDecl := ast.NewProofDecl(proofBody)

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]ast.Node{propDecl, proofDecl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.Proofs) != 0 {
		t.Errorf("expected 0 proof entries (property clears lastFact), got %d", len(c.Module.Proofs))
	}
}

// Test 21: Conjecture then labeled proof → NOT attached.
func TestConjSetup_LabeledProofSkipped(t *testing.T) {
	c := newTestCompiler()

	conjBody := ast.NewAtom("true")
	conjLF := ast.NewLabeledFormula(ast.NewAtom("inv1"), conjBody)
	conjDecl := ast.NewConjectureDecl(conjLF)

	// Labeled proof: the DeclArg is itself a LabeledFormula
	labeledProofBody := ast.NewLabeledFormula(ast.NewAtom("pf_label"), ast.NewAtom("proof_body"))
	proofDecl := ast.NewProofDecl(labeledProofBody)

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]ast.Node{conjDecl, proofDecl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.Proofs) != 0 {
		t.Errorf("labeled proof should be skipped, got %d proof entries", len(c.Module.Proofs))
	}
}

// Test 22: conj1, conj2, proof → proof attaches to conj2 (most recent).
func TestConjSetup_MultipleConjecturesLastFactTracking(t *testing.T) {
	c := newTestCompiler()

	conj1 := ast.NewConjectureDecl(ast.NewLabeledFormula(ast.NewAtom("inv1"), ast.NewAtom("true")))
	conj2 := ast.NewConjectureDecl(ast.NewLabeledFormula(ast.NewAtom("inv2"), ast.NewAtom("true")))
	proof := ast.NewProofDecl(ast.NewAtom("proof_step"))

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]ast.Node{conj1, conj2, proof})
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
	c := newTestCompiler()

	conjDecl := ast.NewConjectureDecl(ast.NewLabeledFormula(ast.NewAtom("inv1"), ast.NewAtom("body")))
	defDecl := ast.NewDefinitionDecl(ast.NewAtom("def_body"))
	proofDecl := ast.NewProofDecl(ast.NewAtom("proof_step"))

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]ast.Node{conjDecl, defDecl, proofDecl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.Proofs) != 0 {
		t.Errorf("definition should clear lastFact; expected 0 proofs, got %d", len(c.Module.Proofs))
	}
}

// Test 24: conjecture, theorem, proof → proof NOT attached (theorem clears lastFact).
func TestConjSetup_TheoremClearsLastFact(t *testing.T) {
	c := newTestCompiler()

	conjDecl := ast.NewConjectureDecl(ast.NewLabeledFormula(ast.NewAtom("inv1"), ast.NewAtom("body")))
	thDecl := ast.NewTheoremDecl(ast.NewAtom("th_body"))
	proofDecl := ast.NewProofDecl(ast.NewAtom("proof_step"))

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]ast.Node{conjDecl, thDecl, proofDecl})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(c.Module.Proofs) != 0 {
		t.Errorf("theorem should clear lastFact; expected 0 proofs, got %d", len(c.Module.Proofs))
	}
}

// Test 25: Conjecture label like "obj.inv1" is preserved.
func TestConjSetup_ConjectureLabel_Preserved(t *testing.T) {
	c := newTestCompiler()

	conjBody := ast.NewAtom("true")
	conjLF := ast.NewLabeledFormula(ast.NewAtom("obj.inv1"), conjBody)
	conjDecl := ast.NewConjectureDecl(conjLF)

	cs := NewConjSetup(c)
	err := cs.ProcessDecls([]ast.Node{conjDecl})
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
	if a, ok := lc.Label.(*ast.Atom); !ok {
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
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.10") // > 1.7 semantically but < "1.7" lexicographically

	mod := module.New()

	// Axiom uses symbol 'f' — use compiled lg.Symbol so structural keys match
	fSym := lg.NewSymbol("f", lg.Boolean)
	axiomLF := makeLabeledFormula("ax1", fSym)
	mod.LabeledAxioms = append(mod.LabeledAxioms, axiomLF)

	// Action assigns to 'f'
	mod.Actions["act1"] = actions.NewAssignAction(fSym, lg.NewSymbol("true_val", lg.Boolean))

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
	mod := module.New()
	// Mark "myint" as interpreted in the signature
	mod.Sig.Interp["myint"] = &lg.UninterpretedSort{Name: "int"}

	// Definition of "myint" should be rejected
	defMyint := makeLabeledDef("def_myint", makeLogicDef("myint"))
	mod.LabeledProps = []*ast.LabeledFormula{defMyint}

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
	mod := module.New()

	defA := makeLabeledDef("defA", makeLogicDef("a"))
	defB := makeLabeledDef("defB", makeLogicDef("b"))
	defC := makeLabeledDef("defC", makeLogicDef("c"))

	mod.LabeledProps = []*ast.LabeledFormula{defA, defB, defC}

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
	mod := module.New()

	// defF: f = true_const (clean, no proof, no stale deps)
	defF := makeLabeledDef("defF", makeLogicDef("f"))

	// defG: g = f (has proof → g becomes stale)
	defG := makeLabeledDef("defG", makeLogicDef("g", "f"))
	mod.Proofs = append(mod.Proofs, module.ProofEntry{Formula: defG, Proof: ast.NewAtom("pf")})

	// defH: h = g (uses stale g → h should NOT move to Definitions, and h becomes stale)
	defH := makeLabeledDef("defH", makeLogicDef("h", "g"))

	// defK: k = h (uses stale h → k should also NOT move to Definitions)
	defK := makeLabeledDef("defK", makeLogicDef("k", "h"))

	mod.LabeledProps = []*ast.LabeledFormula{defF, defG, defH, defK}

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
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.10") // > 1.6 semantically

	mod := module.New()
	conjLF := makeLabeledFormula("this.inv1", ast.NewAtom("conj_body"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	exp1 := &ast.ExportDef{ExportedNode: ast.NewAtom("act1")}
	mod.Exports = append(mod.Exports, exp1)

	CreateConjActions(mod)

	if len(mod.ConjActions) == 0 {
		t.Error("version 1.10 should NOT be treated as <= 1.6; ConjActions should have entries")
	}
}

// Test 31: definesName with Apply LHS — f(x) = body should extract "f", not "f(x)".
func TestCheckDefinitions_ApplyLHSExtractsName(t *testing.T) {
	mod := module.New()

	// Build definition with Apply LHS: f(x) = true_const
	fSym := lg.NewSymbol("f", &lg.FunctionSort{Sorts: []lg.Sort{lg.Boolean, lg.Boolean}})
	xSym := lg.NewSymbol("x", lg.Boolean)
	lhs := &lg.Apply{Func: fSym, Terms: []lg.Expr{xSym}}
	rhs := lg.NewSymbol("true_const", lg.Boolean)
	def := &lg.Definition{Lhs: lhs, Rhs: rhs}

	defLF := ast.NewLabeledFormula(ast.NewAtom("def_f"), def)
	mod.LabeledProps = []*ast.LabeledFormula{defLF}

	err := CheckDefinitions(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have moved to Definitions
	if len(mod.Definitions) != 1 {
		t.Fatalf("expected 1 definition, got %d", len(mod.Definitions))
	}

	// Verify name extraction works correctly via redefinition
	mod2 := module.New()
	def1 := ast.NewLabeledFormula(ast.NewAtom("def1"), &lg.Definition{Lhs: lhs, Rhs: rhs})
	def2 := ast.NewLabeledFormula(ast.NewAtom("def2"), &lg.Definition{Lhs: fSym, Rhs: rhs})
	mod2.LabeledProps = []*ast.LabeledFormula{def1, def2}

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
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

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

	mod := module.New()

	// Axiom uses symbol 'f'
	axiomLF := makeLabeledFormula("ax1", ast.NewAtom("f"))
	mod.LabeledAxioms = append(mod.LabeledAxioms, axiomLF)

	// Action that assigns to 'f'
	fSym := lg.NewSymbol("f", lg.Boolean)
	mod.Actions["act1"] = actions.NewAssignAction(fSym, lg.NewSymbol("true_val", lg.Boolean))

	err := CheckDefinitions(mod)
	if err != nil {
		t.Fatalf("with opt_mutax=true, axiom interference should be allowed, got: %v", err)
	}
}

// Test 33: opt_mutax guard — definition LHS check is NOT skipped even with opt_mutax=true.
func TestCheckDefinitions_OptMutaxStillChecksDefinitionLHS(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

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

	mod := module.New()

	// Definition of 'f'
	defF := makeLabeledDef("defF", makeLogicDef("f"))
	mod.LabeledProps = []*ast.LabeledFormula{defF}

	// Action assigns to 'f'
	fSym := lg.NewSymbol("f", lg.Boolean)
	mod.Actions["act1"] = actions.NewAssignAction(fSym, lg.NewSymbol("true_val", lg.Boolean))

	err := CheckDefinitions(mod)
	if err == nil {
		t.Fatal("even with opt_mutax=true, definition LHS should still be immutable")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "assigned") {
		t.Errorf("error should mention symbol being assigned, got: %s", err.Error())
	}
}

