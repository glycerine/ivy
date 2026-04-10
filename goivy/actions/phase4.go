// phase4.go implements Phase 4.1 functions from the Ivy port.
// These correspond to Python ivy_transrel.py helper functions.
package actions

import (
	"fmt"

	il "github.com/glycerine/ivy/goivy/ivylogic"
	iu "github.com/glycerine/ivy/goivy/ivyutils"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/z3bridge"
)

// --- Rename ---

// Rename renames a symbol using the given rename function.
// Corresponds to Python's rename(sym, rn) = sym.rename(rn).
func Rename(sym *lg.Const, rn func(string) string) *lg.Const {
	newName := rn(sym.Name)
	if newName == sym.Name {
		return sym
	}
	return lg.NewConst(newName, sym.CSort)
}

// --- UpdateFrameConstraint ---

// UpdateFrameConstraint returns clauses constraining all updated symbols
// to keep their previous values. For relation symbols (with arity > 0),
// produces clauses asserting old ↔ new for each argument tuple.
// For individual symbols, produces an equality constraint.
// Corresponds to Python's update_frame_constraint.
func UpdateFrameConstraint(update *Update, relations map[string]int) *module.Clauses {
	if update.ModifiedAll {
		return module.TrueClauses(nil)
	}
	var fmlas []lg.Expr
	for _, sym := range update.Modified {
		arity, isRel := relations[sym.Name]
		if isRel && arity > 0 {
			// Build variables V0, V1, ...
			vars := make([]*lg.Variable, arity)
			varNodes := make([]lg.Expr, arity)
			for i := 0; i < arity; i++ {
				v, _ := lg.NewVariable(fmt.Sprintf("V%d", i), lg.TopS)
				vars[i] = v
				varNodes[i] = v
			}
			// sym(V0,...) <-> new_sym(V0,...)
			newSym := NewConst(sym)
			oldApp, _ := lg.NewApply(sym, varNodes...)
			newApp, _ := lg.NewApply(newSym, varNodes...)
			// Iff(old, new) = And(Or(Not(old), new), Or(old, Not(new)))
			iff := &lg.And{Terms: []lg.Expr{
				&lg.Or{Terms: []lg.Expr{&lg.Not{Body: oldApp}, newApp}},
				&lg.Or{Terms: []lg.Expr{oldApp, &lg.Not{Body: newApp}}},
			}}
			// ForAll V0,...: iff
			fmla := il.ForAll(vars, iff)
			fmlas = append(fmlas, fmla)
		} else {
			// sym = new_sym
			newSym := NewConst(sym)
			fmlas = append(fmlas, &lg.Eq{T1: sym, T2: newSym})
		}
	}
	if len(fmlas) == 0 {
		return module.TrueClauses(nil)
	}
	return module.NewClauses(fmlas, nil, nil)
}

// --- SymbolFrameCond ---

// SymbolFrameCond returns a transition relation implying that sym remains
// unchanged. Uses a frame definition (new_sym = sym).
// Corresponds to Python's symbol_frame_cond.
func SymbolFrameCond(sym *lg.Const) *module.Clauses {
	def := FrameDefConst(sym, NewConst)
	return module.NewClauses(nil, []*il.Definition{def}, nil)
}

// --- Join ---

// Join computes the parallel join of two updates with an explicit
// vocabulary operator (New or Old). This is the generic version;
// JoinAction and JoinState are the specialized wrappers.
// Corresponds to Python's join(s1, s2, op, axioms).
func Join(u1, u2 *Update, op func(*lg.Const) *lg.Const, axioms *module.Clauses) *Update {
	return joinUpdate(u1, u2, op, axioms)
}

// --- Ite ---

// Ite computes the conditional update with an explicit vocabulary operator.
// This is the generic version; IteAction and IteState are the specialized wrappers.
// Corresponds to Python's ite(cond, s1, s2, op, axioms).
func Ite(cond lg.Expr, u1, u2 *Update, op func(*lg.Const) *lg.Const, axioms *module.Clauses) *Update {
	return iteUpdate(cond, u1, u2, op, axioms)
}

// --- ClausesImplyFormulaCex ---

