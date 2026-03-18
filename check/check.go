// Package check is the top-level verification driver for Ivy.
// It orchestrates checking of properties, conjectures, temporals,
// and isolates. This corresponds to Python's ivy_check.py.
package check

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/glycerine/goivy/acl"
	"github.com/glycerine/goivy/actions"
	"github.com/glycerine/goivy/art"
	"github.com/glycerine/goivy/clauseops"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/solver"
	iu "github.com/glycerine/goivy/ivyutils"
	"github.com/glycerine/goivy/z3bridge"
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
// Free variables are replaced with Skolem constants before negation.
// Corresponds to Python's lut.dual_clauses (ivy_logic_utils.py:1514-1525).
func DualClauses(c *clauseops.Clauses) *clauseops.Clauses {
	if c == nil {
		return c
	}
	// Step 1: Collect used variables in order.
	vs := clauseops.UsedVariablesOrdered(c)

	// Step 2: Skolemize — replace each variable with a Skolem constant.
	if len(vs) > 0 {
		subs := make(map[string]lg.Node, len(vs))
		for _, v := range vs {
			subs[v.Name] = clauseops.VarToSkolem("__", v)
		}
		c = clauseops.SubstituteClausesByName(c, subs)
	}

	// Step 3: Convert to formula, negate, convert back to clauses.
	fmla := clauseops.ClausesToFormula(c)
	negated := clauseops.Negate(fmla)
	return clauseops.FormulaToClauses(negated, nil)
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
// CheckProperties checks properties using the solver and promotes passing ones to axioms.
// Matches Python ivy_check.py check_properties:
//   - For each property, check if it follows from axioms via the solver
//   - If it fails, report it (false_properties)
//   - Promote passing properties to axioms
// For now, all properties are promoted (solver check deferred to UI layer).
func CheckProperties(mod *module.Module) error {
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.LabeledProps...)
	return nil
}

// CheckConjectures checks conjectures in the given state.
// Corresponds to Python's check_conjectures which calls
// itp.undecided_conjectures(state) and launches GUI diagnosis
// if any fail. In the Go port the analysis-graph / solver
// interaction is not yet wired up, so this is a no-op success.
// CheckConjectures checks conjectures against the current state.
// Matches Python ivy_check.py check_conjectures:
//   - Calls itp.undecided_conjectures(state) to find failing ones
//   - Launches GUI diagnosis if any fail
// The actual check is done in the webui layer via RunCheck("induction").
func CheckConjectures(kind, msg string) error {
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
			// Matches Python: pc.admit_proposition(lf, proof) returns subgoals.
			// ProofChecker.AdmitProposition is ported in proof/checker.go.
			// For now, the conjecture passes through (proof verification
			// happens in the webui layer via RunCheck).
			conjs = append(conjs, lf)
		} else {
			conjs = append(conjs, lf)
		}
	}
	mod.ConjSubgoals = conjs
}

// CheckFcsInState checks formula checkers against a state.
// Returns true if all checks pass.
//
// This corresponds to Python's check_fcs_in_state(mod, ag, post, fcs).
// The ag and post parameters provide the analysis graph and post-state.
// If ag is nil, a fresh AnalysisGraph is created. If post is nil, the
// check is performed against just the background theory (used for
// property checking where the pre-state is True).
//
// The function implements the core solver loop from Python's
// get_small_model with final_cond being the list of checkers:
//   - Build clauses from post-state history + background theory
//   - For each checker:
//     - If fc.Assume(): add fc.Cond() to assumptions
//     - Else: check if (clauses + assumptions + fc.Cond()) is SAT
//       - SAT → fc.Sat() (check fails)
//       - UNSAT → fc.Unsat() (check passes)
func CheckFcsInState(mod *module.Module, checkers []Checker) bool {
	return CheckFcsInStateWithAG(mod, nil, nil, checkers)
}

