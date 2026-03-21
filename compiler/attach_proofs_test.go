package compiler

import (
	"strings"
	"testing"

	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
)

// helper: make a label-only proof entry (formula field is nil, label is set).
// This is the Go equivalent of Python's (LabeledFormula(label, None), proof).
func makeLabelOnlyProof(label string, proof ast.Node) module.ProofEntry {
	lf := ast.NewLabeledFormula(ast.NewAtom(label), nil)
	return module.ProofEntry{Formula: lf, Proof: proof}
}

// helper: make a direct proof entry (formula is non-nil).
func makeDirectProof(label string, formula ast.Node, proof ast.Node) module.ProofEntry {
	lf := ast.NewLabeledFormula(ast.NewAtom(label), formula)
	return module.ProofEntry{Formula: lf, Proof: proof}
}

// helper: make a LabeledFormula with a label and a dummy non-nil formula body.
func makePropFormula(label string) *ast.LabeledFormula {
	return ast.NewLabeledFormula(ast.NewAtom(label), ast.NewAtom("body_"+label))
}

// proofNode creates a dummy ast.Node to use as a proof value in tests.
func proofNode(name string) ast.Node {
	return ast.NewAtom(name)
}

// ============================================================================
// Unit Tests
// ============================================================================

// Test 1: Direct proofs (non-nil formula) pass through unchanged.
func TestAttachProofs_DirectProofsPassThrough(t *testing.T) {
	mod := module.New()
	body := ast.NewAtom("body")
	pf := makeDirectProof("p1", body, proofNode("proof1"))
	mod.Proofs = []module.ProofEntry{pf}

	err := AttachProofs(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mod.Proofs) != 1 {
		t.Fatalf("expected 1 proof, got %d", len(mod.Proofs))
	}
	if mod.Proofs[0].Proof.String() != "proof1" {
		t.Errorf("expected proof1, got %v", mod.Proofs[0].Proof)
	}
}

// Test 2: Label-only proof matching a property.
func TestAttachProofs_LabelMatchesProperty(t *testing.T) {
	mod := module.New()
	prop := makePropFormula("myprop")
	mod.LabeledProps = []*ast.LabeledFormula{prop}
	mod.Proofs = []module.ProofEntry{makeLabelOnlyProof("myprop", proofNode("proof_obj"))}

	err := AttachProofs(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mod.Proofs) != 1 {
		t.Fatalf("expected 1 proof, got %d", len(mod.Proofs))
	}
	// The formula should be the property's LabeledFormula, not the proof's.
	if mod.Proofs[0].Formula != prop {
		t.Errorf("expected formula to be the property LabeledFormula")
	}
	if mod.Proofs[0].Proof.String() != "proof_obj" {
		t.Errorf("expected proof_obj, got %v", mod.Proofs[0].Proof)
	}
}

// Test 3: Label-only proof matching a conjecture.
func TestAttachProofs_LabelMatchesConjecture(t *testing.T) {
	mod := module.New()
	conj := makePropFormula("myconj")
	mod.LabeledConjs = []*ast.LabeledFormula{conj}
	mod.Proofs = []module.ProofEntry{makeLabelOnlyProof("myconj", proofNode("conj_proof"))}

	err := AttachProofs(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mod.Proofs) != 1 {
		t.Fatalf("expected 1 proof, got %d", len(mod.Proofs))
	}
	if mod.Proofs[0].Formula != conj {
		t.Errorf("expected formula to be the conjecture LabeledFormula")
	}
}

// Test 4: Label-only proof matching an isolate.
func TestAttachProofs_LabelMatchesIsolate(t *testing.T) {
	mod := module.New()
	mod.Isolates["myiso"] = &ast.IsolateDef{}
	mod.Proofs = []module.ProofEntry{makeLabelOnlyProof("myiso", proofNode("iso_proof"))}

	err := AttachProofs(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mod.Proofs) != 0 {
		t.Fatalf("expected 0 proofs in mod.Proofs, got %d", len(mod.Proofs))
	}
	if mod.IsolateProofs["myiso"] == nil {
		t.Errorf("expected iso_proof in IsolateProofs, got nil")
	}
}

