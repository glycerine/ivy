// proofapi.go defines the proof checker interface and factory functions
// that both the compiler and proof packages need. Living in module breaks
// the compiler ↔ proof import cycle: proof imports module, compiler imports module.
package module

import (
	"github.com/glycerine/goivy/ast"
	//lg "github.com/glycerine/goivy/logic"
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

/*
// Config holds per-session module state: factory functions and current module.
type ModConfig struct {
	// NewProofCheckerFn creates a new ProofChecker. Set by proof.RegisterFactories.
	NewProofCheckerFn func(axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) ProofCheckerInterface

	// GoalConcFn extracts the conclusion of a goal. Set by proof.RegisterFactories.
	GoalConcFn func(g *ast.LabeledFormula) lg.Expr

	mu            sync.Mutex
	currentModule *Module
}


// NewConfig creates a new module Config.
func NewModConfig() *ModConfig {
	return &ModConfig{}
}

// CurrentModule returns the currently active module for this config, or nil.
func (cfg *ModConfig) CurrentModule() *Module {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	return cfg.currentModule
}

// SetCurrentModule sets the current module (used by Module.Enter/Exit).
func (cfg *ModConfig) SetCurrentModule(m *Module) {
	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	cfg.currentModule = m
}
*/
