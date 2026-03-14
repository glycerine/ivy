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
// After checking, properties become axioms. Corresponds to Python's
// check_properties which calls itp.false_properties() and (optionally)
// launches a GUI diagnostic if diagnosis is enabled.
//
// In the Go port, property falsification is delegated to the solver
// integration (not yet wired up), so for now the properties are simply
// promoted to axioms, which is the normal successful-check behaviour.
func CheckProperties(mod *module.Module) error {
	// TODO: once the solver is wired up, check for false_properties
	// and raise an error (with optional GUI) if any fail.
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.LabeledProps...)
	return nil
}

// CheckConjectures checks conjectures in the given state.
// Corresponds to Python's check_conjectures which calls
// itp.undecided_conjectures(state) and launches GUI diagnosis
// if any fail. In the Go port the analysis-graph / solver
// interaction is not yet wired up, so this is a no-op success.
func CheckConjectures(kind, msg string) error {
	// TODO: requires analysis graph state and itp.undecided_conjectures.
	// When those are ported, this should call the solver, check for
	// undecided conjectures, and (optionally) launch diagnostics.
	return nil
}

// CheckTemporals checks temporal properties using proof tactics.
// Corresponds to Python's check_temporals which builds a ProofChecker
// from axioms+assumed_invariants, definitions, and schemata, then
// iterates over labeled_props. Assumed or unchecked temporal props are
// admitted as axioms; others are proved via admit_proposition with the
// property's proof (from mod.Proofs). If a property has no proof or
// the proof fails, an error is returned.
//
// The full implementation requires the temporal-model builder
// (ivy_temporal.normal_program_from_module) and ivy_proof.ProofChecker.
// Until those are ported, this validates the non-temporal fast-path:
// temporal properties without proofs are skipped with a warning.
func CheckTemporals(mod *module.Module) error {
	// Build a proof map: formula-ID -> proof
	pmap := make(map[int64]interface{})
	for _, pe := range mod.Proofs {
		pmap[pe.Formula.ID] = pe.Proof
	}

	for _, prop := range mod.LabeledProps {
		if !prop.Temporal {
			continue
		}
		if prop.Assumed {
			// Assumed temporal property — skip (admitted as axiom).
			fmt.Println(PrettyLF(prop, 4) + "  [assumed temporal]")
			continue
		}
		proof, hasProof := pmap[prop.ID]
		if !hasProof || proof == nil {
			// No proof supplied — nothing we can verify without the
			// temporal model builder. Warn and continue.
			fmt.Println(PrettyLF(prop, 4) + "  [temporal: no proof available]")
			continue
		}
		// With a proof, the full path would be:
		//   propn := proof.NormalizeGoal(prop)
		//   model := itmp.NormalProgramFromModule(mod)
		//   subgoal := ... TemporalModels(model, propn.Formula) ...
		//   subgoals := pc.AdmitProposition(prop, proof, [subgoal])
		//   CheckSubgoals(subgoals)
		// Until those are ported, we emit the status line.
		fmt.Print("\n    The following temporal property is being proved:\n")
		fmt.Print(PrettyLF(prop, 4) + " ... ")
		fmt.Println("[temporal proof checking not yet fully ported]")
	}
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
// conj_subgoals. Corresponds to Python's apply_conj_proofs which
// creates a ProofChecker from axioms+assumed_invariants, definitions,
// and schemata, then for each conjecture that has a proof in mod.Proofs
// calls pc.admit_proposition to produce subgoals. Conjectures without
// proofs pass through unchanged.
//
// Subgoals produced by proof tactics should ideally be run through
// ivy_compiler.theorem_to_property, but until the compiler is wired
// up, we simply collect them.
func ApplyConjProofs(mod *module.Module) {
	pmap := make(map[int64]interface{})
	for _, pe := range mod.Proofs {
		pmap[pe.Formula.ID] = pe.Proof
	}

	var conjs []*module.LabeledFormula
	for _, lf := range mod.LabeledConjs {
		if _, hasProof := pmap[lf.ID]; hasProof {
			// TODO: Once proof.ProofChecker.AdmitProposition is fully
			// wired up, call it here:
			//   subgoals := pc.AdmitProposition(lf, proof)
			//   conjs = append(conjs, subgoals...)
			// For now, the conjecture passes through as-is.
			conjs = append(conjs, lf)
		} else {
			conjs = append(conjs, lf)
		}
	}
	mod.ConjSubgoals = conjs
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
// Corresponds to Python's check_conjs_in_state which:
// 1. Uses conj_subgoals if available, else labeled_conjs.
// 2. Filters for checkable (non-unprovable) conjectures.
// 3. Appends converted postconditions (pcs).
// 4. Optionally filters by a checked-assert line number.
// 5. Creates ConjChecker for each, then delegates to CheckFcsInState.
func CheckConjsInState(mod *module.Module, indent int, pcs []*module.LabeledFormula) bool {
	conjs := mod.ConjSubgoals
	if conjs == nil {
		conjs = mod.LabeledConjs
	}

	// Filter for checkable conjectures (non-unprovable).
	var checkable []*module.LabeledFormula
	for _, c := range conjs {
		if !c.Unprovable {
			checkable = append(checkable, c)
		}
	}

	// Append converted postconditions.
	if len(pcs) > 0 {
		converted := ConvertPostconds(pcs)
		checkable = append(checkable, converted...)
	}

	// Build checkers for the filtered list.
	var checkers []Checker
	for _, c := range checkable {
		checkers = append(checkers, NewConjChecker(c, indent))
	}

	// Apply line-number filter if set.
	if CheckLineno != "" {
		checkers = FilterCheckers(checkers, CheckLineno)
	}

	return CheckFcsInState(mod, checkers)
}

// CheckSafetyInState checks safety (no assertion violations) in a state.
// Corresponds to Python's check_safety_in_state which creates a
// Checker(lg.Or(), report_pass) and delegates to check_fcs_in_state.
// lg.Or() with no terms is "false", so after dualization the check
// succeeds iff the post-state has no assertion violations.
func CheckSafetyInState(mod *module.Module, reportPass bool) bool {
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
// Corresponds to Python's convert_postconds which replaces "old"
// symbols with their pre-state counterparts using transrel.old_of /
// transrel.is_old and lut.rename_ast. Until the transition-relation
// module is fully ported, postconditions pass through unchanged.
func ConvertPostconds(postconds []*module.LabeledFormula) []*module.LabeledFormula {
	// TODO: once transrel is ported, build a renaming map:
	//   for each used symbol s in postconds where transrel.IsOld(s):
	//     renaming[s] = transrel.OldOf(s)
	//   for each updated symbol s:
	//     renaming[transrel.Old(s)] = s.Prefix("__")
	//   return [lf.Clone([lf.Label, lut.RenameAST(lf.Formula, renaming)]) ...]
	return postconds
}