// Test 5: Duplicate label → error.
func TestAttachProofs_DuplicateLabelErrors(t *testing.T) {
	mod := module.New()
	prop := makePropFormula("dup")
	mod.LabeledProps = []*ast.LabeledFormula{prop}
	mod.Proofs = []module.ProofEntry{
		makeLabelOnlyProof("dup", proofNode("proof1")),
		makeLabelOnlyProof("dup", proofNode("proof2")),
	}

	err := AttachProofs(mod)
	if err == nil {
		t.Fatal("expected error for duplicate label, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected error to contain 'duplicate', got: %s", err.Error())
	}
}

// Test 6: Unmatched label → error.
func TestAttachProofs_UnmatchedLabelErrors(t *testing.T) {
	mod := module.New()
	mod.Proofs = []module.ProofEntry{makeLabelOnlyProof("nonexistent", proofNode("proof"))}

	err := AttachProofs(mod)
	if err == nil {
		t.Fatal("expected error for unmatched label, got nil")
	}
	if !strings.Contains(err.Error(), "no property") {
		t.Errorf("expected error to contain 'no property', got: %s", err.Error())
	}
}

// Test 7: Empty label → error.
func TestAttachProofs_EmptyLabelErrors(t *testing.T) {
	mod := module.New()
	// Create a label-only proof with an empty-string label.
	lf := ast.NewLabeledFormula(ast.NewAtom(""), nil)
	mod.Proofs = []module.ProofEntry{{Formula: lf, Proof: proofNode("proof")}}

	err := AttachProofs(mod)
	if err == nil {
		t.Fatal("expected error for empty label, got nil")
	}
	if !strings.Contains(err.Error(), "empty label") {
		t.Errorf("expected error to contain 'empty label', got: %s", err.Error())
	}
}

// Test 8: Mix of direct and label-only proofs.
func TestAttachProofs_MixedDirectAndLabeled(t *testing.T) {
	mod := module.New()
	prop := makePropFormula("labeled_prop")
	mod.LabeledProps = []*ast.LabeledFormula{prop}

	body := ast.NewAtom("direct_body")
	mod.Proofs = []module.ProofEntry{
		makeDirectProof("direct1", body, proofNode("dproof")),
		makeLabelOnlyProof("labeled_prop", proofNode("lproof")),
	}

	err := AttachProofs(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mod.Proofs) != 2 {
		t.Fatalf("expected 2 proofs, got %d", len(mod.Proofs))
	}
	// First should be the direct proof.
	if mod.Proofs[0].Proof.String() != "dproof" {
		t.Errorf("expected first proof to be direct, got %v", mod.Proofs[0].Proof)
	}
	// Second should be the labeled proof attached to the property.
	if mod.Proofs[1].Formula != prop {
		t.Errorf("expected second proof formula to be the property")
	}
	if mod.Proofs[1].Proof.String() != "lproof" {
		t.Errorf("expected second proof to be lproof, got %v", mod.Proofs[1].Proof)
	}
}

// Test 9: Empty proofs list → no error, empty output.
func TestAttachProofs_EmptyProofsList(t *testing.T) {
	mod := module.New()

	err := AttachProofs(mod)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mod.Proofs) != 0 {
		t.Fatalf("expected 0 proofs, got %d", len(mod.Proofs))
	}
}

