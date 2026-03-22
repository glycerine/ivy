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

	// MacroFinder corresponds to Python's islv.opt_macro_finder.
	// When true, the Z3 macro finder is enabled (default true in solver).
	MacroFinder bool `json:"macro_finder"`

	// Isolate is the user-specified isolate to check.
	// Python: ivy_compiler.isolate.get()
	// If empty, all isolates are checked.
	Isolate string `json:"isolate"`

	// OptSeparateSet distinguishes "user explicitly set --separate" from
	// "not set". Python's opt_separate is BooleanParameter("separate", None)
	// — three-valued (None/True/False). In Go, OptSeparate is the value
	// and OptSeparateSet indicates whether it was explicitly provided.
	OptSeparateSet bool `json:"separate_set"`

	// SolverClearFn is called by Module.Enter() to clear cached Z3 values
	// when changing the active module/sig. Set by the solver package at
	// init time. Corresponds to Python's ivy_solver.clear() call in
	// Module.__enter__ (ivy_module.py:101).
	SolverClearFn func() `json:"-"`

	// CompleteLogic is the comma-separated logic parameter (Python: param_logic).
	// Default is "" meaning use il.DefaultLogics. Set via CLI --complete flag
	// or programmatically. Corresponds to Python's iu.Parameter("complete", ...).
	CompleteLogic string `json:"complete"`
}

func NewConfig() *Config {
	return &Config{
		Coverage:    true,
		MacroFinder: true, // Python default: islv.opt_macro_finder defaults to true
	}
}
