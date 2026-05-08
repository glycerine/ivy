package goivy

import (
	"fmt"
	"strings"
)

const (
	BackendCGo       = "cgo"
	BackendWazero    = "wazero"
	BackendJSBrowser = "jsbrowser"
)

// GuiArtHook is the type for the analysis-graph GUI hook stored on
// Config.GuiArtHook. It is invoked by GuiArt to display an analysis graph in
// an interactive UI.
//
// The `target` argument is interface{} to mirror Python's gui_art polymorphism:
// callers pass either an *AnalysisGraph (from the ShowCounterexample /
// DisplayCex paths) or a *Trace (from the trace failure path).
// target stays interface{} because the Python entry point accepts both shapes,
// and the hook implementation type-switches on the concrete value.
//
// The `isCti` argument carries the failing-conjecture clauses captured by the
// trace failure path, or nil for non-CTI counterexamples.
//
// The hook is responsible for any blocking UI loop and may call os.Exit if
// it wishes to mirror Python's `exit(1)` at the end of gui_art
// (ivy_check.py:102).
type GuiArtHook func(mod *Module, target interface{}, isCti *Clauses) error

// Z3Backend is the single runtime-selected Z3 implementation boundary. The
// method names intentionally mention Z3 so future non-Z3 SMT backends do not
// get hidden behind generic solver names.
type Z3Backend interface {
	Z3BackendName() string
	NewZ3Context() *Z3Context
	NewInterpolationZ3Context() *Z3Context
	NewZ3Solver(ctx *Z3Context) *Z3Solver
}

// Config holds per-session settings that Python Ivy keeps in module-level
// parameters/globals. This is the main intentional architectural difference
// from Python: Go threads explicit config instead of mutable process globals.
type Config struct {

	// "" means use embeded stdlib files, otherwise
	// look for the include/ directory here:
	IncludePathStdlib string

	// BackendName is the serializable/debuggable Z3 backend name. Backend is
	// the actual runtime object selected from this name at process edges.
	BackendName string    `json:"backend"`
	Backend     Z3Backend `json:"-"`

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

	// CompleteLogic is the comma-separated logic parameter (Python: param_logic).
	// Default is "" meaning use il.DefaultLogics. Set via CLI --complete flag
	// or programmatically. Corresponds to Python's iu.Parameter("complete", ...).
	CompleteLogic string `json:"complete"`

	// OptMutax controls whether mutable-axiom checking is enabled.
	// When true (non-default), axiom symbols are allowed to be modified by actions.
	// Corresponds to Python's opt_mutax = iu.BooleanParameter("mutax", False).
	// Moved from compiler.OptMutax.
	OptMutax bool `json:"mutax"`

	// GlobalIncluded tracks already-included module names to prevent
	// double-includes within a compilation session.
	// Corresponds to Python's stack-based included check.
	GlobalIncluded map[string]bool `json:"-"`

	// AstCfg is the ast config for this session.
	AstCfg *AstConfig `json:"-"`

	// IuCfg is the ivyutils config for this session.
	IuCfg *IvyUtilsConfig `json:"-"`

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
	ActCfg *ActionsConfig `json:"-"`

	// ProofCfg holds the per-session proof configuration (tactic registry).
	ProofCfg *ProofConfig `json:"-"`

	// WebUIConformCheck lets the Go backend know it is being
	// compared to the python so it can omit the new extra z3_contacted flag
	// and not trigger a spurious mismatch report.
	WebUIConformCheck bool
}

// SolverOptions controls per-solver Z3 behavior.
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

// IsolateConfig holds per-session isolate configuration. It replaces
// Python-style package-level globals for multi-tenancy safety.
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
	StripAddedSymbols     []*Const
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
	iuCfg := NewIvyUtilsConfig()
	astCfg := NewAstConfig()
	astCfg.IuCfg = iuCfg
	actCfg := NewActionsConfig()
	actCfg.IuCfg = iuCfg
	backend := defaultZ3Backend()
	return &Config{
		ActCfg:           actCfg,
		BackendName:      backend.Z3BackendName(),
		Backend:          backend,
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

// NormalizeBackendName returns the canonical backend name and treats "" as the
// default native CGo backend for compatibility with older serialized configs.
func NormalizeBackendName(backend string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(backend)) {
	case "", BackendCGo:
		return BackendCGo, nil
	case BackendWazero:
		return BackendWazero, nil
	case BackendJSBrowser:
		return BackendJSBrowser, nil
	default:
		return "", fmt.Errorf("unknown backend %q; expected %q, %q, or %q", backend, BackendCGo, BackendWazero, BackendJSBrowser)
	}
}

