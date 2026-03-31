// proofapi.go defines the proof checker interface, tactic types, and
// proof configuration that both the compiler and proof packages need.
// Living in module breaks the compiler <-> proof import cycle:
// proof imports module, compiler imports module.
package module

import (
	"github.com/glycerine/goivy/ast"
)

// ProofCheckerInterface abstracts the proof checker methods needed by the compiler
// and external tactic implementations. proof.ProofChecker satisfies this interface.
type ProofCheckerInterface interface {
	AdmitDefinition(defn *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)
	AdmitProposition(prop *ast.LabeledFormula, proof ast.Node, existingSubgoals ...*ast.LabeledFormula) ([]*ast.LabeledFormula, error)
	GetSubgoals(prop *ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)
	// SetLastAxiom updates the last admitted axiom (for named_trans).
	SetLastAxiom(prop *ast.LabeledFormula)
	// SetSchema updates a schema entry.
	SetSchema(name string, prop *ast.LabeledFormula)
	// GetModule returns the current module.
	GetModule() *Module
	// GetAstCfg returns the AST configuration.
	GetAstCfg() *ast.AstConfig
	// GetAxioms returns the list of available axioms.
	GetAxioms() []*ast.LabeledFormula
}

// ProofTactic is a function that applies a proof tactic to a goal,
// producing subgoals or an error.
type ProofTactic func(checker ProofCheckerInterface, goals []*ast.LabeledFormula, proof ast.Node) ([]*ast.LabeledFormula, error)

// ProofConfig holds per-session proof state (tactic registry).
type ProofConfig struct {
	Tactics map[string]ProofTactic
}

// TacticNewConfig creates a new ProofConfig with an empty tactic registry.
func TacticNewConfig() *ProofConfig {
	return &ProofConfig{Tactics: make(map[string]ProofTactic)}
}

// RegisterTactic registers a named tactic on this config.
func (cfg *ProofConfig) RegisterTactic(name string, t ProofTactic) {
	cfg.Tactics[name] = t
}
