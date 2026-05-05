// proofapi.go defines the proof checker interface, tactic types, and
// proof configuration that both the compiler and proof packages need.
// Living in module breaks the compiler <-> proof import cycle:
// proof imports module, compiler imports module.
package goivy

// ProofCheckerInterface abstracts the proof checker methods needed by the compiler
// and external tactic implementations. proof.ProofChecker satisfies this interface.
type ProofCheckerInterface interface {
	AdmitDefinition(defn *LabeledFormula, proof Node) ([]*LabeledFormula, error)
	AdmitProposition(prop *LabeledFormula, proof Node, existingSubgoals ...*LabeledFormula) ([]*LabeledFormula, error)
	GetSubgoals(prop *LabeledFormula, proof Node) ([]*LabeledFormula, error)
	// SetLastAxiom updates the last admitted axiom (for named_trans).
	SetLastAxiom(prop *LabeledFormula)
	// SetSchema updates a schema entry.
	SetSchema(name string, prop *LabeledFormula)
	// GetModule returns the current module.
	GetModule() *Module
	// GetAstCfg returns the AST configuration.
	GetAstCfg() *AstConfig
	// GetAxioms returns the list of available axioms.
	GetAxioms() []*LabeledFormula
}

// ProofTactic is a function that applies a proof tactic to a goal,
// producing subgoals or an error.
type ProofTactic func(checker ProofCheckerInterface, goals []*LabeledFormula, proof Node) ([]*LabeledFormula, error)

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