func mustNewZ3BackendByName(name string) Z3Backend {
	backend, err := NewZ3BackendByName(name)
	if err != nil {
		panic(err)
	}
	return backend
}

// NewZ3BackendByName creates the backend object selected by name.
func NewZ3BackendByName(name string) (Z3Backend, error) {
	backendName, err := NormalizeBackendName(name)
	if err != nil {
		return nil, err
	}
	return newZ3BackendByCanonicalName(backendName), nil
}

// SetBackendName resolves name immediately into cfg.Backend.
func (cfg *Config) SetBackendName(name string) error {
	backend, err := NewZ3BackendByName(name)
	if err != nil {
		return err
	}
	cfg.BackendName = backend.Z3BackendName()
	cfg.Backend = backend
	return nil
}

// ResolveBackend ensures cfg.Backend is populated. It preserves a manually
// injected backend object and fills BackendName from it when needed.
func (cfg *Config) ResolveBackend() error {
	if cfg.Backend != nil {
		if cfg.BackendName == "" {
			cfg.BackendName = cfg.Backend.Z3BackendName()
			return nil
		}
		backendName, err := NormalizeBackendName(cfg.BackendName)
		if err != nil {
			return err
		}
		if backendName == cfg.Backend.Z3BackendName() {
			cfg.BackendName = backendName
			return nil
		}
	}
	return cfg.SetBackendName(cfg.BackendName)
}

func z3BackendForConfig(cfg *Config) Z3Backend {
	if cfg == nil {
		return defaultZ3Backend()
	}
	if err := cfg.ResolveBackend(); err != nil {
		panic(err)
	}
	if cfg.Backend == nil {
		return defaultZ3Backend()
	}
	return cfg.Backend
}

// --- ActionContext ---

// IActionContext is the interface for action contexts, matching Python's
// ActionContext class hierarchy (ActionContext, UnrollContext, TypeCheckContext).
type IActionContext interface {
	GetDomain() *Module
	Get(symbol string) Action
	Enter()
	Exit()
}

// ActionsConfig holds per-session actions state.
type ActionsConfig struct {
	Context     IActionContext
	Determinize bool
	// SymexParams is the current symbolic execution parameter list.
	// Corresponds to Python's module-level symex_params in ivy_actions.py.
	SymexParams []Expr

	// IuCfg is the per-session ivyutils config, shared with AstConfig.
	// LocalActionCtr lives on IuCfg so both ast and actions use the same counter.
	IuCfg *IvyUtilsConfig
}

// NewActionsConfig creates a new ActionsConfig with a default ActionContext.
func NewActionsConfig() *ActionsConfig {
	return &ActionsConfig{Context: &ActionContext{}, IuCfg: NewIvyUtilsConfig()}
}

// ActionContext provides context for evaluating states and actions.
// Corresponds to Python's ActionContext class with __enter__/__exit__.
type ActionContext struct {
	Domain     *Module        // module reference (Python: self.domain)
	OldContext IActionContext // saved context for restore on Exit
	Cfg        *ActionsConfig // config this context belongs to
}

func NewActionContext(domain *Module) *ActionContext {
	return &ActionContext{Domain: domain}
}

// NewActionContextOn creates an ActionContext bound to a specific ActionsConfig.
func NewActionContextOn(domain *Module, cfg *ActionsConfig) *ActionContext {
	return &ActionContext{Domain: domain, Cfg: cfg}
}

func (ac *ActionContext) GetDomain() *Module { return ac.Domain }

// Get resolves an action symbol. Corresponds to Python's ActionContext.get
// which delegates to ivy_module.find_action.
func (ac *ActionContext) Get(symbol string) Action {
	if ac.Domain != nil {
		if found, ok := ac.Domain.FindAction(symbol); ok {
			return found
		}
	}
	return nil
}

// Enter implements Python's ActionContext.__enter__: saves the old context
// from the ActionsConfig and installs this one.
// Update: we mostly pass this as a bool argument on the call stack now, for clarity.
func (ac *ActionContext) Enter() {
	if ac.Cfg == nil {
		panic("ActionContext.Enter: Cfg is nil — use NewActionContextOn or set Cfg before calling Enter")
	}
	ac.OldContext = ac.Cfg.Context
	ac.Cfg.Context = ac
}

// Exit implements Python's ActionContext.__exit__: restores the previous context.
func (ac *ActionContext) Exit() {
	if ac.Cfg == nil {
		panic("ActionContext.Exit: Cfg is nil")
	}
	ac.Cfg.Context = ac.OldContext
}

// RunWithActionContext executes fn within this context, ensuring Exit is called.
func RunWithActionContext(ctx IActionContext, fn func()) {
	ctx.Enter()
	defer ctx.Exit()
	fn()
}
