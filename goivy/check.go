// Package check is the top-level verification driver for Ivy.
// It orchestrates checking of properties, conjectures, temporals,
// and isolates. This corresponds to Python's ivy_check.py.
package goivy

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/glycerine/ivy/goivy/xtracer"
)

const checkCheckPrecondFalse = false
const checkCheckPrecondTrue = true

// --- Checker interface and implementations ---

// Checker is the interface for verification condition checkers.
// Each checker wraps a formula condition to be checked against a state.
type Checker interface {
	// Cond returns the clause set representing the negated condition to check.
	Cond() *Clauses
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
	// GetAnnot returns an action/annotation pair for trace reconstruction, or nil.
	GetAnnot() *ActionAnnotation
	// Failed returns whether this checker has failed.
	Failed() bool
	// GetLF returns the labeled formula if this is a conjecture checker.
	GetLF() *LabeledFormula
}

// --- BaseChecker ---

// BaseChecker is the default checker implementation.
// It wraps a formula conjecture and checks it against a state by
// negating it (dualizing) and checking satisfiability.
type BaseChecker struct {
	Mod        *Module
	FC         *Clauses
	ReportPass bool
	Inverted   bool
	FailedFlag bool
}

// NewBaseChecker creates a BaseChecker for the given conjecture formula.
// If invert is true (the default), the formula is dualized for checking.
// Faithful port of Python ivy_check.py Checker.__init__ (lines 218-226).
func NewBaseChecker(mod *Module, conj Expr, reportPass bool, invert bool) *BaseChecker {
	fc := FormulaToClauses(conj, nil)
	if invert {
		// Python: def witness(v): return lg.Symbol('@'+v.name, v.sort)
		//         self.fc = lut.dual_clauses(self.fc, witness)
		witness := func(v *LogicVariable) Expr {
			return VarToSkolem("@", v)
		}
		fc = DualClauses(fc, witness, mod.Instantiator)
	}
	return &BaseChecker{
		Mod:        mod,
		FC:         fc,
		ReportPass: reportPass,
		Inverted:   invert,
	}
}

func (c *BaseChecker) Cond() *Clauses { return c.FC }
func (c *BaseChecker) Start() {
	if c.ReportPass {
		fmt.Print("...\n")
	}
}
func (c *BaseChecker) Sat() bool {
	// Python: return self._pass() if act.check_unprovable.get() else self.fail()
	if c.Mod.Cfg.OnlyCheckUnprovable {
		return c.Pass()
	}
	return c.Fail()
}
func (c *BaseChecker) Unsat() bool {
	// Python: return self.fail() if act.check_unprovable.get() else self._pass()
	if c.Mod.Cfg.OnlyCheckUnprovable {
		return c.Fail()
	}
	return c.Pass()
}
func (c *BaseChecker) Assume() bool                { return false }
func (c *BaseChecker) GetAnnot() *ActionAnnotation { return nil }
func (c *BaseChecker) Failed() bool                { return c.FailedFlag }
func (c *BaseChecker) GetLF() *LabeledFormula      { return nil }

func (c *BaseChecker) Fail() bool {
	fmt.Println("FAIL")
	c.Mod.Cfg.Failures++
	c.FailedFlag = true
	// Python: return not (diagnose.get() or opt_trace.get()) or act.check_unprovable.get()
	return !(c.Mod.Cfg.Diagnose || c.Mod.Cfg.OptTrace) || c.Mod.Cfg.OnlyCheckUnprovable
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
	LF     *LabeledFormula
	Indent int
}

// NewConjChecker creates a ConjChecker for the given labeled formula.
func NewConjChecker(mod *Module, lf *LabeledFormula, indent int) *ConjChecker {
	base := NewBaseChecker(mod, lf.Formula.(Expr), true, true)
	return &ConjChecker{
		BaseChecker: *base,
		LF:          lf,
		Indent:      indent,
	}
}

func (c *ConjChecker) Start() {
	fmt.Print(PrettyLF(c.LF, c.Indent), " ")
	fmt.Print("...\n")
}

func (c *ConjChecker) GetAnnot() *ActionAnnotation {
	// Python: return self.lf.annot if hasattr(self.lf,'annot') else None
	if c.LF != nil && c.LF.Annot != nil {
		return c.LF.Annot
	}
	return nil
}

func (c *ConjChecker) GetLF() *LabeledFormula { return c.LF }

// --- ConjAssumer ---

// ConjAssumer treats a conjecture as assumed (not checked).
type ConjAssumer struct {
	BaseChecker
	LF *LabeledFormula
}

// NewConjAssumer creates a ConjAssumer for the given labeled formula.
func NewConjAssumer(mod *Module, lf *LabeledFormula) *ConjAssumer {
	base := NewBaseChecker(mod, lf.Formula.(Expr), false, false)
	return &ConjAssumer{
		BaseChecker: *base,
		LF:          lf,
	}
}

func (c *ConjAssumer) Start() {
	fmt.Println(PrettyLF(c.LF, 8) + "  [assumed]")
}

