package module

import (
	"github.com/glycerine/ivy/goivy/ast"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
)

// GuiArtHook is the concrete type for the analysis-graph GUI hook stored on
// Config.GuiArtHook. It is invoked by check.GuiArt to display an analysis
// graph in an interactive UI.
//
// The `target` argument is interface{} to mirror Python's gui_art polymorphism:
// callers pass either an *art.AnalysisGraph (from the ShowCounterexample /
// DisplayCex paths) or a *check.MatchHandler (from the trace failure path).
// We cannot name those types here because module cannot import art or check
// (cycle), so target stays interface{} and the hook implementation type-switches.
//
// The `isCti` argument carries the failing-conjecture clauses captured by
// check.MatchHandler.IsCti, or nil for non-CTI counterexamples.
//
// The hook is responsible for any blocking UI loop and may call os.Exit if
// it wishes to mirror Python's `exit(1)` at the end of gui_art
// (ivy_check.py:102).
type GuiArtHook func(mod *Module, target interface{}, isCti *Clauses) error

// Config for check, but module is lower in the import graph.
// - module doesn't import check
// - check imports module (one-way)
// - compiler, isolate, interp, actions, webui — all already import module
type Config struct {

	// "" means use embeded stdlib files, otherwise
	// look for the include/ directory here:
	IncludePathStdlib string

	// module; more internal specific
	NewProofCheckerFn func(mod *Module, axioms, definitions []*ast.LabeledFormula, schemata *iu.InsMap[string, *ast.LabeledFormula]) ProofCheckerInterface

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

	// SomeBounded tracks whether any isolate in this session used the BMC
	// (bounded model checking) verification method. When true, Start prints
	// "BOUNDED" before "OK". Mirrors Python's module-level `some_bounded`
	// global (ivy_check.py:996, 1029, 1046).
	SomeBounded bool

	// GuiArtHook is a closure that displays an analysis graph in a UI. When
	// nil, check.GuiArt prints diagnostic info and returns. Mirrors Python's
	// gui_art delegation to tk_ui.new_ui (ivy_check.py:86-102). The type is
	// defined in this package (above) so the field is fully typed without
	// interface{} boxing.
	GuiArtHook GuiArtHook `json:"-"`

	// CheckedActionFound tracks whether a checked action was found.
	CheckedActionFound bool

	// CheckLineno is the current line number being checked, or empty for all.
	CheckLineno string

	// SolverOpts controls per-solver Z3 behavior.
	// Moved from solver.Options so the full config is accessible
	// from module.Config without import cycles.
	SolverOpts *SolverOptions

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
	// McIteCtr is the counter for ITE elimination fresh variables. (from mc.iteCtr)
	McIteCtr int64 `json:"-"`
	// FullQI controls full quantifier instantiation for transition QE.
	// Python: fullqi = iu.BooleanParameter("fullqi", False) at ivy_mc.py:36
	FullQI bool `json:"-"`
	// OptionAbsInit controls whether the initial state is abstracted. (from art.OptionAbsInit)
	OptionAbsInit bool `json:"-"`
	// VMTVerbose enables verbose VMT output. (from vmt.Verbose)
	VMTVerbose bool `json:"-"`

	// IsolateCfg holds per-session isolate configuration.
	IsolateCfg *IsolateConfig `json:"-"`

	// ActCfg holds the per-session actions config.
	// Thread it to action construction sites so LocalAction (etc.) counters
	// match Python's single global local_action_ctr.
	ActCfg *ModuleActionsConfig `json:"-"`

	// ProofCfg holds the per-session proof configuration (tactic registry).
	ProofCfg *ProofConfig `json:"-"`

	// WebUIConformCheck lets the Go backend know it is being
	// compared to the python so it can omit the new extra z3_contacted flag
	// and not trigger a spurious mismatch report.
	WebUIConformCheck bool
}

// SolverOptions controls per-solver Z3 behavior.
// Moved from solver.Options so the full config is accessible
// from module.Config without import cycles.
type SolverOptions struct {
	Seed        int
	Incremental bool
	MacroFinder bool
	ShowVCs     bool
	UseZ3Enums  bool
}

