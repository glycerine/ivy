// phase4.go implements Phase 4.1 functions from the Ivy port.
// These correspond to Python ivy_transrel.py helper functions.
package goivy

import (
	"fmt"
)

// --- Rename ---

// Rename renames a symbol using the given rename function.
// Corresponds to Python's rename(sym, rn) = sym.rename(rn).
func Rename(sym *Const, rn func(string) string) *Const {
	newName := rn(sym.Name)
	if newName == sym.Name {
		return sym
	}
	return NewConst(newName, sym.CSort)
}

// --- UpdateFrameConstraint ---

// UpdateFrameConstraint returns clauses constraining all updated symbols
// to keep their previous values. For relation symbols (with arity > 0),
// produces clauses asserting old ↔ new for each argument tuple.
// For individual symbols, produces an equality constraint.
// Corresponds to Python's update_frame_constraint.
func UpdateFrameConstraint(update *Update, relations map[string]int) *Clauses {
	if update.ModifiedAll {
		return TrueClauses(nil)
	}
	var fmlas []Expr
	for _, sym := range update.Modified {
		arity, isRel := relations[sym.Name]
		if isRel && arity > 0 {
			// Build variables V0, V1, ...
			vars := make([]*LogicVariable, arity)
			varNodes := make([]Expr, arity)
			for i := 0; i < arity; i++ {
				v, _ := NewVariable(fmt.Sprintf("V%d", i), TopS)
				vars[i] = v
				varNodes[i] = v
			}
			// sym(V0,...) <-> new_sym(V0,...)
			newSym := NewActionConst(sym)
			oldApp, _ := NewApply(sym, varNodes...)
			newApp, _ := NewApply(newSym, varNodes...)
			// Iff(old, new) = LogicAnd(Or(Not(old), new), Or(old, Not(new)))
			iff := &LogicAnd{Terms: []Expr{
				&LogicOr{Terms: []Expr{&LogicNot{Body: oldApp}, newApp}},
				&LogicOr{Terms: []Expr{oldApp, &LogicNot{Body: newApp}}},
			}}
			// ForAll V0,...: iff
			fmla := IvyForAll(vars, iff)
			fmlas = append(fmlas, fmla)
		} else {
			// sym = new_sym
			newSym := NewActionConst(sym)
			fmlas = append(fmlas, &Eq{T1: sym, T2: newSym})
		}
	}
	if len(fmlas) == 0 {
		return TrueClauses(nil)
	}
	return NewClauses(fmlas, nil, nil)
}

// --- SymbolFrameCond ---

// SymbolFrameCond returns a transition relation implying that sym remains
// unchanged. Uses a frame definition (new_sym = sym).
// Corresponds to Python's symbol_frame_cond.
func SymbolFrameCond(sym *Const) *Clauses {
	def := FrameDefConst(sym, NewActionConst)
	return NewClauses(nil, []*IvyDefinition{def}, nil)
}

// --- Join ---

// Join computes the parallel join of two updates with an explicit
// vocabulary operator (ActionNewName or Old). This is the generic version;
// JoinAction and JoinState are the specialized wrappers.
// Corresponds to Python's join(s1, s2, op, axioms).
func Join(u1, u2 *Update, op func(*Const) *Const, axioms *Clauses) *Update {
	return joinUpdate(u1, u2, op, axioms)
}

// --- ActionIte ---

// ActionIte computes the conditional update with an explicit vocabulary operator.
// This is the generic version; IteAction and IteState are the specialized wrappers.
// Corresponds to Python's ite(cond, s1, s2, op, axioms).
func ActionIte(cond Expr, u1, u2 *Update, op func(*Const) *Const, axioms *Clauses) *Update {
	return iteUpdate(cond, u1, u2, op, axioms)
}

// --- ClausesImplyFormulaCex ---

// ClausesImplyFormulaCex checks if clauses imply a formula. If so, returns
// (true, nil). Otherwise returns (false, *CounterExample) containing the
// conjunction of clauses with the negation of the formula.
// Corresponds to Python's clauses_imply_formula_cex.
func ClausesImplyFormulaCex(mod *Module, clauses *Clauses, fmla Expr) (bool, *CounterExample) {
	slv := NewSolver(mod, nil)
	implied, err := slv.ClausesImplyFormula(clauses, fmla)
	if err == nil && implied {
		return true, nil
	}
	// Build counterexample: conjoin(clauses, negate_clauses(formula_to_clauses(fmla)))
	negClauses := NegateClauses(FormulaToClauses(fmla, nil))
	cex := ConjoinClauses(clauses, negClauses)
	return false, &CounterExample{Formula: cex.ToFormula()}
}

// --- ActionImplies ---