// Test 10: Direct proof with label "X" blocks label-only proof for "X" (duplicate).
func TestAttachProofs_DirectProofLabelBlocksLabeledProof(t *testing.T) {
	mod := module.New()
	prop := makePropFormula("shared")
	mod.LabeledProps = []*ast.LabeledFormula{prop}

	body := ast.NewAtom("direct_body")
	mod.Proofs = []module.ProofEntry{
		makeDirectProof("shared", body, proofNode("direct_proof")),
		makeLabelOnlyProof("shared", proofNode("labeled_proof")),
	}

	err := AttachProofs(mod)
	if err == nil {
		t.Fatal("expected error for duplicate label (direct blocks labeled), got nil")
	}
	if !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("expected error to contain 'duplicate', got: %s", err.Error())
	}
}

// ============================================================================
// Fuzz Test
// ============================================================================

func FuzzAttachProofs(f *testing.F) {
	// Seed corpus
	f.Add(uint8(1), uint8(0), uint8(0), uint8(2), uint64(0xDEADBEEF))
	f.Add(uint8(0), uint8(1), uint8(1), uint8(3), uint64(0xCAFE1234))
	f.Add(uint8(3), uint8(3), uint8(2), uint8(5), uint64(0x12345678))
	f.Add(uint8(0), uint8(0), uint8(0), uint8(0), uint64(0))

	f.Fuzz(func(t *testing.T, nProps, nConjs, nIsos, nProofs uint8, seed uint64) {
		// Cap sizes
		if nProps > 5 {
			nProps = 5
		}
		if nConjs > 5 {
			nConjs = 5
		}
		if nIsos > 3 {
			nIsos = 3
		}
		if nProofs > 8 {
			nProofs = 8
		}

		mod := module.New()
		var allLabels []string

		// Create properties
		for i := uint8(0); i < nProps; i++ {
			name := string(rune('a'+i)) + "_prop"
			prop := makePropFormula(name)
			mod.LabeledProps = append(mod.LabeledProps, prop)
			allLabels = append(allLabels, name)
		}

		// Create conjectures
		for i := uint8(0); i < nConjs; i++ {
			name := string(rune('a'+i)) + "_conj"
			conj := makePropFormula(name)
			mod.LabeledConjs = append(mod.LabeledConjs, conj)
			allLabels = append(allLabels, name)
		}

		// Create isolates
		for i := uint8(0); i < nIsos; i++ {
			name := string(rune('a'+i)) + "_iso"
			mod.Isolates[name] = &ast.IsolateDef{}
			allLabels = append(allLabels, name)
		}

		// Create proofs using seed bits to determine type and label
		rng := seed
		for i := uint8(0); i < nProofs; i++ {
			isDirect := (rng & 1) == 1
			rng >>= 1
			labelIdx := rng % uint64(len(allLabels)+2) // +2 to sometimes pick invalid/empty
			rng = rng*6364136223846793005 + 1442695040888963407

			var label string
			if int(labelIdx) < len(allLabels) {
				label = allLabels[labelIdx]
			} else if labelIdx == uint64(len(allLabels)) {
				label = "unknown_label"
			} else {
				label = "" // empty label
			}

			if isDirect {
				body := ast.NewAtom("body")
				mod.Proofs = append(mod.Proofs, makeDirectProof(label, body, proofNode("proof")))
			} else {
				mod.Proofs = append(mod.Proofs, makeLabelOnlyProof(label, proofNode("proof")))
			}
		}

		// The function must not panic.
		err := AttachProofs(mod)

		// If no error, verify invariants.
		if err == nil {
			for _, pe := range mod.Proofs {
				if pe.Formula == nil {
					t.Error("nil Formula in output proof entry")
				}
				if pe.Formula != nil && pe.Formula.Formula == nil {
					t.Error("nil Formula.Formula in output proof entry")
				}
			}
			for k := range mod.IsolateProofs {
				if _, ok := mod.Isolates[k]; !ok {
					t.Errorf("IsolateProofs key %q is not a valid isolate", k)
				}
			}
		} else {
			// If error, verify it's an IvyError.
			if _, ok := err.(*lg.IvyError); !ok {
				t.Errorf("expected *lg.IvyError, got %T", err)
			}
		}
	})
}