func (c *ConjAssumer) Assume() bool           { return true }
func (c *ConjAssumer) GetLF() *LabeledFormula { return c.LF }

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
//
// However note:
// It is confusing because compiler also has CheckProperties.
// This one is not used anywhere. The python port is not
// used anywhere either. So we just ported dead code.
// Comment out for now to avoid confusion.
//
// Both sides have the same pattern: the compiler version (of
// CheckProperties)is live, the check version is
// effectively dead code (Python's is completely uncalled;
// Go's is only referenced from a test). They're parallel
// dead code inherited from the port.
/*
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
*/

// CheckConjectures checks conjectures in the given state.
// Corresponds to Python's check_conjectures which calls
// itp.undecided_conjectures(state) and launches GUI diagnosis
// if any fail. In the Go port the analysis-graph / solver
// interaction is not yet wired up, so this is a no-op success.
// CheckConjectures checks conjectures against the current state.
// Matches Python ivy_check.py check_conjectures (lines 104-117):
//   - Calls itp.undecided_conjectures(state) to find failing ones
//   - Reports error if any fail
func CheckConjectures(cfg *Config, kind, msg string, ag *AnalysisGraph, state *InterpState) error {
	failed := UndecidedConjectures(state)
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
func CheckTemporals(mod *Module) error {
	// Python: pmap = dict((prop.id,p) for prop,p in mod.proofs)
	pmap := make(map[int64]interface{})
	for _, pe := range mod.Proofs {
		pmap[pe.Formula.ID] = pe.Proof
	}

	// Python: pc = ivy_proof.ProofChecker(mod.labeled_axioms+mod.assumed_invariants,
	//                                     mod.definitions, mod.schemata)
	pcAxioms := make([]*LabeledFormula, 0, len(mod.LabeledAxioms)+len(mod.AssumedInvs))
	pcAxioms = append(pcAxioms, mod.LabeledAxioms...)
	pcAxioms = append(pcAxioms, mod.AssumedInvs...)

	// Diagnostic: dump axioms fed to the temporal proof checker
	xtracer.Trace("check.CheckTemporals axiomDump nAxioms=%d nLabeledAxioms=%d nAssumedInvs=%d", len(pcAxioms), len(mod.LabeledAxioms), len(mod.AssumedInvs))
	for idx, ax := range pcAxioms {
		xtracer.Trace("check.CheckTemporals axiomDump[%d] HASH canon=%v", idx, ax.Canon())
	}

	pc := NewProofChecker(mod.Cfg.ProofCfg, mod, pcAxioms, mod.Definitions, ModuleSchemataToAst(mod.Schemata))

	// Use the ACL config loaded by CheckModule and stored on the module.
	aclCfg := mod.AclCfg

	for _, prop := range mod.LabeledProps {
		if !prop.IsTemporal() {
			continue
		}
		xtracer.Trace("check.CheckTemporals prop start label=%s nAssumedInvs=%d nLabeledConjs=%d nLabeledProps=%d nLabeledAxioms=%d",
			fmt.Sprint(prop.Label), len(mod.AssumedInvs), len(mod.LabeledConjs), len(mod.LabeledProps), len(mod.LabeledAxioms))

		// Python: if prop.assumed or opt_unchecked_properties.get() and ivy_acl.is_assumed(prop.label):
		propLabel := fmt.Sprint(prop.Label)
		isAssumedByACL := aclCfg != nil && aclCfg.IsAssumed(propLabel)
		if prop.Assumed || isAssumedByACL {
			fmt.Println("  ivy_check temporal: admitting axiom...\n", PrettyLF(prop, 0))
			if isAssumedByACL {
				fmt.Printf("     ... admitting %s as axiom because it is an externally assumed property and unchecked property file is supplied.\n", propLabel)
			}
			pc.AdmitAxiom(prop)
		} else {
			fmt.Print("\n    The following temporal property is being proved:\n\n")
			fmt.Print(PrettyLF(prop, 4) + " ...\n")

			// Python: proof = pmap.get(prop.id, None)
			pf := pmap[prop.ID]

			// Python: propn = ivy_proof.normalize_goal(prop)
			propn := NormalizeGoal(mod.Cfg.AstCfg, prop)

			// Python: model = itmp.normal_program_from_module(im.module)
			model := NormalProgramFromModule(mod)

			// Python: subgoal = prop.clone([prop.args[0], ivy_ast.TemporalModels(model, propn.args[1])])
			xtracer.Trace("check.temporal propn.Formula type=%s", TypeName(propn.Formula))
			tm := mod.Cfg.AstCfg.NewTemporalModels(model, propn.Formula)
			subgoal := prop.Clone([]Node{prop.Label, tm}).(*LabeledFormula)
			xtracer.Trace("check.temporal subgoal.Formula type=%s", TypeName(subgoal.Formula))

			subgoals := []*LabeledFormula{subgoal}

			// Python: subgoals = pc.admit_proposition(prop, proof, subgoals)
			var pfNode Node
			if pf != nil {
				pfNode, _ = pf.(Node)
			}
			xtracer.Trace("check.temporal proof type=%s pfNode type=%s", TypeName(pf), TypeName(pfNode))
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
func GetConjs(mod *Module) *Clauses {
	var fmlas []Expr
	all := append(mod.LabeledConjs, mod.AssumedInvs...)
	for _, lf := range all {
		if !lf.Explicit && !lf.Unprovable {
			fmlas = append(fmlas, lf.Formula.(Expr))
		}
	}
	return NewClauses(fmlas, nil, EmptyAnnotation{})
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
func ApplyConjProofs(mod *Module) {
	// Python: pc = ivy_proof.ProofChecker(mod.labeled_axioms+mod.assumed_invariants, mod.definitions, mod.schemata)
	// The proof package uses ast.LabeledFormula (with ast.Node fields) while
	// module uses ast.LabeledFormula (with lg.Expr fields). These are separate
	// type hierarchies — a porting mistake (Python has one LabeledFormula class).
	// Until the two are unified, we attempt proof application when the formula's
	// concrete type satisfies ast.Node, and fall through otherwise.
	pcAxioms := make([]*LabeledFormula, 0, len(mod.LabeledAxioms)+len(mod.AssumedInvs))
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
	pcDefs := make([]*LabeledFormula, 0, len(mod.Definitions))
	for _, lf := range mod.Definitions {
		if alf := ModuleLFToAstLF(lf); alf != nil {
			pcDefs = append(pcDefs, alf)
		}
	}
	pc := NewProofChecker(mod.Cfg.ProofCfg, mod, pcAxioms, pcDefs, ModuleSchemataToAst(mod.Schemata))

	pmap := make(map[int64]interface{})
	for _, pe := range mod.Proofs {
		pmap[pe.Formula.ID] = pe.Proof
	}

	conjs := make([]*LabeledFormula, 0)
	for _, lf := range mod.LabeledConjs {
		if p, hasProof := pmap[lf.ID]; hasProof {
			// Python: subgoals = pc.admit_proposition(lf, proof)
			astLF := ModuleLFToAstLF(lf)
			astProof, _ := p.(Node)
			if astLF != nil && astLF.Formula != nil {
				// Python: subgoals = pc.admit_proposition(lf, proof)
				subgoals, err := pc.AdmitProposition(astLF, astProof)
				if err == nil && len(subgoals) > 0 {
					// Python: subgoals = list(map(ivy_compiler.theorem_to_property, subgoals))
					for _, sg := range subgoals {
						modSG := AstLFToModuleLF(sg)
						modSG = TheoremToProperty(modSG, mod)
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
func CheckFcsInState(mod *Module, checkers []Checker) bool {
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
func CheckFcsInStateWithAG(mod *Module, ag *AnalysisGraph, post *State, checkers []Checker) bool {
	// Python check_fcs_in_state has NO early return for empty fcs.
	// Even with zero checkers it builds history, fetches background_theory,
	// emits the checkFcsNormalPath xtrace, and calls history.satisfy(...).

	// Get history and background theory
	var history *History
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
func checkFcsTracePath(mod *Module, ag *AnalysisGraph, post *State,
	history *History, axioms *Clauses, checkers []Checker) bool {

	if history == nil || history.Post == nil {
		// No history — fall back to normal path
		return checkFcsNormalPath(mod, ag, post, history, axioms, checkers)
	}

	// Python: clauses = history.post; clauses = lut.and_clauses(clauses, axioms)
	clauses := AndClausesTyped(history.Post, axioms)

	// Python: ffcs = filter_fcs(fcs)
	ffcs := FilterCheckers(checkers, mod.Cfg.CheckLineno)

	// Python: model = itr.small_model_clauses(clauses, ffcs, shrink=True)
	var finalConds []FinalCond
	for _, fc := range ffcs {
		finalConds = append(finalConds, fc)
	}
	model, modelSlv := SmallModelClauses(clauses, finalConds, true, mod)

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
				mclauses = AndClausesTyped(mclauses, c.Cond())
			}
		}

		// Python: vocab = lut.used_symbols_clauses(mclauses)
		vocabMap := mclauses.Symbols()
		vocab := make([]*Const, 0, vocabMap.Len())
		traceVocab := make([]Expr, 0, vocabMap.Len())
		for _, sym := range vocabMap.All() {
			if c, ok := sym.(*Const); ok {
				vocab = append(vocab, c)
				traceVocab = append(traceVocab, c)
			}
		}

		// Python: handler = ivy_trace.Trace(mclauses, model, vocab)
		hm := NewHerbrandModel(modelSlv, model.Solver, model.Model, vocab)
		handler := NewTraceForModule(mod.Cfg.IuCfg, mod, mclauses, hm, traceVocab, true)
		if modClauses, err := modelSlv.ClausesModelToClausesWithModel(mclauses, model, nil, true); err == nil {
			handler.SetEqsFromClauses(modClauses)
		}

		// Python: thing = failed[-1].get_annot()
		thing := failed[len(failed)-1].GetAnnot()
		if thing == nil {
			// Python: actions = [mod.actions[a] if isinstance(a,str) else a for a in history.actions]
			//         action = act.Sequence(*actions); annot = clauses.annot
			// In Go, history.Actions is []lg.Expr. String action names are *lg.Const.
			var actionExprs []Expr
			for _, a := range history.Actions {
				// Python: mod.actions[a] if isinstance(a, str) else a
				if sym, ok := a.(*Const); ok {
					if act, exists := mod.Actions.Get2(sym.Name); exists {
						if actAction, ok := act.(ActionsAction); ok {
							actionExprs = append(actionExprs, actAction)
							continue
						}
					}
				}
				actionExprs = append(actionExprs, a)
			}
			action := NewSequence(actionExprs...)
			MatchAnnotation(action, clauses.Annot, handler, mod)
		} else {
			// Python: action, annot = thing
			MatchAnnotation(thing.Action, thing.Annot, handler, mod)
		}
		handler.End()

		// C5 / Python ivy_check.py:406-407 (and ivy_l2s.py:1310-1313):
		// Apply the trace hook attached by L2S tactics, if any. The hook is
		// stored on Module/LabeledFormula.
		if mod.TraceHook != nil {
			handler = mod.TraceHook(handler, ffcs)
		}

		// Python: ff = failed[0]
		// handler.is_cti = lut.formula_to_clauses(ff.lf.formula) if isinstance(ff, ConjChecker) else None
		ff := failed[0]
		if cc, ok := ff.(*ConjChecker); ok {
			handler.IsCti = FormulaToClauses(cc.LF.Formula.(Expr), nil)
		}

		// Python ivy_check.py:411-415:
		//   if not opt_trace.get(): gui_art(handler)
		//   else: print(str(handler)); exit(0)
		if mod.Cfg.OptTrace {
			fmt.Println(handler.String())
			os.Exit(0)
		} else {
			// Python: gui_art(handler) — passes the trace handler. The handler
			// carries the IsCti clauses for the GUI's CTI display mode.
			if err := GuiArt(mod, handler, handler.IsCti); err != nil {
				fmt.Fprintf(os.Stderr, "GuiArt: %v\n", err)
			}
			// Python: gui_art ends with exit(1) after the Tk mainloop returns.
			os.Exit(1)
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
func checkFcsNormalPath(mod *Module, ag *AnalysisGraph, post *State,
	history *History, axioms *Clauses, checkers []Checker) bool {

	// Python: filter_fcs(fcs) — filter by check_lineno
	filteredCheckers := FilterCheckers(checkers, mod.Cfg.CheckLineno)

	// Convert checkers to solver.FinalCond for history.SatisfyWithCond.
	// Python: filter_fcs(fcs) ALWAYS returns a list (possibly empty), never None.
	// Initialize as an empty slice (non-nil) so GetSmallModelWithCond takes
	// the "Python list" branch even when no checkers survived filtering.
	finalConds := make([]FinalCond, 0, len(filteredCheckers))
	for _, fc := range filteredCheckers {
		finalConds = append(finalConds, fc)
	}

	if history != nil {
		xtracer.Trace("check.checkFcsNormalPath history path postFmlas=%d postDefs=%d axiomFmlas=%d axiomDefs=%d checkers=%d",
			len(history.Post.Fmlas), len(history.Post.Defs), len(axioms.Fmlas), len(axioms.Defs), len(filteredCheckers))
		// Python: gmc = lambda cls, final_cond: itr.small_model_clauses(cls, final_cond, shrink=diagnose.get())
		gmc := func(cls *Clauses, fc []FinalCond) *ModelResult {
			mr, _ := SmallModelClauses(cls, fc, mod.Cfg.Diagnose, mod)
			return mr
		}

		// Python: res = history.satisfy(axioms, gmc, filter_fcs(fcs))
		res := history.SatisfyWithCond(axioms, gmc, finalConds)

		// Python: if res is not None and diagnose.get(): show_counterexample(ag, post, res)
		if res != nil && mod.Cfg.Diagnose {
			ShowCounterexample(ag, post, res)
		}
	} else {
		// No history — fall back to direct solver check.
		// This happens when ag/post are nil (e.g., property checking with true pre-state).
		// Python never takes this path (always provides ag/post), but Go has
		// CheckFcsInState which can reach here. We must invoke checker callbacks
		// just as history.SatisfyWithCond would.
		baseClauses := TrueClauses(EmptyAnnotation{})
		combined := AndClausesTyped(baseClauses, axioms)

		for _, fc := range filteredCheckers {
			fc.Start()
			if fc.Assume() {
				continue
			}
			// Check this individual checker against the combined clauses.
			fcConds := []FinalCond{fc}
			mr, _ := SmallModelClauses(combined, fcConds, mod.Cfg.Diagnose, mod)
			if mr != nil {
				// SAT — checker fails (condition is satisfiable)
				if !fc.Sat() {
					break
				}
			} else {
				// UNSAT — checker passes
				fc.Unsat()
			}
		}
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
func CheckConjsInState(mod *Module, indent int, pcs []*LabeledFormula) bool {
	conjs := mod.ConjSubgoals
	if conjs == nil {
		conjs = mod.LabeledConjs
	}

	// Filter for checkable conjectures using is_check_mod_unprovable.
	// Python: conjs = [x for x in conjs if is_check_mod_unprovable(x)]
	var checkable []*LabeledFormula
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

	// Apply line-number filter if set (file:line format).
	// Python: check_lineno = act.checked_assert.get()
	checkLineno := mod.Cfg.CheckLineno
	if checkLineno != "" {
		var filtered []*LabeledFormula
		for _, c := range checkable {
			if c.GetLineno().FileLineKey() == checkLineno {
				filtered = append(filtered, c)
			}
		}
		checkable = filtered
	}

	// Build checkers for the filtered list.
	var checkers []Checker
	for _, c := range checkable {
		checkers = append(checkers, NewConjChecker(mod, c, indent))
	}

	return CheckFcsInState(mod, checkers)
}

// CheckSafetyInState checks safety (no assertion violations) in a state.
// Corresponds to Python's check_safety_in_state which creates a
// Checker(lg.Or(), report_pass) and delegates to check_fcs_in_state.
// lg.Or() with no terms is "false", so after dualization the check
// succeeds iff the post-state has no assertion violations.
func CheckSafetyInState(mod *Module, reportPass bool) bool {
	checker := NewBaseChecker(mod, &LogicOr{}, reportPass, true)
	return CheckFcsInState(mod, []Checker{checker})
}

// GetCheckedActions returns the list of actions to be checked.
// If a specific action is set via the "action" parameter, only that
// action is returned. Otherwise all public actions are returned sorted.
func GetCheckedActions(mod *Module) []string {
	cact := mod.Cfg.CheckedAction
	if cact != "" {
		extName := "ext:" + cact
		if mod.PublicActions.Get(extName) {
			cact = extName
		}
	}
	if cact != "" && !mod.PublicActions.Get(cact) {
		return nil
	}
	mod.Cfg.CheckedActionFound = true
	if cact != "" {
		return []string{cact}
	}
	result := make([]string, 0, mod.PublicActions.Len())
	for name := range mod.PublicActions.All() {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

// GetPrioritizedActions returns the list of prioritized actions parsed
// from the "prioritize" parameter. Each name is prefixed with "ext:".
func GetPrioritizedActions(cfg *Config) []string {
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

// ConvertPostconds converts postconditions WITHOUT an update context.
// Python's convert_postconds(state, postconds) always receives a state
// with an update. This no-update convenience wrapper passes postconds
// through UNCHANGED (no old-symbol renaming). Callers with a post-state
// should use ConvertPostcondsWithUpdate or CheckConjsInStateWithAG.
func ConvertPostconds(postconds []*LabeledFormula) []*LabeledFormula {
	return ConvertPostcondsWithUpdate(nil, postconds)
}

// ConvertPostcondsWithUpdate is the full version that uses the state's update
// to build a renaming for old symbols. Matches Python convert_postconds(state, postconds).
func ConvertPostcondsWithUpdate(update *Update, postconds []*LabeledFormula) []*LabeledFormula {
	if len(postconds) == 0 {
		return postconds
	}
	if update == nil {
		return postconds
	}

	// Collect all symbols used in postcondition formulas
	renaming := make(map[NodeKey]*Const)
	for _, pc := range postconds {
		if pc.Formula == nil {
			continue
		}
		usedSyms := UsedSymbolsAST(pc.Formula.(Expr))
		for _, sym := range usedSyms.All() {
			if s, ok := sym.(*Const); ok && IsOld(s.Name) {
				renaming[Key(s)] = NewConst(OldOf(s.Name), s.CSort)
			}
		}
	}

	// Python: for s in updated: renaming[itr.old(s)] = s.prefix('__')
	for _, s := range update.Modified {
		oldName := LogicOld(s.Name)
		oldSym := NewConst(oldName, s.CSort)
		renaming[Key(oldSym)] = NewConst("__"+s.Name, s.CSort)
	}

	if len(renaming) == 0 {
		return postconds
	}

	// Python: [x.clone([x.args[0], lut.rename_ast(x.formula, renaming)]) for x in postconds]
	result := make([]*LabeledFormula, len(postconds))
	for i, pc := range postconds {
		renamed := RenameAST(pc.Formula.(Expr), renaming)
		result[i] = pc.Clone([]Node{pc.Label, renamed}).(*LabeledFormula)
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
			if lf, ok2 := args[0].(*LabeledFormula); ok2 {
				return lf.Unprovable
			}
		}
	}
	return false
}

// IsGuaranteeModUnprovable checks guarantee modulo unprovable flag.
// Python: is_unprovable_assert(asrt) == act.check_unprovable.get()
func IsGuaranteeModUnprovable(cfg *Config, asrt interface{}) bool {
	return IsUnprovableAssert(asrt) == cfg.OnlyCheckUnprovable
}

// IsCheckModUnprovable checks if a labeled formula should be checked given the unprovable flag.
// Python: lf.unprovable == act.check_unprovable.get()
func IsCheckModUnprovable(cfg *Config, lf *LabeledFormula) bool {
	return lf.Unprovable == cfg.OnlyCheckUnprovable
}

// DisplayCex displays a counterexample with a message.
// Corresponds to Python's display_cex (ivy_check.py:54-60).
//
// Python:
//
//	def display_cex(msg,ag):
//	    if diagnose.get():
//	        from . import tk_ui as ui
//	        iu.set_parameters({'mode':'induction'})
//	        ui.ui_main_loop(ag)
//	        exit(1)
//	    raise iu.IvyError(None,msg)
func DisplayCex(mod *Module, msg string, ag interface{}) error {
	if mod.Cfg.Diagnose {
		// Python: ui.ui_main_loop(ag); exit(1)
		if err := GuiArt(mod, ag, nil); err != nil {
			fmt.Fprintf(os.Stderr, "DisplayCex GuiArt: %v\n", err)
		}
		os.Exit(1)
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
func ShowCounterexample(ag *AnalysisGraph, state *State, bmcRes interface{}) {
	// bmcRes should be the (universe, path) pair returned by History.Satisfy.
	res, ok := bmcRes.(*SatisfyResult)
	if !ok {
		otherArt := NewAnalysisGraph(ag.Domain)
		ag.CopyPath(state, otherArt, nil)
		// Python: gui_art(other_art)
		if err := GuiArt(ag.Domain, otherArt, nil); err != nil {
			fmt.Fprintf(os.Stderr, "ShowCounterexample GuiArt: %v\n", err)
		}
		return
	}

	otherArt := NewAnalysisGraph(ag.Domain)
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
			s.Universe = res.Universes
		}
	}

	// Python: gui_art(other_art)
	if err := GuiArt(ag.Domain, otherArt, nil); err != nil {
		fmt.Fprintf(os.Stderr, "ShowCounterexample GuiArt: %v\n", err)
	}
}

// PreprocessAssumedIgnoredProperties applies ACL filtering to axioms,
// properties, and conjectures. Properties matched by the ACL's ignore list
// are removed; those matched by the assume list are admitted as axioms.
// Corresponds to Python's preprocess_assumed_ignored_properties.
func PreprocessAssumedIgnoredProperties(mod *Module, aclCfg *ACLConfig) {
	if mod == nil {
		return
	}

	getLabel := func(lf *LabeledFormula) string {
		if lf.Label == nil {
			return ""
		}
		return fmt.Sprintf("%v", lf.Label)
	}

	// Print info about changes
	type taggedLF struct {
		lf  *LabeledFormula
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
	var filteredProps []*LabeledFormula
	for _, lf := range mod.LabeledProps {
		label := getLabel(lf)
		if (aclCfg.IsAssumed(label) && !lf.IsTemporal()) || aclCfg.IsIgnored(label) {
			continue
		}
		filteredProps = append(filteredProps, lf)
	}
	mod.LabeledProps = filteredProps

	// Python line 488: filter axioms
	var filteredAxioms []*LabeledFormula
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
	var filteredConjs []*LabeledFormula
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
func MCTactic(prover interface{}, goals []*LabeledFormula, proofNode Node, mod *Module) ([]*LabeledFormula, error) {
	xtracer.Trace("check.MCTactic ENTER nGoals=%d", len(goals))
	if len(goals) == 0 {
		return nil, nil
	}
	goals, err := applyTemporalTacticChain(prover, goals, proofNode)
	if err != nil {
		return nil, err
	}
	xtracer.Trace("check.MCTactic postTacticChain nGoals=%d", len(goals))
	// Python: check_subgoals(goals[0:1], method=ivy_mc.check_isolate)
	mcMethod := func(m *Module) error {
		res, mcErr := MCCheckIsolate(m, "mc")
		if mcErr != nil {
			return mcErr
		}
		if res != nil && !res.Proved {
			if res.Error != nil {
				return res.Error
			}
			if res.DecodedTrace != nil {
				return &MCCounterexampleFailure{Trace: res.DecodedTrace}
			}
			return fmt.Errorf("model checking failed")
		}
		return nil
	}
	err = CheckSubgoals(goals[0:1], mcMethod, mod)
	xtracer.Trace("check.MCTactic EXIT nRemainingGoals=%d err=%v", len(goals[1:]), err)
	return goals[1:], err
}

// VMTTactic exports the verification problem in VMT format and checks it.
// Corresponds to Python's vmt_tactic (ivy_check.py:819-831).
func VMTTactic(prover interface{}, goals []*LabeledFormula, proofNode Node, mod *Module) ([]*LabeledFormula, error) {
	if len(goals) == 0 {
		return nil, nil
	}
	goals, err := applyTemporalTacticChain(prover, goals, proofNode)
	if err != nil {
		return nil, err
	}
	// Python: check_subgoals(goals[0:1], method=ivy_vmt.check_isolate)
	vmtMethod := func(m *Module) error {
		return VMTCheckIsolate("vmt", m)
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
func applyTemporalTacticChain(prover interface{}, goals []*LabeledFormula, proofNode Node) ([]*LabeledFormula, error) {
	if len(goals) == 0 {
		return goals, nil
	}
	goal := goals[0]
	// Check for TemporalModels via the formula directly, since
	// GoalConc returns lg.Expr and TemporalModels is ast.Node.
	// This matches the pattern in CheckSubgoals.
	var tm *TemporalModels
	var isTM bool
	if sb, ok := goal.Formula.(*SchemaBody); ok {
		if c := sb.Conc(); c != nil {
			tm, isTM = c.(*TemporalModels)
		}
	} else if goal.Formula != nil {
		tm, isTM = goal.Formula.(*TemporalModels)
	}
	if !isTM {
		return goals, nil
	}
	if tm.Fmla != nil {
		if fmlaExpr, ok := tm.Fmla.(Expr); ok && IsTrue(fmlaExpr) {
			return goals, nil
		}
	}
	// Apply temporal tactic chain
	pc, _ := prover.(ProofCheckerInterface)
	var err error
	goals, err = Tempind(pc, goals, proofNode)
	if err != nil {
		return nil, err
	}
	goals, err = Skolemizenp(pc, goals, proofNode)
	if err != nil {
		return nil, err
	}
	// Python: l2s_pf = proof.clone([proof.args[0], TacticLets()] + list(proof.args[2:]))
	l2sPf := cloneProofWithTacticLets(proofNode)
	goals, err = L2STacticFull(pc, goals, l2sPf)
	if err != nil {
		return nil, err
	}
	return goals, nil
}

// cloneProofWithTacticLets clones a proof node, replacing its body (args[1])
// with an empty TacticLets. Matches Python:
// proof.clone([proof.args[0], ivy_ast.TacticLets()] + list(proof.args[2:]))
func cloneProofWithTacticLets(proofNode Node) Node {
	if proofNode == nil {
		return nil
	}
	args := proofNode.Args()
	if len(args) < 2 {
		return proofNode
	}
	newArgs := make([]Node, len(args))
	newArgs[0] = args[0]
	newArgs[1] = proofNode.GetAstConfig().NewTacticLets(nil)
	copy(newArgs[2:], args[2:])
	return proofNode.Clone(newArgs)
}

// RegisterTactics registers the mc and vmt tactics on the given proof config,
// plus all ivy_tactics.py proof tactics (vcgen, skolemize, skolemizenp, tempind, tempcase, sorry).
func RegisterTactics(proofCfg *ProofConfig, mod *Module) {
	proofCfg.RegisterTactic("mc", func(pc ProofCheckerInterface, goals []*LabeledFormula, p Node) ([]*LabeledFormula, error) {
		return MCTactic(pc, goals, p, pc.GetModule())
	})
	proofCfg.RegisterTactic("vmt", func(pc ProofCheckerInterface, goals []*LabeledFormula, p Node) ([]*LabeledFormula, error) {
		return VMTTactic(pc, goals, p, pc.GetModule())
	})
	// Register all ivy_tactics.py proof tactics.
	RegisterProofTactics(proofCfg)
	// Register temporal and l2s tactics — Python does this at import time.
	RegisterTemporalTactics(proofCfg)
	RegisterL2STactics(proofCfg)
	RegisterAutoInstTactics(proofCfg)
}

/*
// deprecated. duplicate of Start().
// is the entry point for the ivy_check command.
// Corresponds to Python's start() (ivy_check.py:975-1005).
func deprecated_dup_start_9828(args []string) error {
	if len(args) < 1 || !strings.HasSuffix(args[0], ".ivy") {
		return fmt.Errorf("%s", Usage())
	}

	mod := module.New()
	if mod.Cfg == nil {
		mod.Cfg = module.NewConfig()
	}

	// Python ivy_check.py:1028-1029: some_bounded = False at start() entry
	mod.Cfg.SomeBounded = false
	mod.Cfg.ProofCfg = TacticNewConfig()
	RegisterTactics(mod.Cfg.ProofCfg, mod)

	if mod.Cfg.OptIvyStats {
		fmt.Printf(" +++ IVY_STATS starting checking file %s\n", args[0])
	}

	// Python: ivy_init.source_file(sys.argv[1], ivy_init.open_read(sys.argv[1]), create_isolate=False)
	if err := compiler.SourceFile(args[0], mod, mod.Sig, map[string]interface{}{
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
	if err := CheckModule(mod); err != nil { // in Start()
		return err
	}

	// Python ivy_check.py:1046-1047: if some_bounded: print("BOUNDED")
	if mod.Cfg.SomeBounded {
		fmt.Println("BOUNDED")
	}
	// Python: if ivy_tactics.used_sorry: print("OK, but used 'sorry'")
	// else: print("OK")
	if mod.Cfg.UsedSorry {
		fmt.Println("OK, but used 'sorry'")
	} else {
		fmt.Println("OK")
	}
	return nil
}
*/

// Main is the main entry point, wrapping Start with error handling.
// Corresponds to Python's main() (ivy_check.py:1025-1041).
func Main(args []string) int {
	// Python: ivy_alpha.test_bottom = False
	// Python: ivy_init.read_params()
	// Python: if profiling.get(): cProfile.runctx(...) else: start()
	err := Start(args, nil)
	if err != nil {
		var reported *ReportedFailure
		if errors.As(err, &reported) {
			return 1
		}
		printTopLevelError(err)
		return 1
	}
	return 0
}

// printTopLevelError prints err to stderr. For *ParseError it prints the
// message raw, matching Python ParseError.__repr__ format. For anything
// else it prepends "error: " as the existing convention.
func printTopLevelError(err error) {
	var pe *ParseError
	if errors.As(err, &pe) {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
}

// StartWithConfig is like Start but accepts a pre-populated Config.
// This allows the CLI to parse key=value parameters and apply them
// before the module is created, matching how Python's ivy_init.read_params()
// sets Parameter objects before start() creates the Module context.
// If cfg is nil, we will supply a default from module.NewConfig().
func Start(args []string, cfg *Config) error {
	//fmt.Printf("check.go:1228 top Start().\n")

	if cfg == nil {
		cfg = NewConfig()
	}
	if len(args) >= 1 {
		// Python: set_macro_finder(True) at ivy_solver.py:53 during module import.
		// Emit matching trace before check.start.
		{
			truthStr := "True"
			if !cfg.SolverOpts.MacroFinder {
				truthStr = "False"
			}
			xtracer.Trace("ivy_solver.py:45 set_macro_finder() ENTER truth=%s", truthStr)
		}
		xtracer.Trace("check.start ENTER file=%s", args[0])
	}
	if len(args) < 1 || !strings.HasSuffix(args[0], ".ivy") {
		return fmt.Errorf("%s", Usage())
	}

	return startLoaded(args[0], cfg, func(mod *Module) error {
		// Python: ivy_init.source_file(sys.argv[1], ivy_init.open_read(sys.argv[1]), create_isolate=False)
		return SourceFile(args[0], mod, mod.Sig, map[string]interface{}{
			"create_isolate": false,
		})
	})
}

// StartSourceWithConfig is the browser/goldweb counterpart of Start. It runs
// the goivy_check pipeline against an in-memory .ivy source string while using
// filename for traces and diagnostics. This keeps cmd/goivy_check out of the
// wasm build while preserving its command semantics.
func StartSourceWithConfig(filename, source string, cfg *Config) error {
	fmt.Printf("check.go:1261 top StartSourceWithConfig().\n")

	if cfg == nil {
		cfg = NewConfig()
	}
	if filename == "" {
		filename = "browser_input.ivy"
	}
	if !strings.HasSuffix(filename, ".ivy") {
		return fmt.Errorf("%s", Usage())
	}

	truthStr := "True"
	if !cfg.SolverOpts.MacroFinder {
		truthStr = "False"
	}
	xtracer.Trace("ivy_solver.py:45 set_macro_finder() ENTER truth=%s", truthStr)
	xtracer.Trace("check.start ENTER file=%s", filename)

	xoff := os.Getenv("XTRACE_OFF")
	fmt.Printf("check.go:1281 on stdout after 1st two xtraces in StartSourceWithConfig(). xtracer.Enabled=%v; xoff='%v'\n", xtracer.Enabled, xoff)
	fmt.Fprintf(os.Stderr, "check.go:1282 on stderr after 1st two xtraces in StartSourceWithConfig(). xtracer.Enabled=%v\n", xtracer.Enabled)

	return startLoaded(filename, cfg, func(mod *Module) error {
		return SourceString(filename, source, mod, mod.Sig, map[string]interface{}{
			"create_isolate": false,
		})
	})
}

func startLoaded(filename string, cfg *Config, load func(mod *Module) error) error {
	// Python ivy_check.py:1028-1029: some_bounded = False at start() entry.
	// Reset the flag in case the caller is reusing a Config across runs.
	cfg.SomeBounded = false

	mod := New()
	mod.Cfg = cfg
	mod.Cfg.ProofCfg = TacticNewConfig()
	RegisterTactics(mod.Cfg.ProofCfg, mod)

	if mod.Cfg.OptIvyStats {
		fmt.Printf(" +++ IVY_STATS starting checking file %s\n", filename)
	}

	if err := load(mod); err != nil {
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
	if err := CheckModule(mod); err != nil { // in StartWithConfig()
		return err
	}

	// Python ivy_check.py:1046-1047: if some_bounded: print("BOUNDED")
	if mod.Cfg.SomeBounded {
		fmt.Println("BOUNDED")
	}
	// Python: if ivy_tactics.used_sorry: print("OK, but used 'sorry'")
	// else: print("OK")
	if mod.Cfg.UsedSorry {
		fmt.Println("OK, but used 'sorry'")
	} else {
		fmt.Println("OK")
	}
	xtracer.Trace("check.start EXIT")
	return nil
}

// MainWithConfig is like Main but accepts a pre-populated Config.
func MainWithConfig(args []string, cfg *Config) (code int) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(*IvyError); ok {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				code = 1
				return
			}
			panic(r)
		}
	}()
	err := Start(args, cfg)
	if err != nil {
		var reported *ReportedFailure
		if errors.As(err, &reported) {
			return 1
		}
		printTopLevelError(err)
		return 1
	}
	return 0
}
