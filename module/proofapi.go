// proofapi.go defines the proof checker interface and factory functions
// that both the compiler and proof packages need. Living in module breaks
// the compiler ↔ proof import cycle: proof imports module, compiler imports module.
package module

import (
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
)

// ProofCheckerInterface abstracts the proof checker methods needed by the compiler.
// proof.ProofChecker satisfies this interface.
type ProofCheckerInterface interface {
	AdmitProposition(prop *ast.LabeledFormula, proof ast.Node, existingSubgoals ...*ast.LabeledFormula) ([]*ast.LabeledFormula, error)
	GetSubgoals(prop *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)
	// SetLastAxiom updates the last admitted axiom (for named_trans).
	SetLastAxiom(prop *ast.LabeledFormula)
	// SetSchema updates a schema entry.
	SetSchema(name string, prop *ast.LabeledFormula)
}

// NewProofCheckerFn is set by the proof package's init() to avoid circular imports.
// It creates a new ProofChecker.
var NewProofCheckerFn func(axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) ProofCheckerInterface

// GoalConcFn extracts the conclusion of a goal. Set by proof package to avoid cycle.
var GoalConcFn func(g *ast.LabeledFormula) lg.Expr