// DefaultSolverOptions returns the default solver options.
func DefaultSolverOptions() *SolverOptions {
	return &SolverOptions{
		Seed:        0,
		Incremental: true,
		MacroFinder: true,
		ShowVCs:     false,
		UseZ3Enums:  true,
	}
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
	StripAddedSymbols     []*lg.Const
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
		ActCfg:           NewModuleActionsConfig(),
		Coverage:         true,
		SolverOpts:       DefaultSolverOptions(),
		GlobalIncluded:   make(map[string]bool),
		AstCfg:           astCfg,
		IuCfg:            iuCfg,
		IsolateCfg:       NewIsolateConfig(),
		HandleRangeSorts: true, // default matches solver.HandleRangeSorts = true
		AlphaTestBottom:  true, // default matches alpha.TestBottom = true
		L2SDebug:         true, // default true so prints match Python's unconditional output
		AutoinstVerbose:  true, // default matches autoinst.Verbose = true
		TraceDetailed:    true, // default matches trace.OptionDetailed = true
		ProofCfg:         TacticNewConfig(),
	}
}

// --- ModuleActionContext ---

// ModuleIActionContext is the interface for action contexts, matching Python's
// ModuleActionContext class hierarchy (ModuleActionContext, UnrollContext, TypeCheckContext).
type ModuleIActionContext interface {
	GetDomain() *Module
	Get(symbol string) Action
	Enter()
	Exit()
}

// ModuleActionsConfig holds per-session actions state.
type ModuleActionsConfig struct {
	Context     ModuleIActionContext
	Determinize bool
	// SymexParams is the current symbolic execution parameter list.
	// Corresponds to Python's module-level symex_params in ivy_actions.py.
	SymexParams []lg.Expr

	// IuCfg is the per-session ivyutils config, shared with AstConfig.
	// LocalActionCtr lives on IuCfg so both ast and actions use the same counter.
	IuCfg *iu.IvyUtilsConfig
}

// NewModuleActionsConfig creates a new ModuleActionsConfig with a default ModuleActionContext.
func NewModuleActionsConfig() *ModuleActionsConfig {
	return &ModuleActionsConfig{Context: &ModuleActionContext{}, IuCfg: iu.NewIvyUtilsConfig()}
}

// ModuleActionContext provides context for evaluating states and actions.
// Corresponds to Python's ModuleActionContext class with __enter__/__exit__.
type ModuleActionContext struct {
	Domain     *Module              // module reference (Python: self.domain)
	OldContext ModuleIActionContext // saved context for restore on Exit
	Cfg        *ModuleActionsConfig // config this context belongs to
}

func NewModuleActionContext(domain *Module) *ModuleActionContext {
	return &ModuleActionContext{Domain: domain}
}

// NewModuleActionContextOn creates an ModuleActionContext bound to a specific ModuleActionsConfig.
func NewModuleActionContextOn(domain *Module, cfg *ModuleActionsConfig) *ModuleActionContext {
	return &ModuleActionContext{Domain: domain, Cfg: cfg}
}

func (ac *ModuleActionContext) GetDomain() *Module { return ac.Domain }

// Get resolves an action symbol. Corresponds to Python's ModuleActionContext.get
// which delegates to ivy_module.find_action.
func (ac *ModuleActionContext) Get(symbol string) Action {
	if ac.Domain != nil {
		if found, ok := ac.Domain.FindAction(symbol); ok {
			return found
		}
	}
	return nil
}

// Enter implements Python's ModuleActionContext.__enter__: saves the old context
// from the ModuleActionsConfig and installs this one.
// Update: we mostly pass this as a bool argument on the call stack now, for clarity.
func (ac *ModuleActionContext) Enter() {
	if ac.Cfg == nil {
		panic("ModuleActionContext.Enter: Cfg is nil — use NewModuleActionContextOn or set Cfg before calling Enter")
	}
	ac.OldContext = ac.Cfg.Context
	ac.Cfg.Context = ac
}

// Exit implements Python's ModuleActionContext.__exit__: restores the previous context.
func (ac *ModuleActionContext) Exit() {
	if ac.Cfg == nil {
		panic("ModuleActionContext.Exit: Cfg is nil")
	}
	ac.Cfg.Context = ac.OldContext
}

// RunWithModuleActionContext executes fn within this context, ensuring Exit is called.
func RunWithModuleActionContext(ctx ModuleIActionContext, fn func()) {
	ctx.Enter()
	defer ctx.Exit()
	fn()
}
