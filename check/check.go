// Package check is the top-level verification driver for Ivy.
// It orchestrates checking of properties, conjectures, temporals,
// and isolates. This corresponds to Python's ivy_check.py.
package check

import (
	"fmt"
	"sort"
	"strings"

	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	iu "github.com/glycerine/goivy/ivyutils"
)

// --- Package-level parameters ---

var (
	Diagnose             = iu.NewBooleanParameter("diagnose", false)
	Coverage             = iu.NewBooleanParameter("coverage", true)
	CheckedAction        = iu.NewParameter("action", "")
	OptTrusted           = iu.NewBooleanParameter("trusted", false)
	OptMC                = iu.NewBooleanParameter("mc", false)
	OptTrace             = iu.NewBooleanParameter("trace", false)
	OptSeparate          = iu.NewParameter("separate", nil)
	OptUncheckedProps    = iu.NewParameter("unchecked_properties", nil)
	OptIvyStats          = iu.NewBooleanParameter("ivy_stats", false)
	PriorityActions      = iu.NewParameter("prioritize", nil)
	NoCheckGuarantees    = iu.NewBooleanParameter("no_check_guarantees", false)
	Profiling            = iu.NewBooleanParameter("profile", false)
	OptSummary           = iu.NewBooleanParameter("summary", false)
)

// Failures tracks the number of failed checks during verification.
var Failures int

// CheckedActionFound tracks whether a checked action was found.
var CheckedActionFound bool

// CheckLineno is the current line number being checked, or empty for all.
var CheckLineno string

// --- Checker interface and implementations ---

// Checker is the interface for verification condition checkers.
// Each checker wraps a formula condition to be checked against a state.
type Checker interface {
	// Cond returns the clause set representing the negated condition to check.
	Cond() *clauseops.Clauses
	// Start is called before checking begins (prints status).
	Start()
	// Sat is called when the SMT solver finds the condition satisfiable
	// (meaning the check fails for normal checks). Returns true to continue.
	Sat() bool
	// Unsat is called when the SMT solver finds the condition unsatisfiable
	// (meaning the check passes for normal checks). Returns true to continue.
	Unsat() bool
	// Assume returns true if this checker should be assumed (not checked).
	Assume() bool
	// Fail marks the check as failed.
	Fail() bool
	// Pass marks the check as passed.
	Pass() bool
	// GetAnnot returns the annotation for trace reconstruction, or nil.
	GetAnnot() interface{}
	// Failed returns whether this checker has failed.
	Failed() bool
	// GetLF returns the labeled formula if this is a conjecture checker.
	GetLF() *module.LabeledFormula
}

// --- BaseChecker ---

// BaseChecker is the default checker implementation.
// It wraps a formula conjecture and checks it against a state by
// negating it (dualizing) and checking satisfiability.
type BaseChecker struct {
	FC         *clauseops.Clauses
	ReportPass bool
	Inverted   bool
	FailedFlag bool
}

// NewBaseChecker creates a BaseChecker for the given conjecture formula.
// If invert is true (the default), the formula is dualized for checking.
func NewBaseChecker(conj lg.Node, reportPass bool, invert bool) *BaseChecker {
	fc := clauseops.FormulaToClauses(conj, nil)
	if invert {
		fc = DualClauses(fc)
	}
	return &BaseChecker{
		FC:         fc,
		ReportPass: reportPass,
		Inverted:   invert,
	}
}

func (c *BaseChecker) Cond() *clauseops.Clauses { return c.FC }
func (c *BaseChecker) Start() {
	if c.ReportPass {
		fmt.Print("... ")
	}
}
func (c *BaseChecker) Sat() bool    { return c.Fail() }
func (c *BaseChecker) Unsat() bool  { return c.Pass() }
func (c *BaseChecker) Assume() bool { return false }
func (c *BaseChecker) GetAnnot() interface{} { return nil }
func (c *BaseChecker) Failed() bool { return c.FailedFlag }
func (c *BaseChecker) GetLF() *module.LabeledFormula { return nil }

func (c *BaseChecker) Fail() bool {
	fmt.Println("FAIL")
	Failures++
	c.FailedFlag = true
	// Ignore failures if not diagnosing
	return !(Diagnose.GetBool() || OptTrace.GetBool())
}

func (c *BaseChecker) Pass() bool {
	if c.ReportPass {
		fmt.Println("PASS")
	}
	return true
}

// --- ConjChecker ---

// ConjChecker checks a single labeled conjecture formula.
type ConjChecker struct {
	BaseChecker
	LF     *module.LabeledFormula
	Indent int
}

// NewConjChecker creates a ConjChecker for the given labeled formula.
func NewConjChecker(lf *module.LabeledFormula, indent int) *ConjChecker {
	base := NewBaseChecker(lf.Formula, true, true)
	return &ConjChecker{
		BaseChecker: *base,
		LF:          lf,
		Indent:      indent,
	}
}

func (c *ConjChecker) Start() {
	fmt.Print(PrettyLF(c.LF, c.Indent), " ")
	fmt.Print("... ")
}

func (c *ConjChecker) GetAnnot() interface{} {
	return nil // Stub: annotations not yet ported
}

func (c *ConjChecker) GetLF() *module.LabeledFormula { return c.LF }

// --- ConjAssumer ---

// ConjAssumer treats a conjecture as assumed (not checked).
type ConjAssumer struct {
	BaseChecker
	LF *module.LabeledFormula
}

// NewConjAssumer creates a ConjAssumer for the given labeled formula.
func NewConjAssumer(lf *module.LabeledFormula) *ConjAssumer {
	base := NewBaseChecker(lf.Formula, false, false)
	return &ConjAssumer{
		BaseChecker: *base,
		LF:          lf,
	}
}