// ActionImplies checks whether update s1 implies (refines) update s2,
// using the given axioms and vocabulary operator (old for state, new for action).
// Returns true if s1 implies s2, or a *CounterExample if not.
// Corresponds to Python's implies(s1, s2, axioms, relations, op).
func ActionImplies(mod *Module, s1, s2 *Update, axioms *Clauses, op func(*Const) *Const) (bool, *CounterExample) {
	if s1.ModifiedAll && !s2.ModifiedAll {
		return false, nil
	}

	c1 := AndClausesTyped(s1.TR, axioms, DiffFrameConstUpdate(s1, s2, op, axioms))
	p1 := s1.Pre

	// Python: Clauses-to-Clauses implication path. Python's implies()
	// branches on isinstance(c2, Clauses), but Go's Update always carries
	// *module.Clauses, so the non-Clauses branch is unreachable.
	c2 := s2.TR
	p2 := s2.Pre
	if !c2.IsUniversalFirstOrder() || !p2.IsUniversalFirstOrder() {
		return false, nil
	}
	c2 = AndClausesTyped(c2, DiffFrameConst(s2.Modified, s1.Modified, op, axioms))

	// Use z3bridge.ClausesImply for Clauses-to-Clauses implication
	slv := NewSolver(mod, nil)
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
func ImpliesState(mod *Module, s1, s2 *Update, axioms *Clauses) (bool, *CounterExample) {
	return ActionImplies(mod, s1, s2, axioms, OldConst)
}

// --- ImpliesAction ---

// ImpliesAction checks if s1 implies s2 in action style (using new vocabulary).
// Corresponds to Python's implies_action(s1, s2, axioms, relations).
func ImpliesAction(mod *Module, s1, s2 *Update, axioms *Clauses) (bool, *CounterExample) {
	return ActionImplies(mod, s1, s2, axioms, NewActionConst)
}

// --- Clausify ---

// Clausify converts a formula to Clauses if it isn't already.
// Corresponds to Python's clausify(f).
func Clausify(f interface{}) *Clauses {
	if cls, ok := f.(*Clauses); ok {
		return cls
	}
	if node, ok := f.(Expr); ok {
		return FormulaToClauses(node, nil)
	}
	return TrueClauses(nil)
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
func RemoveTautEqsClauses(clauses *Clauses) *Clauses {
	if clauses == nil {
		return nil
	}
	var kept []Expr
	for _, f := range clauses.Fmlas {
		if !isTautologyEquality(f) {
			kept = append(kept, f)
		}
	}
	return NewClauses(kept, clauses.Defs, clauses.Annot)
}

// isTautologyEquality checks if a formula is of the form x = x.
func isTautologyEquality(f Expr) bool {
	eq, ok := f.(*Eq)
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
func ExtractPrePostModel(mod *Module, cfg *IvyUtilsConfig, clauses *Clauses, model *ModelResult, updated []*Const) (*Clauses, *Clauses) {
	// Build renaming: sym -> new_sym for updated symbols
	renaming := make(map[string]string, len(updated))
	for _, sym := range updated {
		renaming[sym.Name] = ActionNewName(sym.Name)
	}

	numerals := cfg.UseNumerals

	// Pre-state: ignore skolems and new_ symbols
	slv := NewSolver(mod, nil)
	preClauses, err := slv.ClausesModelToClausesWithModel(
		clauses,
		model,
		func(s *Const) bool {
			return IsSkolem(s.Name) || IsNew(s.Name)
		},
		numerals,
	)
	if err != nil {
		preClauses = TrueClauses(nil)
	}

	// Post-state: ignore skolems and symbols that were updated (old version)
	postClauses, err := slv.ClausesModelToClausesWithModel(
		clauses,
		model,
		func(s *Const) bool {
			return IsSkolem(s.Name) || (!IsNew(s.Name) && renaming[s.Name] != "")
		},
		numerals,
	)
	if err != nil {
		postClauses = TrueClauses(nil)
	}

	// Rename new_ back to base names
	inverseMap := make(map[NodeKey]*Const, len(renaming))
	for _, sym := range updated {
		newSym := NewConst(ActionNewName(sym.Name), sym.CSort)
		inverseMap[Key(newSym)] = NewConst(sym.Name, sym.CSort)
	}
	postClauses = RenameClauses(postClauses, inverseMap)

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
func SmallModelClauses(cls *Clauses, finalCond []FinalCond, shrink bool, m *Module) (*ModelResult, *Solver) {
	// Python: get_small_model(cls, ivy_logic.uninterpreted_sorts(), [], final_cond=final_cond, shrink=shrink)
	var sorts []Sort
	if m != nil && m.Sig != nil {
		sorts = UninterpretedSorts(m.Sig)
	}
	var sopts *SolverOptions
	if m != nil && m.Cfg != nil {
		sopts = m.Cfg.SolverOpts
	}
	slv := NewSolver(m, sopts)
	model, err := slv.GetSmallModelWithCond(cls, sorts, nil, finalCond, shrink)
	if err != nil {
		return nil, slv
	}
	return model, slv
}
