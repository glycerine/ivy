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
	"github.com/glycerine/goivy/ast"
	"github.com/glycerine/goivy/clauseops"
	"github.com/glycerine/goivy/compiler"
	"github.com/glycerine/goivy/interp"
	"github.com/glycerine/goivy/ivyinit"
	"github.com/glycerine/goivy/l2s"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/mc"
	"github.com/glycerine/goivy/module"
	"github.com/glycerine/goivy/proof"
	"github.com/glycerine/goivy/solver"
	"github.com/glycerine/goivy/tactics"
	"github.com/glycerine/goivy/temporal"
	tr "github.com/glycerine/goivy/transrel"
	"github.com/glycerine/goivy/vmt"
)

func init() {
	// Wire AdmitDefinitionFactory so that compiler.CheckDefinitions can call
	// proof.ProofChecker.AdmitDefinition without a direct import cycle.
	// Python: prover.admit_definition(d, pmap[d.id])
	compiler.AdmitDefinitionFactory = func(mod *module.Module) func(defn *ast.LabeledFormula, pf ast.Node) error {
		return func(defn *ast.LabeledFormula, pf ast.Node) error {
			typedSchemata := ModuleSchemataToAst(mod.Schemata)
			prover := proof.NewProofChecker(nil, mod.LabeledAxioms, nil, typedSchemata)
			_, err := prover.AdmitDefinition(defn, pf)
			return err
		}
	}
}

// --- Package-level parameters ---

/*
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
	// CheckUnprovable corresponds to Python's act.check_unprovable
	// (ivy_actions.py:25). When true, only unprovable assertions are checked.
	CheckUnprovable      = iu.NewBooleanParameter("unprovable", false)
)


// Failures tracks the number of failed checks during verification.
var Failures int

// CheckedActionFound tracks whether a checked action was found.
var CheckedActionFound bool

// CheckLineno is the current line number being checked, or empty for all.
var CheckLineno string
*/

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
	GetLF() *ast.LabeledFormula
}

// --- BaseChecker ---

// BaseChecker is the default checker implementation.
// It wraps a formula conjecture and checks it against a state by
// negating it (dualizing) and checking satisfiability.
type BaseChecker struct {
	Cfg        *module.Config
	FC         *clauseops.Clauses
	ReportPass bool
	Inverted   bool
	FailedFlag bool
}