func (c *ConjAssumer) Start() {
	fmt.Println(PrettyLF(c.LF, 8) + "  [assumed]")
}

func (c *ConjAssumer) Assume() bool { return true }
func (c *ConjAssumer) GetLF() *module.LabeledFormula { return c.LF }

// --- DualClauses ---

// DualClauses negates a clause set for checking: the negated
// clauses are satisfiable iff the original are not entailed.
// Skolem witnesses are created for existential variables.
// This is a simplified stub of Python's lut.dual_clauses.
func DualClauses(c *clauseops.Clauses) *clauseops.Clauses {
	if c == nil {
		return c
	}
	negated := make([]lg.Node, len(c.Fmlas))
	for i, f := range c.Fmlas {
		negated[i] = &lg.Not{Body: f}
	}
	if len(negated) == 0 {
		return clauseops.FalseClauses(c.Annot)
	}
	// Dual of And(f1, f2, ...) is Or(Not(f1), Not(f2), ...)
	return clauseops.NewClauses([]lg.Node{&lg.Or{Terms: negated}}, c.Defs, c.Annot)
}

// --- Check functions ---

// CheckProperties checks all non-temporal properties in the module.
// After checking, properties become axioms.
// This is a stub that corresponds to Python's check_properties.
func CheckProperties(mod *module.Module) error {
	// Stub: in Python this calls itp.false_properties() and
	// optionally launches GUI diagnostics.
	// For now, move props to axioms.
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.LabeledProps...)
	return nil
}

// CheckConjectures checks conjectures in the given state.
// This is a stub corresponding to Python's check_conjectures.
func CheckConjectures(kind, msg string) error {
	// Stub: requires interp.State and analysis graph
	return nil
}

// CheckTemporals checks temporal properties using proof tactics.
// This is a stub corresponding to Python's check_temporals.
func CheckTemporals(mod *module.Module) error {
	// Stub: requires temporal model, proof checker, l2s
	return nil
}

// GetConjs returns the conjecture clauses for the pre-state of inductive checks.
// Only implicit (non-explicit), non-unprovable conjectures and assumed invariants
// are included.
func GetConjs(mod *module.Module) *clauseops.Clauses {
	var fmlas []lg.Node
	all := append(mod.LabeledConjs, mod.AssumedInvs...)
	for _, lf := range all {
		if !lf.Explicit && !lf.Unprovable {
			fmlas = append(fmlas, lf.Formula)
		}
	}
	return clauseops.NewClauses(fmlas, nil, actions.EmptyAnnotation{})
}

// ApplyConjProofs applies proof tactics to conjectures to produce
// conj_subgoals. This is a stub corresponding to Python's apply_conj_proofs.
func ApplyConjProofs(mod *module.Module) {
	// Stub: requires ProofChecker integration
	mod.ConjSubgoals = mod.LabeledConjs
}

// CheckFcsInState checks formula checkers against a state.
// Returns true if all checks pass. This is a stub.
func CheckFcsInState(mod *module.Module, checkers []Checker) bool {
	// Stub: requires analysis graph, history, solver
	for _, fc := range checkers {
		fc.Start()
		// In the real implementation, this would:
		// 1. Get the history from the analysis graph
		// 2. Ask the solver to check satisfiability
		// 3. Call fc.Sat() or fc.Unsat() accordingly
		fc.Pass()
	}
	for _, fc := range checkers {
		if fc.Failed() {
			return false
		}
	}
	return true
}

// CheckConjsInState checks conjectures in a state.
// This is a stub corresponding to Python's check_conjs_in_state.
func CheckConjsInState(mod *module.Module, indent int, pcs []*module.LabeledFormula) bool {
	conjs := mod.ConjSubgoals
	if conjs == nil {
		conjs = mod.LabeledConjs
	}
	var checkers []Checker
	for _, c := range conjs {
		checkers = append(checkers, NewConjChecker(c, indent))
	}
	return CheckFcsInState(mod, checkers)
}

// CheckSafetyInState checks safety (no assertion violations) in a state.
// This is a stub corresponding to Python's check_safety_in_state.
func CheckSafetyInState(mod *module.Module, reportPass bool) bool {
	// Check with an empty Or (false) as the conjecture
	checker := NewBaseChecker(&lg.Or{}, reportPass, true)
	return CheckFcsInState(mod, []Checker{checker})
}

// GetCheckedActions returns the list of actions to be checked.
// If a specific action is set via the "action" parameter, only that
// action is returned. Otherwise all public actions are returned sorted.
func GetCheckedActions(mod *module.Module) []string {
	cact := CheckedAction.GetString()
	if cact != "" {
		extName := "ext:" + cact
		if mod.PublicActions[extName] {
			cact = extName
		}
	}
	if cact != "" && !mod.PublicActions[cact] {
		return nil
	}
	CheckedActionFound = true
	if cact != "" {
		return []string{cact}
	}
	result := make([]string, 0, len(mod.PublicActions))
	for name := range mod.PublicActions {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

// GetPrioritizedActions returns the list of prioritized actions parsed
// from the "prioritize" parameter. Each name is prefixed with "ext:".
func GetPrioritizedActions() []string {
	pas := PriorityActions.Get()
	if pas == nil {
		return nil
	}
	s, ok := pas.(string)
	if !ok || s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = "ext:" + strings.TrimSpace(p)
	}
	sort.Strings(result)
	return result
}

// ConvertPostconds converts postconditions by renaming old symbols.
// This is a stub corresponding to Python's convert_postconds.
func ConvertPostconds(postconds []*module.LabeledFormula) []*module.LabeledFormula {
	// Stub: requires transrel.OldOf, transrel.IsOld, rename_ast
	return postconds
}
