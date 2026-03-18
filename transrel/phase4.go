// phase4.go implements Phase 4.1 functions from the Ivy port.
// These correspond to Python ivy_transrel.py helper functions.
package transrel

import (
	"fmt"
	"sync"

	co "github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
	"github.com/glycerine/goivy/solver"
)

// --- Rename ---

// Rename renames a symbol using the given rename function.
// Corresponds to Python's rename(sym, rn) = sym.rename(rn).
func Rename(sym *lg.Symbol, rn func(string) string) *lg.Symbol {
	newName := rn(sym.Name)
	if newName == sym.Name {
		return sym
	}
	return lg.NewSymbol(newName, sym.CSort)
}

// --- UpdateFrameConstraint ---

// UpdateFrameConstraint returns clauses constraining all updated symbols
// to keep their previous values. For relation symbols (with arity > 0),
// produces clauses asserting old ↔ new for each argument tuple.
// For individual symbols, produces an equality constraint.
// Corresponds to Python's update_frame_constraint.
func UpdateFrameConstraint(update *Update, relations map[string]int) *co.Clauses {
	if update.Modified == nil {
		return co.TrueClauses(nil)
	}
	var fmlas []lg.Node
	for _, sym := range update.Modified {
		arity, isRel := relations[sym.Name]
		if isRel && arity > 0 {
			// Build variables V0, V1, ...
			vars := make([]*lg.Variable, arity)
			varNodes := make([]lg.Node, arity)
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
			iff := &lg.And{Terms: []lg.Node{
				&lg.Or{Terms: []lg.Node{&lg.Not{Body: oldApp}, newApp}},
				&lg.Or{Terms: []lg.Node{oldApp, &lg.Not{Body: newApp}}},
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
		return co.TrueClauses(nil)
	}
	return co.NewClauses(fmlas, nil, nil)
}

// --- SymbolFrameCond ---

// SymbolFrameCond returns a transition relation implying that sym remains
// unchanged. Uses a frame definition (new_sym = sym).
// Corresponds to Python's symbol_frame_cond.
func SymbolFrameCond(sym *lg.Symbol) *co.Clauses {
	def := FrameDefConst(sym, NewConst)
	return co.NewClauses(nil, []*il.Definition{def}, nil)
}

// --- Join ---

// Join computes the parallel join of two updates with an explicit
// vocabulary operator (New or Old). This is the generic version;
// JoinAction and JoinState are the specialized wrappers.
// Corresponds to Python's join(s1, s2, op, axioms).
func Join(u1, u2 *Update, op func(*lg.Symbol) *lg.Symbol, axioms *co.Clauses) *Update {
	return joinUpdate(u1, u2, op, axioms)
}

// --- Ite ---

// Ite computes the conditional update with an explicit vocabulary operator.
// This is the generic version; IteAction and IteState are the specialized wrappers.
// Corresponds to Python's ite(cond, s1, s2, op, axioms).
func Ite(cond lg.Node, u1, u2 *Update, op func(*lg.Symbol) *lg.Symbol, axioms *co.Clauses) *Update {
	return iteUpdate(cond, u1, u2, op, axioms)
}

// --- ClausesImplyFormulaCex ---

// ClausesImplyFormulaCex checks if clauses imply a formula. If so, returns
// (true, nil). Otherwise returns (false, *CounterExample) containing the
// conjunction of clauses with the negation of the formula.
// Corresponds to Python's clauses_imply_formula_cex.
func ClausesImplyFormulaCex(clauses *co.Clauses, fmla lg.Node) (bool, *CounterExample) {
	slv := solver.New()
	implied, err := slv.ClausesImplyFormula(clauses, fmla)
	if err == nil && implied {
		return true, nil
	}
	// Build counterexample: conjoin(clauses, negate_clauses(formula_to_clauses(fmla)))
	negClauses := co.NegateClauses(co.FormulaToClauses(fmla, nil))
	cex := ConjoinClauses(clauses, negClauses)
	return false, &CounterExample{Formula: cex.ToFormula()}
}

// --- Implies ---

// Implies checks whether update s1 implies (refines) update s2,
// using the given axioms and vocabulary operator (old for state, new for action).
// Returns true if s1 implies s2, or a *CounterExample if not.
// Corresponds to Python's implies(s1, s2, axioms, relations, op).
func Implies(s1, s2 *Update, axioms *co.Clauses, op func(*lg.Symbol) *lg.Symbol) (bool, *CounterExample) {
	if s1.Modified == nil && s2.Modified != nil {
		return false, nil
	}

	c1 := co.AndClausesTyped(s1.TR, axioms, DiffFrameConst(s1.Modified, s2.Modified, op, axioms))
	c2 := s2.TR
	p1 := s1.Pre
	p2 := s2.Pre

	c2 = co.AndClausesTyped(c2, DiffFrameConst(s2.Modified, s1.Modified, op, axioms))

	// Check p1 implies p2
	ok1, cex1 := ClausesImplyFormulaCex(p1, p2.ToFormula())
	if !ok1 {
		return false, cex1
	}
	// Check c1 implies c2
	ok2, cex2 := ClausesImplyFormulaCex(c1, c2.ToFormula())
	if !ok2 {
		return false, cex2
	}
	return true, nil
}

// --- ImpliesState ---

// ImpliesState checks if s1 implies s2 in state style (using old vocabulary).
// Corresponds to Python's implies_state(s1, s2, axioms, relations).
func ImpliesState(s1, s2 *Update, axioms *co.Clauses) (bool, *CounterExample) {
	return Implies(s1, s2, axioms, OldConst)
}

// --- ImpliesAction ---

// ImpliesAction checks if s1 implies s2 in action style (using new vocabulary).
// Corresponds to Python's implies_action(s1, s2, axioms, relations).
func ImpliesAction(s1, s2 *Update, axioms *co.Clauses) (bool, *CounterExample) {
	return Implies(s1, s2, axioms, NewConst)
}

// --- Clausify ---

// Clausify converts a formula to Clauses if it isn't already.
// Corresponds to Python's clausify(f).
func Clausify(f interface{}) *co.Clauses {
	if cls, ok := f.(*co.Clauses); ok {
		return cls
	}
	if node, ok := f.(lg.Node); ok {
		return co.FormulaToClauses(node, nil)
	}
	return co.TrueClauses(nil)
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
func RemoveTautEqsClauses(clauses *co.Clauses) *co.Clauses {
	if clauses == nil {
		return nil
	}
	var kept []lg.Node
	for _, f := range clauses.Fmlas {
		if !isTautologyEquality(f) {
			kept = append(kept, f)
		}
	}
	return co.NewClauses(kept, clauses.Defs, clauses.Annot)
}

// isTautologyEquality checks if a formula is of the form x = x.
func isTautologyEquality(f lg.Node) bool {
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
func ExtractPrePostModel(clauses *co.Clauses, model interface{}, updated []*lg.Symbol) (*co.Clauses, *co.Clauses) {
	// Build renaming: sym -> new_sym for updated symbols
	renaming := make(map[string]string, len(updated))
	for _, sym := range updated {
		renaming[sym.Name] = New(sym.Name)
	}

	// Pre-state: ignore skolems and new_ symbols
	preFmlas := filterModelFmlas(clauses, func(name string) bool {
		return IsSkolem(name) || IsNew(name)
	})
	preClauses := co.NewClauses(preFmlas, nil, nil)

	// Post-state: ignore skolems and symbols that were updated (old version)
	postFmlas := filterModelFmlas(clauses, func(name string) bool {
		return IsSkolem(name) || (!IsNew(name) && renaming[name] != "")
	})
	// Rename new_ back to base names
	inverseMap := make(map[string]*lg.Symbol, len(renaming))
	for k, v := range renaming {
		inverseMap[v] = lg.NewSymbol(k, lg.TopS)
	}
	postClauses := co.RenameClauses(co.NewClauses(postFmlas, nil, nil), inverseMap)

	return RemoveTautEqsClauses(preClauses), RemoveTautEqsClauses(postClauses)
}

// filterModelFmlas filters formulas from clauses, keeping those that don't
// reference symbols matched by the ignore predicate.
func filterModelFmlas(clauses *co.Clauses, ignore func(string) bool) []lg.Node {
	var result []lg.Node
	for _, f := range clauses.Fmlas {
		used := usedSymbolNames(f)
		keep := true
		for name := range used {
			if ignore(name) {
				keep = false
				break
			}
		}
		if keep {
			result = append(result, f)
		}
	}
	return result
}

// --- SmallModelClauses ---

// SmallModelClauses finds a small model satisfying the given clauses.
// Corresponds to Python's small_model_clauses.
func SmallModelClauses(cls *co.Clauses, finalCond *co.Clauses, shrink bool) interface{} {
	slv := solver.New()
	model, err := slv.GetSmallModel(cls, nil, nil)
	if err != nil {
		return nil
	}
	return model
}

// --- UseNumerals ---

var (
	useNumeralsMu sync.RWMutex
	useNumeralsVal = true
)

// UseNumerals returns whether numerals should be used in model display.
// Corresponds to Python's use_numerals().
func UseNumerals() bool {
	useNumeralsMu.RLock()
	defer useNumeralsMu.RUnlock()
	return useNumeralsVal
}

// SetUseNumerals sets the use_numerals flag.
func SetUseNumerals(v bool) {
	useNumeralsMu.Lock()
	defer useNumeralsMu.Unlock()
	useNumeralsVal = v
}