// ClausesImplyFormulaCex checks if clauses imply a formula. If so, returns
// (true, nil). Otherwise returns (false, *CounterExample) containing the
// conjunction of clauses with the negation of the formula.
// Corresponds to Python's clauses_imply_formula_cex.
func ClausesImplyFormulaCex(clauses *module.Clauses, fmla lg.Expr) (bool, *CounterExample) {
	slv := z3bridge.NewSolver(nil, nil)
	implied, err := slv.ClausesImplyFormula(clauses, fmla)
	if err == nil && implied {
		return true, nil
	}
	// Build counterexample: conjoin(clauses, negate_clauses(formula_to_clauses(fmla)))
	negClauses := module.NegateClauses(module.FormulaToClauses(fmla, nil))
	cex := ConjoinClauses(clauses, negClauses)
	return false, &CounterExample{Formula: cex.ToFormula()}
}

// --- Implies ---

// Implies checks whether update s1 implies (refines) update s2,
// using the given axioms and vocabulary operator (old for state, new for action).
// Returns true if s1 implies s2, or a *CounterExample if not.
// Corresponds to Python's implies(s1, s2, axioms, relations, op).
func Implies(s1, s2 *Update, axioms *module.Clauses, op func(*lg.Const) *lg.Const) (bool, *CounterExample) {
	if s1.ModifiedAll && !s2.ModifiedAll {
		return false, nil
	}

	c1 := module.AndClausesTyped(s1.TR, axioms, DiffFrameConstUpdate(s1, s2, op, axioms))
	p1 := s1.Pre

	// Python: Clauses-to-Clauses implication path. Python's implies()
	// branches on isinstance(c2, Clauses), but Go's Update always carries
	// *module.Clauses, so the non-Clauses branch is unreachable.
	c2 := s2.TR
	p2 := s2.Pre
	if !c2.IsUniversalFirstOrder() || !p2.IsUniversalFirstOrder() {
		return false, nil
	}
	c2 = module.AndClausesTyped(c2, DiffFrameConst(s2.Modified, s1.Modified, op, axioms))

	// Use z3bridge.ClausesImply for Clauses-to-Clauses implication
	slv := z3bridge.NewSolver(nil, nil)
	ok1, err := slv.ClausesImply(p1, p2)
	if err != nil || !ok1 {
		return false, nil
	}
	ok2, err := slv.ClausesImply(c1, c2)
	if err != nil || !ok2 {
		return false, nil
	}
	return true, nil
}

// --- ImpliesState ---

// ImpliesState checks if s1 implies s2 in state style (using old vocabulary).
// Corresponds to Python's implies_state(s1, s2, axioms, relations).
func ImpliesState(s1, s2 *Update, axioms *module.Clauses) (bool, *CounterExample) {
	return Implies(s1, s2, axioms, OldConst)
}

// --- ImpliesAction ---

// ImpliesAction checks if s1 implies s2 in action style (using new vocabulary).
// Corresponds to Python's implies_action(s1, s2, axioms, relations).
func ImpliesAction(s1, s2 *Update, axioms *module.Clauses) (bool, *CounterExample) {
	return Implies(s1, s2, axioms, NewConst)
}

// --- Clausify ---

// Clausify converts a formula to Clauses if it isn't already.
// Corresponds to Python's clausify(f).
func Clausify(f interface{}) *module.Clauses {
	if cls, ok := f.(*module.Clauses); ok {
		return cls
	}
	if node, ok := f.(lg.Expr); ok {
		return module.FormulaToClauses(node, nil)
	}
	return module.TrueClauses(nil)
}

// --- ClausifyState ---

// ClausifyState ensures the TR and Pre of an update are in Clauses form.
// Corresponds to Python's clausify_state.
func ClausifyState(u *Update) *Update {
	return &Update{
		Modified: u.Modified,
		TR:       u.TR,
		Pre:      u.Pre,
	}
}

// --- RemoveTautEqsClauses ---