// NewBaseChecker creates a BaseChecker for the given conjecture formula.
// If invert is true (the default), the formula is dualized for checking.
func NewBaseChecker(cfg *module.Config, conj lg.Expr, reportPass bool, invert bool) *BaseChecker {
	fc := clauseops.FormulaToClauses(conj, nil)
	if invert {
		fc = DualClauses(fc)
	}
	return &BaseChecker{
		Cfg:        cfg,
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
func (c *BaseChecker) Sat() bool {
	// Python: return self._pass() if act.check_unprovable.get() else self.fail()
	if c.Cfg.OnlyCheckUnprovable {
		return c.Pass()
	}
	return c.Fail()
}
func (c *BaseChecker) Unsat() bool {
	// Python: return self.fail() if act.check_unprovable.get() else self._pass()
	if c.Cfg.OnlyCheckUnprovable {
		return c.Fail()
	}
	return c.Pass()
}
func (c *BaseChecker) Assume() bool               { return false }
func (c *BaseChecker) GetAnnot() interface{}      { return nil }
func (c *BaseChecker) Failed() bool               { return c.FailedFlag }
func (c *BaseChecker) GetLF() *ast.LabeledFormula { return nil }

func (c *BaseChecker) Fail() bool {
	fmt.Println("FAIL")
	c.Cfg.Failures++
	c.FailedFlag = true
	// Python: return not (diagnose.get() or opt_trace.get()) or act.check_unprovable.get()
	return !(c.Cfg.Diagnose || c.Cfg.OptTrace) || c.Cfg.OnlyCheckUnprovable
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
	LF     *ast.LabeledFormula
	Indent int
}

// NewConjChecker creates a ConjChecker for the given labeled formula.
func NewConjChecker(cfg *module.Config, lf *ast.LabeledFormula, indent int) *ConjChecker {
	base := NewBaseChecker(cfg, lf.Formula.(lg.Expr), true, true)
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
	// Python: return self.lf.annot if hasattr(self.lf,'annot') else None
	if c.LF != nil && c.LF.Annot != nil {
		return c.LF.Annot
	}
	return nil
}

func (c *ConjChecker) GetLF() *ast.LabeledFormula { return c.LF }

// --- ConjAssumer ---

// ConjAssumer treats a conjecture as assumed (not checked).
type ConjAssumer struct {
	BaseChecker
	LF *ast.LabeledFormula
}

// NewConjAssumer creates a ConjAssumer for the given labeled formula.
func NewConjAssumer(cfg *module.Config, lf *ast.LabeledFormula) *ConjAssumer {
	base := NewBaseChecker(cfg, lf.Formula.(lg.Expr), false, false)
	return &ConjAssumer{
		BaseChecker: *base,
		LF:          lf,
	}
}

func (c *ConjAssumer) Start() {
	fmt.Println(PrettyLF(c.LF, 8) + "  [assumed]")
}

func (c *ConjAssumer) Assume() bool               { return true }
func (c *ConjAssumer) GetLF() *ast.LabeledFormula { return c.LF }

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
		subs := make(map[string]lg.Expr, len(vs))
		for _, v := range vs {
			subs[v.Name] = clauseops.VarToSkolem("@", v)
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
// Matches Python ivy_check.py check_properties (lines 61-74):
//   - Calls itp.false_properties() to find properties not implied by axioms
//   - If any fail, reports error (optionally launches diagnosis)
//   - Promotes all properties to axioms
//   - Calls mod.UpdateTheory() to rebuild background theory
func CheckProperties(mod *module.Module) error {
	failed := interp.FalseProperties(mod)
	if len(failed) > 0 {
		if mod.Cfg.Diagnose {
			fmt.Println("Some properties failed.")
		}
		return fmt.Errorf("some properties failed")
	}
	mod.LabeledAxioms = append(mod.LabeledAxioms, mod.LabeledProps...)
	mod.UpdateTheory()
	return nil
}

// CheckConjectures checks conjectures in the given state.
// Corresponds to Python's check_conjectures which calls
// itp.undecided_conjectures(state) and launches GUI diagnosis
// if any fail. In the Go port the analysis-graph / solver
// interaction is not yet wired up, so this is a no-op success.
// CheckConjectures checks conjectures against the current state.
// Matches Python ivy_check.py check_conjectures (lines 104-117):
//   - Calls itp.undecided_conjectures(state) to find failing ones
//   - Reports error if any fail
func CheckConjectures(cfg *module.Config, kind, msg string, ag *art.AnalysisGraph, state *interp.State) error {
	failed := interp.UndecidedConjectures(state)
	if len(failed) > 0 {
		if cfg.Diagnose {
			fmt.Printf("%s failed.\n", kind)
		}
		return fmt.Errorf("%s failed", kind)
	}
	return nil
}

// CheckTemporals checks temporal properties using proof tactics.
// Corresponds to Python's check_temporals (ivy_check.py:127-157) which builds
// a ProofChecker from axioms+assumed_invariants, definitions, and schemata,
// then iterates over labeled_props. Assumed or unchecked temporal props are
// admitted as axioms; others are proved via admit_proposition with the
// property's proof (from mod.Proofs).
func CheckTemporals(mod *module.Module) error {
	// Python: pmap = dict((prop.id,p) for prop,p in mod.proofs)
	pmap := make(map[int64]interface{})
	for _, pe := range mod.Proofs {
		pmap[pe.Formula.ID] = pe.Proof
	}

	// Python: pc = ivy_proof.ProofChecker(mod.labeled_axioms+mod.assumed_invariants,
	//                                     mod.definitions, mod.schemata)
	pcAxioms := make([]*ast.LabeledFormula, 0, len(mod.LabeledAxioms)+len(mod.AssumedInvs))
	pcAxioms = append(pcAxioms, mod.LabeledAxioms...)
	pcAxioms = append(pcAxioms, mod.AssumedInvs...)
	pc := proof.NewProofChecker(nil, pcAxioms, mod.Definitions, ModuleSchemataToAst(mod.Schemata))

	// Build ACL config if unchecked properties file is specified
	var aclCfg *acl.Config
	if mod.Cfg.OptUncheckedProps != "" {
		aclCfg = acl.NewConfig()
	}

	for _, prop := range mod.LabeledProps {
		if !prop.IsTemporal() {
			continue
		}

		// Python: if prop.assumed or opt_unchecked_properties.get() and ivy_acl.is_assumed(prop.label):
		propLabel := fmt.Sprint(prop.Label)
		isAssumedByACL := aclCfg != nil && aclCfg.IsAssumed(propLabel)
		if prop.Assumed || isAssumedByACL {
			fmt.Println("  ivy_check temporal: admitting axiom...", PrettyLF(prop, 0))
			if isAssumedByACL {
				fmt.Printf("     ... admitting %s as axiom because it is an externally assumed property and unchecked property file is supplied.\n", propLabel)
			}
			pc.AdmitAxiom(prop)
		} else {
			fmt.Print("\n    The following temporal property is being proved:\n")
			fmt.Print(PrettyLF(prop, 4) + " ... ")

			// Python: proof = pmap.get(prop.id, None)
			pf := pmap[prop.ID]

			// Python: propn = ivy_proof.normalize_goal(prop)
			propn := proof.NormalizeGoal(prop)

			// Python: model = itmp.normal_program_from_module(im.module)
			model := temporal.NormalProgramFromModule(mod)

			// Python: subgoal = prop.clone([prop.args[0], ivy_ast.TemporalModels(model, propn.args[1])])
			tm := &ast.TemporalModels{Model: model, Fmla: propn.Formula}
			subgoal := prop.Clone([]ast.Node{prop.Label, tm}).(*ast.LabeledFormula)

			subgoals := []*ast.LabeledFormula{subgoal}

			// Python: subgoals = pc.admit_proposition(prop, proof, subgoals)
			var pfNode ast.Node
			if pf != nil {
				pfNode, _ = pf.(ast.Node)
			}
			var err error
			subgoals, err = pc.AdmitProposition(prop, pfNode, subgoals...)
			if err != nil {
				return err
			}

			// Python: check_subgoals(subgoals)
			if err := CheckSubgoals(subgoals, nil, mod); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetConjs returns the conjecture clauses for the pre-state of inductive checks.
// Only implicit (non-explicit), non-unprovable conjectures and assumed invariants
// are included.
func GetConjs(mod *module.Module) *clauseops.Clauses {
	var fmlas []lg.Expr
	all := append(mod.LabeledConjs, mod.AssumedInvs...)
	for _, lf := range all {
		if !lf.Explicit && !lf.Unprovable {
			fmlas = append(fmlas, lf.Formula.(lg.Expr))
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
	// Python: pc = ivy_proof.ProofChecker(mod.labeled_axioms+mod.assumed_invariants, mod.definitions, mod.schemata)
	// The proof package uses ast.LabeledFormula (with ast.Node fields) while
	// module uses ast.LabeledFormula (with lg.Expr fields). These are separate
	// type hierarchies — a porting mistake (Python has one LabeledFormula class).
	// Until the two are unified, we attempt proof application when the formula's
	// concrete type satisfies ast.Node, and fall through otherwise.
	pcAxioms := make([]*ast.LabeledFormula, 0, len(mod.LabeledAxioms)+len(mod.AssumedInvs))
	for _, lf := range mod.LabeledAxioms {
		if alf := ModuleLFToAstLF(lf); alf != nil {
			pcAxioms = append(pcAxioms, alf)
		}
	}
	for _, lf := range mod.AssumedInvs {
		if alf := ModuleLFToAstLF(lf); alf != nil {
			pcAxioms = append(pcAxioms, alf)
		}
	}
	pcDefs := make([]*ast.LabeledFormula, 0, len(mod.Definitions))
	for _, lf := range mod.Definitions {
		if alf := ModuleLFToAstLF(lf); alf != nil {
			pcDefs = append(pcDefs, alf)
		}
	}
	pc := proof.NewProofChecker(nil, pcAxioms, pcDefs, ModuleSchemataToAst(mod.Schemata))

	pmap := make(map[int64]interface{})
	for _, pe := range mod.Proofs {
		pmap[pe.Formula.ID] = pe.Proof
	}

	var conjs []*ast.LabeledFormula
	for _, lf := range mod.LabeledConjs {
		if p, hasProof := pmap[lf.ID]; hasProof {
			// Python: subgoals = pc.admit_proposition(lf, proof)
			astLF := ModuleLFToAstLF(lf)
			astProof, _ := p.(ast.Node)
			if astLF != nil && astLF.Formula != nil {
				// Python: subgoals = pc.admit_proposition(lf, proof)
				subgoals, err := pc.AdmitProposition(astLF, astProof)
				if err == nil && len(subgoals) > 0 {
					// Python: subgoals = list(map(ivy_compiler.theorem_to_property, subgoals))
					for _, sg := range subgoals {
						modSG := AstLFToModuleLF(sg)
						modSG = compiler.TheoremToProperty(modSG, mod)
						conjs = append(conjs, modSG)
					}
					continue
				}
			}
			// If conversion or proof application fails, pass through unchanged
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
//   - If fc.Assume(): add fc.Cond() to assumptions
//   - Else: check if (clauses + assumptions + fc.Cond()) is SAT
//   - SAT → fc.Sat() (check fails)
//   - UNSAT → fc.Unsat() (check passes)
func CheckFcsInState(mod *module.Module, checkers []Checker) bool {
	return CheckFcsInStateWithAG(mod, nil, nil, checkers)
}

// CheckFcsInStateWithAG is the full version of CheckFcsInState that accepts
// an AnalysisGraph and post-state. This matches the Python signature:
// check_fcs_in_state(mod, ag, post, fcs) (lines 373-416).
//
// Two paths:
//  1. trace/diagnose: Build model via SmallModelClauses, create Trace,
//     call MatchAnnotation, display trace
//  2. normal: Call history.SatisfyWithCond(axioms, gmc, fcs)
func CheckFcsInStateWithAG(mod *module.Module, ag *art.AnalysisGraph, post *art.State, checkers []Checker) bool {
	if len(checkers) == 0 {
		return true
	}

	// Get history and background theory
	var history *tr.History
	if ag != nil && post != nil {
		history = ag.GetHistory(post, nil)
	}
	axioms := mod.BackgroundTheory(nil)

	if mod.Cfg.OptTrace || mod.Cfg.Diagnose {
		// Trace/diagnose path (Python lines 379-411)
		return checkFcsTracePath(mod, ag, post, history, axioms, checkers)
	}

	// Normal path (Python lines 412-415)
	return checkFcsNormalPath(mod, ag, post, history, axioms, checkers)
}

// checkFcsTracePath implements the trace/diagnose branch of check_fcs_in_state.
// Python lines 379-411.
func checkFcsTracePath(mod *module.Module, ag *art.AnalysisGraph, post *art.State,
	history *tr.History, axioms *clauseops.Clauses, checkers []Checker) bool {

	if history == nil || history.Post == nil {
		// No history — fall back to normal path
		return checkFcsNormalPath(mod, ag, post, history, axioms, checkers)
	}

	// Python: clauses = history.post; clauses = lut.and_clauses(clauses, axioms)
	postClauses := clauseops.NewClauses([]lg.Expr{history.Post}, nil, nil)
	clauses := clauseops.AndClausesTyped(postClauses, axioms)

	// Python: ffcs = filter_fcs(fcs)
	ffcs := FilterCheckers(checkers, mod.Cfg.CheckLineno)

	// Python: model = itr.small_model_clauses(clauses, ffcs, shrink=True)
	var finalConds []solver.FinalCond
	for _, fc := range ffcs {
		finalConds = append(finalConds, fc)
	}
	model, modelSlv := tr.SmallModelClauses(clauses, finalConds, true, mod)

	if model != nil {
		// Python: failed = [c for c in ffcs if c.failed]
		var failed []Checker
		for _, c := range ffcs {
			if c.Failed() {
				failed = append(failed, c)
			}
		}
		if len(failed) == 0 {
			// No failures despite SAT model — all passed
			return true
		}

		// Python: mclauses = lut.and_clauses(*([clauses] + [c.cond() for c in failed]))
		mclauses := clauses
		for _, c := range failed {
			if c.Cond() != nil {
				mclauses = clauseops.AndClausesTyped(mclauses, c.Cond())
			}
		}

		// Python: vocab = lut.used_symbols_clauses(mclauses)
		vocabMap := mclauses.Symbols()
		vocab := make([]*lg.Symbol, 0, len(vocabMap))
		for _, expr := range vocabMap {
			if sym, ok := expr.(*lg.Symbol); ok {
				vocab = append(vocab, sym)
			}
		}

		// Python: handler = ivy_trace.Trace(mclauses, model, vocab)
		// In Go, MatchHandler implements actions.AnnotationHandler and does
		// the same Eqs extraction as Python's Trace class.
		// Build handler using the model and solver from SmallModelClauses.
		handler := NewMatchHandler(mclauses, model, vocab, modelSlv)

		// Python: thing = failed[-1].get_annot()
		thing := failed[len(failed)-1].GetAnnot()
		if thing == nil {
			// Python: actions = [mod.actions[a] if isinstance(a,str) else a for a in history.actions]
			//         action = act.Sequence(*actions); annot = clauses.annot
			// In Go, history.Actions is []lg.Expr. String action names are *lg.Symbol.
			var actionExprs []lg.Expr
			for _, a := range history.Actions {
				// Python: mod.actions[a] if isinstance(a, str) else a
				if sym, ok := a.(*lg.Symbol); ok {
					if act, exists := mod.Actions[sym.Name]; exists {
						if actAction, ok := act.(actions.Action); ok {
							actionExprs = append(actionExprs, actions.WrapAction(actAction))
							continue
						}
					}
				}
				actionExprs = append(actionExprs, a)
			}
			action := actions.NewSequence(actionExprs...)
			var annot actions.Annotation
			if clauses.Annot != nil {
				annot, _ = clauses.Annot.(actions.Annotation)
			}
			if annot != nil {
				actions.MatchAnnotation(action, annot, handler, mod)
			}
		} else {
			// Python: action, annot = thing
			type annotPair struct {
				Action actions.Action
				Annot  actions.Annotation
			}
			if pair, ok := thing.(*annotPair); ok {
				actions.MatchAnnotation(pair.Action, pair.Annot, handler, mod)
			}
		}
		handler.End()

		// Python: if hasattr(mod,"trace_hook"): handler = mod.trace_hook(handler, ffcs)
		// trace_hook is set by l2s for temporal property diagnostics.

		// Python: ff = failed[0]
		// handler.is_cti = lut.formula_to_clauses(ff.lf.formula) if isinstance(ff, ConjChecker) else None
		ff := failed[0]
		if cc, ok := ff.(*ConjChecker); ok {
			handler.IsCti = clauseops.FormulaToClauses(cc.LF.Formula.(lg.Expr), nil)
		}

		// Python: if not opt_trace.get(): gui_art(handler)
		// else: print(str(handler)); exit(0)
		if mod.Cfg.OptTrace {
			fmt.Println(handler.String())
			os.Exit(0)
		} else {
			// GUI display not supported in Go; print trace instead
			fmt.Println(handler.String())
		}
	}

	return !anyFailed(checkers)
}

// checkFcsNormalPath implements the normal (non-trace) branch of check_fcs_in_state.
// Python (lines 412-415):
//
//	res = history.satisfy(axioms, gmc, filter_fcs(fcs))
//	if res is not None and diagnose.get():
//	    show_counterexample(ag, post, res)
func checkFcsNormalPath(mod *module.Module, ag *art.AnalysisGraph, post *art.State,
	history *tr.History, axioms *clauseops.Clauses, checkers []Checker) bool {

	// Python: filter_fcs(fcs) — filter by check_lineno
	filteredCheckers := FilterCheckers(checkers, mod.Cfg.CheckLineno)

	// Convert checkers to solver.FinalCond for history.SatisfyWithCond
	var finalConds []solver.FinalCond
	for _, fc := range filteredCheckers {
		finalConds = append(finalConds, fc)
	}

	if history != nil {
		// Python: gmc = lambda cls, final_cond: itr.small_model_clauses(cls, final_cond, shrink=diagnose.get())
		gmc := func(cls *clauseops.Clauses, fc []solver.FinalCond) *solver.ModelResult {
			mr, _ := tr.SmallModelClauses(cls, fc, mod.Cfg.Diagnose, mod)
			return mr
		}

		// Python: res = history.satisfy(axioms, gmc, filter_fcs(fcs))
		axiomExpr := clauseops.ClausesToFormula(axioms)
		res := history.SatisfyWithCond(axiomExpr, gmc, finalConds)

		// Python: if res is not None and diagnose.get(): show_counterexample(ag, post, res)
		if res != nil && mod.Cfg.Diagnose {
			ShowCounterexample(ag, post, res)
		}
	} else {
		// No history — fall back to direct solver check.
		// This happens when ag/post are nil (e.g., property checking with true pre-state).
		baseClauses := clauseops.TrueClauses(actions.EmptyAnnotation{})
		combined := clauseops.AndClausesTyped(baseClauses, axioms)

		gmc := func(cls *clauseops.Clauses, fc []solver.FinalCond) *solver.ModelResult {
			mr, _ := tr.SmallModelClauses(cls, fc, mod.Cfg.Diagnose, mod)
			return mr
		}
		gmc(combined, finalConds)
	}

	return !anyFailed(checkers)
}

// anyFailed returns true if any checker has failed.
func anyFailed(checkers []Checker) bool {
	for _, fc := range checkers {
		if fc.Failed() {
			return true
		}
	}
	return false
}

// CheckConjsInState checks conjectures in a state.
// Corresponds to Python's check_conjs_in_state which:
// 1. Uses conj_subgoals if available, else labeled_conjs.
// 2. Filters for checkable (non-unprovable) conjectures.
// 3. Appends converted postconditions (pcs).
// 4. Optionally filters by a checked-assert line number.
// 5. Creates ConjChecker for each, then delegates to CheckFcsInState.
func CheckConjsInState(mod *module.Module, indent int, pcs []*ast.LabeledFormula) bool {
	conjs := mod.ConjSubgoals
	if conjs == nil {
		conjs = mod.LabeledConjs
	}

	// Filter for checkable conjectures using is_check_mod_unprovable.
	// Python: conjs = [x for x in conjs if is_check_mod_unprovable(x)]
	var checkable []*ast.LabeledFormula
	for _, c := range conjs {
		if IsCheckModUnprovable(mod.Cfg, c) {
			checkable = append(checkable, c)
		}
	}

	// Append converted postconditions.
	if len(pcs) > 0 {
		converted := ConvertPostconds(pcs)
		checkable = append(checkable, converted...)
	}

	// Apply line-number filter if set.
	// Python: check_lineno = act.checked_assert.get()
	checkLineno := mod.Cfg.CheckLineno
	if checkLineno != "" {
		var filtered []*ast.LabeledFormula
		for _, c := range checkable {
			if fmt.Sprintf("%d", c.Lineno) == checkLineno {
				filtered = append(filtered, c)
			}
		}
		checkable = filtered
	}

	// Build checkers for the filtered list.
	var checkers []Checker
	for _, c := range checkable {
		checkers = append(checkers, NewConjChecker(mod.Cfg, c, indent))
	}

	return CheckFcsInState(mod, checkers)
}

// CheckSafetyInState checks safety (no assertion violations) in a state.
// Corresponds to Python's check_safety_in_state which creates a
// Checker(lg.Or(), report_pass) and delegates to check_fcs_in_state.
// lg.Or() with no terms is "false", so after dualization the check
// succeeds iff the post-state has no assertion violations.
func CheckSafetyInState(mod *module.Module, reportPass bool) bool {
	checker := NewBaseChecker(mod.Cfg, &lg.Or{}, reportPass, true)
	return CheckFcsInState(mod, []Checker{checker})
}

// GetCheckedActions returns the list of actions to be checked.
// If a specific action is set via the "action" parameter, only that
// action is returned. Otherwise all public actions are returned sorted.
func GetCheckedActions(mod *module.Module) []string {
	cact := mod.Cfg.CheckedAction
	if cact != "" {
		extName := "ext:" + cact
		if mod.PublicActions[extName] {
			cact = extName
		}
	}
	if cact != "" && !mod.PublicActions[cact] {
		return nil
	}
	mod.Cfg.CheckedActionFound = true
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
func GetPrioritizedActions(cfg *module.Config) []string {
	if cfg.PriorityActions == "" {
		return nil
	}
	parts := strings.Split(cfg.PriorityActions, ",")
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
// Matches Python ivy_check.py convert_postconds (lines 418-426):
//   - For symbols that are "old" (old_X), rename to their base name
//   - For updated symbols, map old(s) → __s (pre-state prefix)
//
// The update parameter may be nil, in which case postconds pass through.
func ConvertPostconds(postconds []*ast.LabeledFormula) []*ast.LabeledFormula {
	return ConvertPostcondsWithUpdate(nil, postconds)
}

// ConvertPostcondsWithUpdate is the full version that uses the state's update
// to build a renaming for old symbols. Matches Python convert_postconds(state, postconds).
func ConvertPostcondsWithUpdate(update *tr.Update, postconds []*ast.LabeledFormula) []*ast.LabeledFormula {
	if len(postconds) == 0 {
		return postconds
	}
	if update == nil {
		return postconds
	}

	// Collect all symbols used in postcondition formulas
	renaming := make(map[lg.NodeKey]*lg.Symbol)
	for _, pc := range postconds {
		if pc.Formula == nil {
			continue
		}
		usedSyms := clauseops.UsedSymbolsAST(pc.Formula.(lg.Expr))
		for _, node := range usedSyms {
			sym, ok := node.(*lg.Symbol)
			if !ok {
				continue
			}
			if tr.IsOld(sym.Name) {
				// Python: renaming[s] = itr.old_of(s) — maps old symbol to base name
				renaming[lg.Key(sym)] = lg.NewSymbol(tr.OldOf(sym.Name), sym.CSort)
			}
		}
	}

	// Python: for s in updated: renaming[itr.old(s)] = s.prefix('__')
	for _, s := range update.Modified {
		oldName := tr.Old(s.Name)
		oldSym := lg.NewSymbol(oldName, s.CSort)
		renaming[lg.Key(oldSym)] = lg.NewSymbol("__"+s.Name, s.CSort)
	}

	if len(renaming) == 0 {
		return postconds
	}

	// Python: [x.clone([x.args[0], lut.rename_ast(x.formula, renaming)]) for x in postconds]
	result := make([]*ast.LabeledFormula, len(postconds))
	for i, pc := range postconds {
		renamed := clauseops.RenameAST(pc.Formula.(lg.Expr), renaming)
		result[i] = &ast.LabeledFormula{
			Label:      pc.Label,
			Formula:    renamed,
			Lineno:     pc.Lineno,
			Temporal:   pc.Temporal,
			ID:         pc.ID,
			Explicit:   pc.Explicit,
			Assumed:    pc.Assumed,
			Unprovable: pc.Unprovable,
		}
	}
	return result
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
			if lf, ok2 := args[0].(*ast.LabeledFormula); ok2 {
				return lf.Unprovable
			}
		}
	}
	return false
}

// IsGuaranteeModUnprovable checks guarantee modulo unprovable flag.
// Python: is_unprovable_assert(asrt) == act.check_unprovable.get()
func IsGuaranteeModUnprovable(cfg *module.Config, asrt interface{}) bool {
	return IsUnprovableAssert(asrt) == cfg.OnlyCheckUnprovable
}

// IsCheckModUnprovable checks if a labeled formula should be checked given the unprovable flag.
// Python: lf.unprovable == act.check_unprovable.get()
func IsCheckModUnprovable(cfg *module.Config, lf *ast.LabeledFormula) bool {
	return lf.Unprovable == cfg.OnlyCheckUnprovable
}

// DisplayCex displays a counterexample with a message.
// In Go, the web UI handles display differently from Python's Tk UI.
// Corresponds to Python's display_cex.
func DisplayCex(cfg *module.Config, msg string, ag interface{}) error {
	if cfg.Diagnose {
		// In the Go port, diagnostics are handled by the web UI.
		// The Tk-based display_cex from Python is replaced by web-based CEX rendering.
		return fmt.Errorf("%s (use web UI for interactive diagnostics)", msg)
	}
	return fmt.Errorf("%s", msg)
}

// ShowCounterexample displays a counterexample trace from BMC.
// Corresponds to Python's show_counterexample (lines 76-83).
// Python: universe, path = bmc_res; other_art = AnalysisGraph();
//
//	ag.copy_path(state, other_art, None);
//	for state, value in zip(other_art.states[-len(path):], path):
//	    state.value = value; state.universe = universe
//	gui_art(other_art)
func ShowCounterexample(ag *art.AnalysisGraph, state *art.State, bmcRes interface{}) {
	// bmcRes should be a (universe, path) pair from BMC
	type bmcResult struct {
		Universe interface{}
		Path     []*tr.Update
	}
	res, ok := bmcRes.(*bmcResult)
	if !ok {
		fmt.Println("Counterexample found (use web UI for visualization)")
		return
	}

	otherArt := art.NewAnalysisGraph(ag.Domain)
	ag.CopyPath(state, otherArt, nil)

	// Assign values and universe to the copied states
	pathLen := len(res.Path)
	statesLen := len(otherArt.States)
	startIdx := statesLen - pathLen
	if startIdx < 0 {
		startIdx = 0
	}
	for i, s := range otherArt.States[startIdx:] {
		if i < len(res.Path) {
			s.Value = res.Path[i]
			s.Universe = res.Universe
		}
	}

	// In Go port, display is handled by web UI, not Tk
	fmt.Println("Counterexample found (use web UI for visualization)")
}

// PreprocessAssumedIgnoredProperties applies ACL filtering to axioms,
// properties, and conjectures. Properties matched by the ACL's ignore list
// are removed; those matched by the assume list are admitted as axioms.
// Corresponds to Python's preprocess_assumed_ignored_properties.
func PreprocessAssumedIgnoredProperties(mod *module.Module, aclCfg *acl.Config) {
	if mod == nil {
		return
	}

	getLabel := func(lf *ast.LabeledFormula) string {
		if lf.Label == nil {
			return ""
		}
		return fmt.Sprintf("%v", lf.Label)
	}

	// Print info about changes
	type taggedLF struct {
		lf  *ast.LabeledFormula
		tag string
	}
	var allTagged []taggedLF
	for _, lf := range mod.LabeledAxioms {
		allTagged = append(allTagged, taggedLF{lf, "[axiom]"})
	}
	for _, lf := range mod.LabeledConjs {
		allTagged = append(allTagged, taggedLF{lf, "[conjecture]"})
	}
	for _, lf := range mod.LabeledProps {
		allTagged = append(allTagged, taggedLF{lf, "[property]"})
	}
	fmt.Println("\n  Preprocessing list of axioms, properties, conjectures via user-supplied list of unchecked properties.")
	fmt.Println("\n     The following properties are newly ignored: ")
	for _, t := range allTagged {
		if aclCfg.IsIgnored(getLabel(t.lf)) {
			fmt.Println(t.tag + " " + PrettyLF(t.lf, 8))
		}
	}
	fmt.Println("\n     The following properties are newly assumed: ")
	for _, t := range allTagged {
		if aclCfg.IsAssumed(getLabel(t.lf)) {
			fmt.Println(t.tag + " " + PrettyLF(t.lf, 8))
		}
	}

	// Python line 486: remove assumed non-temporal props and ignored props
	// mod.labeled_props = [lf for lf in mod.labeled_props
	//     if not ((ivy_acl.is_assumed(lf.label) and not(lf.temporal)) or ivy_acl.is_ignored(lf.label))]
	var filteredProps []*ast.LabeledFormula
	for _, lf := range mod.LabeledProps {
		label := getLabel(lf)
		if (aclCfg.IsAssumed(label) && !lf.IsTemporal()) || aclCfg.IsIgnored(label) {
			continue
		}
		filteredProps = append(filteredProps, lf)
	}
	mod.LabeledProps = filteredProps

	// Python line 488: filter axioms
	var filteredAxioms []*ast.LabeledFormula
	for _, lf := range mod.LabeledAxioms {
		if !aclCfg.IsIgnored(getLabel(lf)) {
			filteredAxioms = append(filteredAxioms, lf)
		}
	}
	mod.LabeledAxioms = filteredAxioms

	// Python line 489: assumed non-temporal props+conjs → AssumedInvs
	// mod.assumed_invariants.extend([lf for lf in mod.labeled_props+mod.labeled_conjs
	//     if ivy_acl.is_assumed(lf.label) and not(lf.temporal)])
	for _, lf := range mod.LabeledProps {
		if aclCfg.IsAssumed(getLabel(lf)) && !lf.IsTemporal() {
			mod.AssumedInvs = append(mod.AssumedInvs, lf)
		}
	}
	for _, lf := range mod.LabeledConjs {
		if aclCfg.IsAssumed(getLabel(lf)) && !lf.IsTemporal() {
			mod.AssumedInvs = append(mod.AssumedInvs, lf)
		}
	}

	// Python line 491: filter conjs
	// mod.labeled_conjs = [lf for lf in mod.labeled_conjs
	//     if not(ivy_acl.is_ignored(lf.label)) and not(ivy_acl.is_assumed(lf.label) and not(lf.temporal))]
	var filteredConjs []*ast.LabeledFormula
	for _, lf := range mod.LabeledConjs {
		label := getLabel(lf)
		if aclCfg.IsIgnored(label) {
			continue
		}
		if aclCfg.IsAssumed(label) && !lf.IsTemporal() {
			continue
		}
		filteredConjs = append(filteredConjs, lf)
	}
	mod.LabeledConjs = filteredConjs
}

// MCTactic implements the model-checking tactic.
// Corresponds to Python's mc_tactic (ivy_check.py:805-817).
func MCTactic(prover interface{}, goals []*ast.LabeledFormula, proofNode ast.Node, mod *module.Module) ([]*ast.LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, nil
	}
	goals, err := applyTemporalTacticChain(prover, goals, proofNode)
	if err != nil {
		return nil, err
	}
	// Python: check_subgoals(goals[0:1], method=ivy_mc.check_isolate)
	mcMethod := func() error {
		res, mcErr := mc.CheckIsolate(mod, "mc")
		if mcErr != nil {
			return mcErr
		}
		if res != nil && !res.Proved {
			return fmt.Errorf("model checking failed")
		}
		return nil
	}
	err = CheckSubgoals(goals[0:1], mcMethod, mod)
	return goals[1:], err
}

// VMTTactic exports the verification problem in VMT format and checks it.
// Corresponds to Python's vmt_tactic (ivy_check.py:819-831).
func VMTTactic(prover interface{}, goals []*ast.LabeledFormula, proofNode ast.Node, mod *module.Module) ([]*ast.LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, nil
	}
	goals, err := applyTemporalTacticChain(prover, goals, proofNode)
	if err != nil {
		return nil, err
	}
	// Python: check_subgoals(goals[0:1], method=ivy_vmt.check_isolate)
	vmtMethod := func() error {
		return vmt.CheckIsolate("vmt", mod)
	}
	err = CheckSubgoals(goals[0:1], vmtMethod, mod)
	return goals[1:], err
}

// applyTemporalTacticChain applies tempind, skolemizenp, and l2s_tactic_full
// to TemporalModels goals with non-true formula. Shared by MCTactic and VMTTactic.
// Python: if isinstance(conc, TemporalModels) and not lg.is_true(conc.fmla):
//
//	goals = ivy_tactics.tempind(prover, goals, proof)
//	goals = ivy_tactics.skolemizenp(prover, goals, proof)
//	l2s_pf = proof.clone([proof.args[0], TacticLets()] + list(proof.args[2:]))
//	goals = ivy_l2s.l2s_tactic_full(prover, goals, l2s_pf)
func applyTemporalTacticChain(prover interface{}, goals []*ast.LabeledFormula, proofNode ast.Node) ([]*ast.LabeledFormula, error) {
	if len(goals) == 0 {
		return goals, nil
	}
	goal := goals[0]
	// Check for TemporalModels via the formula directly, since
	// GoalConc returns lg.Expr and TemporalModels is ast.Node.
	// This matches the pattern in CheckSubgoals.
	var tm *ast.TemporalModels
	var isTM bool
	if sb, ok := goal.Formula.(*ast.SchemaBody); ok {
		if c := sb.Conc(); c != nil {
			tm, isTM = c.(*ast.TemporalModels)
		}
	} else if goal.Formula != nil {
		tm, isTM = goal.Formula.(*ast.TemporalModels)
	}
	if !isTM {
		return goals, nil
	}
	if tm.Fmla != nil {
		if fmlaExpr, ok := tm.Fmla.(lg.Expr); ok && lg.IsTrue(fmlaExpr) {
			return goals, nil
		}
	}
	// Apply temporal tactic chain
	pc, _ := prover.(*proof.ProofChecker)
	var err error
	goals, err = tactics.Tempind(pc, goals, proofNode)
	if err != nil {
		return nil, err
	}
	goals, err = tactics.Skolemizenp(pc, goals, proofNode)
	if err != nil {
		return nil, err
	}
	// Python: l2s_pf = proof.clone([proof.args[0], TacticLets()] + list(proof.args[2:]))
	l2sPf := cloneProofWithTacticLets(proofNode)
	goals, err = l2s.L2STacticFull(pc, goals, l2sPf)
	if err != nil {
		return nil, err
	}
	return goals, nil
}

// cloneProofWithTacticLets clones a proof node, replacing its body (args[1])
// with an empty TacticLets. Matches Python:
// proof.clone([proof.args[0], ivy_ast.TacticLets()] + list(proof.args[2:]))
func cloneProofWithTacticLets(proofNode ast.Node) ast.Node {
	if proofNode == nil {
		return nil
	}
	args := proofNode.Args()
	if len(args) < 2 {
		return proofNode
	}
	newArgs := make([]ast.Node, len(args))
	newArgs[0] = args[0]
	newArgs[1] = &ast.TacticLets{}
	copy(newArgs[2:], args[2:])
	return proofNode.Clone(newArgs)
}

// RegisterTactics registers the mc and vmt tactics on the given proof config,
// plus all ivy_tactics.py proof tactics (vcgen, skolemize, skolemizenp, tempind, tempcase, sorry).
func RegisterTactics(proofCfg *proof.Config, mod *module.Module) {
	proofCfg.RegisterTactic("mc", func(pc *proof.ProofChecker, goals []*ast.LabeledFormula, p ast.Node) ([]*ast.LabeledFormula, error) {
		return MCTactic(pc, goals, p, mod)
	})
	proofCfg.RegisterTactic("vmt", func(pc *proof.ProofChecker, goals []*ast.LabeledFormula, p ast.Node) ([]*ast.LabeledFormula, error) {
		return VMTTactic(pc, goals, p, mod)
	})
	// Register all ivy_tactics.py proof tactics.
	tactics.RegisterProofTactics(proofCfg)
}

// Start is the entry point for the ivy_check command.
// Corresponds to Python's start() (ivy_check.py:975-1005).
func Start(args []string) error {
	if len(args) < 1 || !strings.HasSuffix(args[0], ".ivy") {
		return fmt.Errorf("%s", Usage())
	}

	someBounded := false

	mod := module.New()
	if mod.Cfg == nil {
		mod.Cfg = module.NewConfig()
	}

	if mod.Cfg.OptIvyStats {
		fmt.Printf(" +++ IVY_STATS starting checking file %s\n", args[0])
	}

	// Python: ivy_init.source_file(sys.argv[1], ivy_init.open_read(sys.argv[1]), create_isolate=False)
	if err := ivyinit.SourceFile(args[0], mod, mod.Sig, map[string]interface{}{
		"create_isolate": false,
	}); err != nil {
		return err
	}

	// Python: if isinstance(act.checked_assert.get(), iu.LocationTuple) and
	//         act.checked_assert.get().filename == 'none.ivy' and act.checked_assert.get().line == 0:
	//     print('NOT CHECKED'); exit(0)
	if mod.Cfg.CheckLineno == "none.ivy:0" {
		fmt.Println("NOT CHECKED")
		return nil
	}

	// Python: check_module()
	if err := CheckModule(mod); err != nil {
		return err
	}

	// Python: if some_bounded: print("BOUNDED")
	if someBounded {
		fmt.Println("BOUNDED")
	}
	// Python: if ivy_tactics.used_sorry: print("OK, but used 'sorry'")
	// else: print("OK")
	if tactics.UsedSorry {
		fmt.Println("OK, but used 'sorry'")
	} else {
		fmt.Println("OK")
	}
	return nil
}

// Main is the main entry point, wrapping Start with error handling.
// Corresponds to Python's main() (ivy_check.py:1025-1041).
func Main(args []string) int {
	// Python: ivy_alpha.test_bottom = False
	// Python: ivy_init.read_params()
	// Python: if profiling.get(): cProfile.runctx(...) else: start()
	err := Start(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}