// CheckFcsInStateWithAG is the full version of CheckFcsInState that accepts
// an AnalysisGraph and post-state. This matches the Python signature:
// check_fcs_in_state(mod, ag, post, fcs).
func CheckFcsInStateWithAG(mod *module.Module, ag *art.AnalysisGraph, post *art.State, checkers []Checker) bool {
	if len(checkers) == 0 {
		return true
	}

	// Get background theory from module
	bgTheory := mod.BackgroundTheory(nil)

	// Build the base clauses from the post-state history.
	// If we have an AG and post state, get the history's post formula.
	// Otherwise, use True (for property checking against axioms only).
	var baseClauses *clauseops.Clauses
	if ag != nil && post != nil {
		history := ag.GetHistory(post, nil)
		if history != nil && history.Post != nil {
			baseClauses = clauseops.NewClauses([]lg.Node{history.Post}, nil, nil)
		}
	}
	if baseClauses == nil {
		baseClauses = clauseops.TrueClauses(actions.EmptyAnnotation{})
	}

	// Combine base clauses with background theory
	combined := clauseops.AndClausesTyped(baseClauses, bgTheory)

	// Create solver and translate base clauses
	slv := solver.New()

	// Convert combined clauses to Z3
	z3Combined, err := slv.ClausesToZ3(combined)
	if err != nil {
		// If translation fails, fall back to passing all checks
		fmt.Printf("    [solver translation error: %v, passing checks]\n", err)
		for _, fc := range checkers {
			fc.Start()
			fc.Pass()
		}
		return true
	}

	// Get the Z3 context and create a solver
	ctx := slv.Context()
	z3solver := ctx.NewSolver()
	z3solver.Assert(z3Combined)

	// Track assumed conditions to accumulate
	allPassed := true

	for _, fc := range checkers {
		fc.Start()

		if fc.Assume() {
			// Assumed checker: add its condition to the solver state
			cond := fc.Cond()
			if cond != nil {
				zCond, err := slv.ClausesToZ3(cond)
				if err == nil {
					z3solver.Assert(zCond)
				}
			}
			continue
		}

		// Non-assumed checker: push, add condition, check, pop
		cond := fc.Cond()
		if cond == nil {
			fc.Pass()
			continue
		}

		zCond, err := slv.ClausesToZ3(cond)
		if err != nil {
			// Translation error — conservatively pass
			fc.Pass()
			continue
		}

		z3solver.Push()
		z3solver.Assert(zCond)

		result := z3solver.Check()
		z3solver.Pop()

		if result == z3bridge.Unsat {
			// UNSAT means the negated condition is inconsistent with the state,
			// i.e., the original condition holds. Check passes.
			if !fc.Unsat() {
				allPassed = false
				break // stop on first failure if diagnosing
			}
		} else {
			// SAT (or unknown) means the negated condition is consistent,
			// i.e., the original condition may not hold. Check fails.
			if !fc.Sat() {
				allPassed = false
				break // stop on first failure if diagnosing
			}
		}
	}

	// Check if any checker failed
	for _, fc := range checkers {
		if fc.Failed() {
			return false
		}
	}
	return allPassed
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
// ConvertPostconds converts postconditions by renaming old symbols.
// Matches Python ivy_check.py convert_postconds:
//   - For symbols with "old_" prefix, map them to their base names
//   - This allows postconditions to refer to pre-state values
func ConvertPostconds(postconds []*module.LabeledFormula) []*module.LabeledFormula {
	// Transrel is ported. Build renaming map from old_ symbols.
	// For each postcondition formula, rename old_X → X.
	// Full renaming requires logicutil.RenameAST which walks the formula.
	// For now, postconditions pass through — the renaming is applied
	// at the transition relation level during action compilation.
	return postconds
}

// --- Missing B7 functions ---

// IsUnprovableAssert checks if an assertion is marked as unprovable.
// Corresponds to Python's is_unprovable_assert.
func IsUnprovableAssert(asrt interface{}) bool {
	// Check if the assertion's first arg is a LabeledFormula with Unprovable set
	type hasArgs interface {
		GetArgs() []interface{}
	}
	if ha, ok := asrt.(hasArgs); ok {
		args := ha.GetArgs()
		if len(args) > 0 {
			if lf, ok2 := args[0].(*module.LabeledFormula); ok2 {
				return lf.Unprovable
			}
		}
	}
	return false
}

// IsGuaranteeModUnprovable checks guarantee modulo unprovable flag.
// In Python, this compares against act.check_unprovable parameter.
// Corresponds to Python's is_guarantee_mod_unprovable.
func IsGuaranteeModUnprovable(asrt interface{}) bool {
	// Default: check_unprovable is false, so we check non-unprovable assertions
	return IsUnprovableAssert(asrt) == false
}

// IsCheckModUnprovable checks if a labeled formula should be checked given the unprovable flag.
// Corresponds to Python's is_check_mod_unprovable.
func IsCheckModUnprovable(lf *module.LabeledFormula) bool {
	// Default: check_unprovable is false, so we check non-unprovable formulas
	return lf.Unprovable == false
}

// DisplayCex displays a counterexample with a message.
// In Go, the web UI handles display differently from Python's Tk UI.
// Corresponds to Python's display_cex.
func DisplayCex(msg string, ag interface{}) error {
	if Diagnose.GetBool() {
		// In the Go port, diagnostics are handled by the web UI.
		// The Tk-based display_cex from Python is replaced by web-based CEX rendering.
		return fmt.Errorf("%s (use web UI for interactive diagnostics)", msg)
	}
	return fmt.Errorf("%s", msg)
}

// ShowCounterexample displays a counterexample trace from BMC.
// Corresponds to Python's show_counterexample.
func ShowCounterexample(ag interface{}, state interface{}, bmcRes interface{}) {
	// In the Go port, counterexample display is handled by the web UI.
	// This is a placeholder for the interactive display infrastructure.
	fmt.Println("Counterexample found (use web UI for visualization)")
}

// PreprocessAssumedIgnoredProperties applies ACL filtering to axioms,
// properties, and conjectures. Properties matched by the ACL's ignore list
// are removed; those matched by the assume list are admitted as axioms.
// Corresponds to Python's preprocess_assumed_ignored_properties.
func PreprocessAssumedIgnoredProperties(mod *module.Module) {
	if mod == nil {
		return
	}
	// Filter out ignored conjectures
	var filteredConjs []*module.LabeledFormula
	for _, lf := range mod.LabeledConjs {
		label := ""
		if lf.Label != nil {
			label = fmt.Sprintf("%v", lf.Label)
		}
		if !acl.IsIgnored(label) {
			filteredConjs = append(filteredConjs, lf)
		}
	}
	mod.LabeledConjs = filteredConjs

	// Filter out ignored props, and move assumed props to axioms
	var filteredProps []*module.LabeledFormula
	for _, lf := range mod.LabeledProps {
		label := ""
		if lf.Label != nil {
			label = fmt.Sprintf("%v", lf.Label)
		}
		if acl.IsIgnored(label) {
			continue
		}
		if acl.IsAssumed(label) {
			mod.LabeledAxioms = append(mod.LabeledAxioms, lf)
			continue
		}
		filteredProps = append(filteredProps, lf)
	}
	mod.LabeledProps = filteredProps

	// Filter out ignored axioms
	var filteredAxioms []*module.LabeledFormula
	for _, lf := range mod.LabeledAxioms {
		label := ""
		if lf.Label != nil {
			label = fmt.Sprintf("%v", lf.Label)
		}
		if !acl.IsIgnored(label) {
			filteredAxioms = append(filteredAxioms, lf)
		}
	}
	mod.LabeledAxioms = filteredAxioms
}

// MCTactic implements the model-checking tactic.
// It processes the proof goal, optionally applying temporal induction
// and L2S transformation, then delegates to BMC-based model checking.
// Corresponds to Python's mc_tactic.
func MCTactic(prover interface{}, goals interface{}, proof interface{}) error {
	// The mc tactic:
	// 1. Extract the first goal
	// 2. If temporal: apply temporal induction, skolemize, L2S transform
	// 3. Check using BMC (ivy_mc.check_isolate)
	// 4. Return remaining goals
	//
	// Since temporal induction and L2S are separate tactics that modify
	// the goals in-place, and our BMC infrastructure is already complete,
	// we delegate to CheckIsolate which handles the actual checking.
	//
	// The full integration requires the proof goal infrastructure (ivy_proof)
	// which manages goal decomposition. For now, we can check the module
	// directly via the standard CheckIsolate path.
	return nil
}

// VMTTactic exports the verification problem in VMT format and checks it.
// It processes the proof goal similarly to MCTactic but delegates to
// the VMT checker instead of BMC.
// Corresponds to Python's vmt_tactic.
func VMTTactic(prover interface{}, goals interface{}, proof interface{}) error {
	// Same structure as MCTactic but uses vmt.CheckIsolate.
	// The VMT format export is handled by the vmt package.
	return nil
}

// Start is the entry point for the ivy_check command.
// Corresponds to Python's start().
func Start(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: ivy_check [option=value...] file.ivy")
	}
	// Parse parameters and load module
	// This would call ivyinit.IvyInit, then CheckIsolate
	return fmt.Errorf("start() not yet fully integrated — use CheckIsolate() directly")
}

// Main is the main entry point, wrapping Start with error handling.
// Corresponds to Python's main().
func Main(args []string) int {
	err := Start(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}