// RemoveTautEqsClauses removes tautological equalities (x = x) from clauses.
// Corresponds to Python's remove_taut_eqs_clauses.
func RemoveTautEqsClauses(clauses *module.Clauses) *module.Clauses {
	if clauses == nil {
		return nil
	}
	var kept []lg.Expr
	for _, f := range clauses.Fmlas {
		if !isTautologyEquality(f) {
			kept = append(kept, f)
		}
	}
	return module.NewClauses(kept, clauses.Defs, clauses.Annot)
}

// isTautologyEquality checks if a formula is of the form x = x.
func isTautologyEquality(f lg.Expr) bool {
	eq, ok := f.(*lg.Eq)
	if !ok {
		return false
	}
	return eq.T1 != nil && eq.T2 != nil && eq.T1.Equal(eq.T2)
}

// --- ExtractPrePostModel ---

// ExtractPrePostModel extracts pre-state and post-state models from a
// satisfying model of a two-vocabulary formula.
// Returns (pre_clauses, post_clauses).
// Corresponds to Python's extract_pre_post_model.
func ExtractPrePostModel(cfg *iu.IvyUtilsConfig, clauses *module.Clauses, model *z3bridge.ModelResult, updated []*lg.Const) (*module.Clauses, *module.Clauses) {
	// Build renaming: sym -> new_sym for updated symbols
	renaming := make(map[string]string, len(updated))
	for _, sym := range updated {
		renaming[sym.Name] = New(sym.Name)
	}

	numerals := cfg.UseNumerals

	// Pre-state: ignore skolems and new_ symbols
	slv := z3bridge.NewSolver(nil, nil)
	preClauses, err := slv.ClausesModelToClausesWithModel(
		clauses,
		model,
		func(s *lg.Const) bool {
			return IsSkolem(s.Name) || IsNew(s.Name)
		},
		numerals,
	)
	if err != nil {
		preClauses = module.TrueClauses(nil)
	}

	// Post-state: ignore skolems and symbols that were updated (old version)
	postClauses, err := slv.ClausesModelToClausesWithModel(
		clauses,
		model,
		func(s *lg.Const) bool {
			return IsSkolem(s.Name) || (!IsNew(s.Name) && renaming[s.Name] != "")
		},
		numerals,
	)
	if err != nil {
		postClauses = module.TrueClauses(nil)
	}

	// Rename new_ back to base names
	inverseMap := make(map[lg.NodeKey]*lg.Const, len(renaming))
	for _, sym := range updated {
		newSym := lg.NewConst(New(sym.Name), sym.CSort)
		inverseMap[lg.Key(newSym)] = lg.NewConst(sym.Name, sym.CSort)
	}
	postClauses = module.RenameClauses(postClauses, inverseMap)

	return RemoveTautEqsClauses(preClauses), RemoveTautEqsClauses(postClauses)
}

// --- SmallModelClauses ---

// SmallModelClauses finds a small model satisfying the given clauses.
// Uses uninterpreted sorts from the current module's signature as sorts to minimize.
// Corresponds to Python's small_model_clauses.
// SmallModelClauses finds a small model satisfying the given clauses.
// Returns the model result and the solver that produced it.
// The solver is needed by callers that want to call ClausesModelToClausesWithModel
// or build a HerbrandModel. This matches Python where get_small_model returns a
// HerbrandModel that wraps both the solver and the model together.
func SmallModelClauses(cls *module.Clauses, finalCond []z3bridge.FinalCond, shrink bool, m *module.Module) (*z3bridge.ModelResult, *z3bridge.Solver) {
	// Python: get_small_model(cls, ivy_logic.uninterpreted_sorts(), [], final_cond=final_cond, shrink=shrink)
	var sorts []lg.Sort
	if m != nil && m.Sig != nil {
		sorts = il.UninterpretedSorts(m.Sig)
	}
	var sig *il.Sig
	var sopts *module.SolverOptions
	if m != nil {
		sig = m.Sig
		if m.Cfg != nil {
			sopts = m.Cfg.SolverOpts
		}
	}
	slv := z3bridge.NewSolver(sig, sopts)
	model, err := slv.GetSmallModelWithCond(cls, sorts, nil, finalCond, shrink)
	if err != nil {
		return nil, slv
	}
	return model, slv
}
