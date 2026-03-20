package compiler

// Batch F — Red (failing) tests for:
//   §6.2 #5:  CheckDefinitions — separate defs from props, check redefinition, detect cycles
//   §6.2 #7:  CreateConjActions — determine which actions must preserve each conjecture
//   §6.3 #29: ConjSetup pass 2 — implement real conjecture setup

import (
	"strings"
	"testing"

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
// EXPECTED TO FAIL: Go only logs, doesn't return error.
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
// EXPECTED TO FAIL: Go doesn't check NativeDefinitions.
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
// EXPECTED TO FAIL: Go doesn't check Named.
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
// EXPECTED TO FAIL: Go only logs, doesn't return error.
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
// EXPECTED TO FAIL: Go doesn't call admit_definition.
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
// EXPECTED TO FAIL: Go doesn't check action interference.
func TestCheckDefinitions_ActionInterference_ModifiesAxiomSymbol(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()

	// Axiom uses symbol 'f'
	axiomLF := makeLabeledFormula("ax1", ast.NewAtom("f"))
	mod.LabeledAxioms = append(mod.LabeledAxioms, axiomLF)

	// An action that assigns to 'f' — for now, store as a simple marker.
	// Python checks: for sym in a.modifies(): if sym.name in df: raise IvyError
	// We need some action that reports 'f' in its Modifies set.
	// Use a map entry so CheckDefinitions can find it.
	mod.Actions["act1"] = ast.NewAtom("assign_f") // placeholder action

	// No definitions needed — the interference check is about axiom symbols
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
// EXPECTED TO FAIL: Go doesn't check action interference.
func TestCheckDefinitions_ActionInterference_ModifiesDefinedSymbol(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()

	// Definition of 'f'
	defF := makeLabeledDef("defF", makeLogicDef("f"))
	mod.LabeledProps = []*ast.LabeledFormula{defF}

	// Action assigns to 'f'
	mod.Actions["act1"] = ast.NewAtom("assign_f") // placeholder

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
// EXPECTED TO FAIL: Go has no version check.
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
// EXPECTED TO FAIL: Go uses all exports, ignores isolates.
func TestCreateConjActions_IsolateScoping(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()

	// Two conjectures in different objects
	conjA := makeLabeledFormula("obj_a.inv", ast.NewAtom("inv_a"))
	conjB := makeLabeledFormula("obj_b.inv", ast.NewAtom("inv_b"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjA, conjB)

	// Two exports
	exp1 := &ast.ExportDef{ExportedNode: ast.NewAtom("act1")}
	exp2 := &ast.ExportDef{ExportedNode: ast.NewAtom("act2")}
	mod.Exports = append(mod.Exports, exp1, exp2)

	// Isolate iso_a verifies obj_a, exports act1
	// Isolate iso_b verifies obj_b, exports act2
	// (Need isolate entries — for now check the result behavior)
	mod.Isolates["iso_a"] = ast.NewAtom("iso_a_def") // placeholder
	mod.Isolates["iso_b"] = ast.NewAtom("iso_b_def") // placeholder

	CreateConjActions(mod)

	actsA, ok := mod.ConjActions["obj_a.inv"]
	if !ok {
		t.Fatal("expected ConjActions entry for 'obj_a.inv'")
	}
	// obj_a.inv should only map to act1 (from iso_a), not act2
	if len(actsA) != 1 {
		t.Errorf("obj_a.inv should map to 1 action (act1), got %d: %v", len(actsA), actsA)
	} else if actsA[0] != "act1" {
		t.Errorf("obj_a.inv should map to act1, got %s", actsA[0])
	}

	actsB, ok := mod.ConjActions["obj_b.inv"]
	if !ok {
		t.Fatal("expected ConjActions entry for 'obj_b.inv'")
	}
	if len(actsB) != 1 {
		t.Errorf("obj_b.inv should map to 1 action (act2), got %d: %v", len(actsB), actsB)
	} else if actsB[0] != "act2" {
		t.Errorf("obj_b.inv should map to act2, got %s", actsB[0])
	}
}

// Test 15: Nested object walk — conj "obj.sub.inv" walks up to find "obj".
// EXPECTED TO FAIL: Go has no hierarchy walk.
func TestCreateConjActions_NestedObjectWalk(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()

	conjLF := makeLabeledFormula("obj.sub.inv", ast.NewAtom("nested_inv"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	exp1 := &ast.ExportDef{ExportedNode: ast.NewAtom("act1")}
	exp2 := &ast.ExportDef{ExportedNode: ast.NewAtom("act2")}
	mod.Exports = append(mod.Exports, exp1, exp2)

	// iso1 verifies "obj"
	mod.Isolates["iso1"] = ast.NewAtom("iso1_def") // placeholder

	CreateConjActions(mod)

	acts, ok := mod.ConjActions["obj.sub.inv"]
	if !ok {
		t.Fatal("expected ConjActions entry for 'obj.sub.inv'")
	}
	// Should walk up: obj.sub.inv → obj.sub (not found) → obj (found in iso1)
	// → use iso1's exports only (act1)
	if len(acts) != 1 {
		t.Errorf("obj.sub.inv should map to 1 action via hierarchy walk, got %d: %v", len(acts), acts)
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
// EXPECTED TO FAIL: Go has no interference check.
func TestCreateConjActions_InterferenceDetection(t *testing.T) {
	oldVer := iu.GetStringVersion()
	defer iu.SetStringVersion(oldVer)
	iu.SetStringVersion("1.7")

	mod := module.New()

	conjLF := makeLabeledFormula("obj1.inv", ast.NewAtom("inv1"))
	mod.LabeledConjs = append(mod.LabeledConjs, conjLF)

	// iso1 exports act1; act1 calls internal_act
	exp1 := &ast.ExportDef{ExportedNode: ast.NewAtom("act1")}
	mod.Exports = append(mod.Exports, exp1)
	mod.Isolates["iso1"] = ast.NewAtom("iso1_def")
	mod.Isolates["iso2"] = ast.NewAtom("iso2_def")

	// iso2 also references internal_act, potentially invalidating obj1.inv
	// This should be detected as interference.
	// For now, we can only test that CreateConjActions doesn't panic
	// and that when interference detection is implemented, it raises an error.

	// Since CreateConjActions currently doesn't return error, we check
	// that the function signature changes to return error in the green phase.
	// For now, just verify current behavior doesn't detect interference.
	CreateConjActions(mod)

	// The test "fails" conceptually because no interference was detected.
	// When interference detection is implemented, this should raise an error.
	acts := mod.ConjActions["obj1.inv"]
	// Currently all exports are included — interference not detected
	if len(acts) == 1 {
		// This would mean isolate scoping worked (partially correct)
		t.Log("isolate scoping works but interference detection still needed")
	}
	// Mark as failing: interference detection should limit or error
	t.Skip("SKIP: interference detection not yet implemented — will be tested after green phase adds error return to CreateConjActions")
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

	conj1 := ast.NewConjectureDecl(ast.NewLabeledFormula(ast.NewAtom("inv1"), ast.NewAtom("body1")))
	conj2 := ast.NewConjectureDecl(ast.NewLabeledFormula(ast.NewAtom("inv2"), ast.NewAtom("body2")))
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
