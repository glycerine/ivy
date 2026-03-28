package module

import (
	"github.com/glycerine/goivy/ast"
	iu "github.com/glycerine/goivy/ivyutils"
	lg "github.com/glycerine/goivy/logic"
)

// Config for check, but module is lower in the import graph.
// - module doesn't import check
// - check imports module (one-way)
// - compiler, isolate, interp, actions, webui — all already import module
type Config struct {

	// "" means use embeded stdlib files, otherwise
	// look for the include/ directory here:
	IncludePathStdlib string

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

	// OptMutax controls whether mutable-axiom checking is enabled.
	// When true (non-default), axiom symbols are allowed to be modified by actions.
	// Corresponds to Python's opt_mutax = iu.BooleanParameter("mutax", False).
	// Moved from compiler.OptMutax.
	OptMutax bool `json:"mutax"`

	// AdmitDefinitionFactory creates an AdmitDefinitionFn for a given module.
	// Set by packages that can import both compiler and proof (e.g. check).
	// If nil, AdmitDefinition is skipped during compilation.
	// Moved from compiler.AdmitDefinitionFactory.
	AdmitDefinitionFactory func(mod *Module) func(defn *ast.LabeledFormula, proof ast.Node) error `json:"-"`

	// GlobalIncluded tracks already-included module names to prevent
	// double-includes within a compilation session.
	// Corresponds to Python's stack-based included check.
	GlobalIncluded map[string]bool `json:"-"`

	// AstCfg is the ast config for this session.
	AstCfg *ast.AstConfig `json:"-"`

	// IuCfg is the ivyutils config for this session.
	IuCfg *iu.IvyUtilsConfig `json:"-"`

	// --- Fields moved from package-level globals for multi-tenancy ---

	// UsedSorry tracks whether the 'sorry' tactic was used. (from tactics.UsedSorry)
	UsedSorry bool `json:"-"`
	// HandleRangeSorts enables range sort clamped arithmetic. (from solver.HandleRangeSorts)
	HandleRangeSorts bool `json:"handle_range_sorts"`
	// ExtAction is the combined external action name. (from isolate.ExtAction)
	ExtAction string `json:"-"`
	// RankingDebug enables ranking debug output. (from ranking.Debug)
	RankingDebug bool `json:"-"`
	// L2SDebug enables l2s debug output. (from l2s.Debug)
	L2SDebug bool `json:"-"`
	// ComposeDebug enables compose tactic debug output. (from compose.Debug)
	ComposeDebug bool `json:"-"`
	// AlphaTestBottom controls alpha bottom testing. (from alpha.TestBottom)
	AlphaTestBottom bool `json:"-"`
	// AlphaLog controls alpha logging. (from alpha.Log)
	AlphaLog bool `json:"-"`
	// AutoinstVerbose enables autoinstance verbose output. (from autoinst.Verbose)
	AutoinstVerbose bool `json:"-"`
	// TraceDetailed enables detailed trace information. (from trace.OptionDetailed)
	TraceDetailed bool `json:"-"`
	// MCVerbose enables verbose MC output. (from mc.Verbose)
	MCVerbose bool `json:"-"`
	// OptionAbsInit controls whether the initial state is abstracted. (from art.OptionAbsInit)
	OptionAbsInit bool `json:"-"`
	// VMTVerbose enables verbose VMT output. (from vmt.Verbose)
	VMTVerbose bool `json:"-"`
	// CheckLineno is the lineno filter for check/mc/vmt. (from check.CheckLineno, mc.CheckedAssert, vmt.CheckedAssertValue)
	CheckLineno string `json:"-"`
	// Failures counts verification failures. (from check.Failures)
	Failures int `json:"-"`
	// CheckedActionFound tracks whether a checked action was found. (from check.CheckedActionFound)
	CheckedActionFound bool `json:"-"`

	// IsolateCfg holds per-session isolate configuration.
	IsolateCfg *IsolateConfig `json:"-"`
}

// IsolateConfig holds per-session isolate configuration. Replaces former
// package-level globals in isolate/ for multi-tenancy safety.
// Defined in module/ to avoid a circular import (isolate imports module).
type IsolateConfig struct {
	ShowCompiled          bool
	ConeOfInfluence       bool
	FilterSymbols         bool
	CreateImports         bool
	EnforceAxioms         bool
	DoCheckInterference   bool
	Pedantic              bool
	PreferImpls           bool
	KeepDestructors       bool
	IsolateMode           string
	CompileWithInvariants bool
	AssumeInvariants      bool
	InterpretAllSorts     bool
	NumIsolateParams      int
	StripAddedSymbols     []*lg.Symbol
	VPrivates             map[string]bool
	IvyVersion            string
	ExtAction             string
}

// NewIsolateConfig creates a fresh IsolateConfig with defaults matching Python.
func NewIsolateConfig() *IsolateConfig {
	return &IsolateConfig{
		ConeOfInfluence:     true,
		FilterSymbols:       true,
		DoCheckInterference: true,
		IsolateMode:         "check",
		AssumeInvariants:    true,
		VPrivates:           make(map[string]bool),
		IvyVersion:          "1.7",
	}
}

func NewConfig() *Config {
	iuCfg := iu.NewIvyUtilsConfig()
	astCfg := ast.NewAstConfig()
	astCfg.IuCfg = iuCfg
	return &Config{
		Coverage:         true,
		MacroFinder:      true,  // Python default: islv.opt_macro_finder defaults to true
		GlobalIncluded:   make(map[string]bool),
		AstCfg:           astCfg,
		IuCfg:            iuCfg,
		IsolateCfg:       NewIsolateConfig(),
		HandleRangeSorts: true,  // default matches solver.HandleRangeSorts = true
		AlphaTestBottom:  true,  // default matches alpha.TestBottom = true
		AutoinstVerbose:  true,  // default matches autoinst.Verbose = true
		TraceDetailed:    true,  // default matches trace.OptionDetailed = true
	}
}
