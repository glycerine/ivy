package module

import (
	"github.com/glycerine/goivy/ast"
	lg "github.com/glycerine/goivy/logic"
)

// Config for check, but module is lower in the import graph.
// - module doesn't import check
// - check imports module (one-way)
// - compiler, isolate, interp, actions, webui — all already import module
type Config struct {

	// module; more internal specific
	NewProofCheckerFn func(axioms, definitions []*ast.LabeledFormula, schemata map[string]*ast.LabeledFormula) ProofCheckerInterface

	// GoalConcFn extracts the conclusion of a goal.
	// Set by proof.RegisterFactories.
	GoalConcFn    func(g *ast.LabeledFormula) lg.Expr
	CurrentModule *Module

	// more for external clients like check/ sub package.
	Diagnose          bool   `json:"diagnose"`
	Coverage          bool   `json:"coverage"`
	CheckedAction     string `json:"action"`
	OptTrusted        bool   `json:"trusted"`
	OptMC             bool   `json:"mc"`
	OptTrace          bool   `json:"trace"`
	OptSeparate       bool   `json:"separate"`
	OptUncheckedProps string `json:"unchecked_properties"` // a filename
	OptIvyStats       bool   `json:"ivy_stats"`

	// string packed with comma-separated action-names
	PriorityActions string `json:"prioritize"` // ivy_check.py:195

	NoCheckGuarantees bool `json:"no_check_guarantees"`
	Profiling         bool `json:"profile"`
	OptSummary        bool `json:"summary"`

	// CheckUnprovable corresponds to Python's act.check_unprovable
	// (ivy_actions.py:25). When true, only unprovable assertions are checked.
	OnlyCheckUnprovable bool `json:"unprovable"`

	// Failures tracks the number of failed checks during verification.
	Failures int

	// CheckedActionFound tracks whether a checked action was found.
	CheckedActionFound bool

	// CheckLineno is the current line number being checked, or empty for all.
	CheckLineno string
}

func NewConfig() *Config {
	return &Config{
		Coverage: true,
	}
}
