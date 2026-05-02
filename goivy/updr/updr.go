// Ported to Go from ivy_updr.py.

package updr

import (
	"fmt"
	"strings"

	"github.com/glycerine/ivy/goivy/actions"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/module"
	"github.com/glycerine/ivy/goivy/z3bridge"
)

// UPDRResult holds the result of running UPDR on a module.
type UPDRResult struct {
	Valid     bool   // true if all properties hold
	Invariant string // human-readable invariant (if valid)
	Error     string // description of counterexample (if invalid)
	Stats     UPDRStats
}

// UPDRStats captures performance statistics.
type UPDRStats struct {
	NumFrames      int
	NumIterations  int
	NumSATQueries  int
	NumClauses     int
	NumUnivClauses int
}

// forwardClausesIvy renames clauses from current-state to next-state vocabulary
// at the Ivy level (before Z3 conversion). This matches Python's forward_clauses
// in ivy_updr.py which uses lu.rename_clauses + lu.used_symbols_clauses.
//
// Symbols in inflex (inflexible/constant) are NOT renamed.
func forwardClausesIvy(clauses *module.Clauses, inflex map[lg.NodeKey]bool) *module.Clauses {
	used := module.UsedSymbolsClauses(clauses)
	subs := make(map[lg.NodeKey]*lg.Const)
	for k, sym := range used.All() {
		c, ok := sym.(*lg.Const)
		if !ok {
			continue
		}
		if c.Name == "=" {
			continue
		}
		if inflex[k] {
			continue
		}
		subs[k] = lg.NewConst(actions.New(c.Name), c.CSort)
	}
	return module.RenameClauses(clauses, subs)
}

// CheckModule runs UPDR on an Ivy module to verify safety properties.
//
// Faithful port of Python ivy_updr.py __main__ block:
//  1. Extracts the transition system from the module
//  2. Classifies symbols: flexible (updated), inflexible (constant), skolem
//  3. Computes error states via reverse image of error action
//  4. Makes frames explicit via FrameUpdate
//  5. Converts init/trans/bad/axioms to Z3 via solver
//  6. Runs PDR.Run()
//  7. Returns result with statistics
func CheckModule(mod *module.Module) (*UPDRResult, error) {
	if mod == nil {
		return nil, fmt.Errorf("updr: nil module")
	}

	// Python: axioms = state.domain.background_theory(state.in_scope)
	axioms := mod.BackgroundTheory(nil)

	// Python: err_act = ag.actions["error"].update(ag.domain, state.in_scope)
	// Python: error = tr.reverse_image([], axioms, err_act)
	var errorClauses *module.Clauses
	if errAct, ok := mod.Actions.Get2("error"); ok {
		updateCtx := makeUpdateContext(mod)
		errUpd := actions.GetUpdate(errAct, updateCtx)
		if errUpd != nil {
			errorClauses = actions.ReverseImage(
				module.TrueClauses(nil), axioms, errUpd)
		}
	}
	if errorClauses == nil {
		errorClauses = module.TrueClauses(nil)
	}

	// Python: actions_list = [ag.actions[lab] for lab in ag.actions if lab != "error"]
	updateCtx := makeUpdateContext(mod)
	var updates []*actions.Update
	for name, act := range mod.Actions.All() {
		if name == "error" {
			continue
		}
		upd := actions.GetUpdate(act, updateCtx)
		if upd != nil {
			updates = append(updates, upd)
		}
	}

	if len(updates) == 0 {
		return &UPDRResult{Valid: true, Invariant: "trivially safe (no actions)"}, nil
	}

	// Python: flex = set(sym for u in updates for sym in u[0])
	flexSet := make(map[lg.NodeKey]*lg.Const)
	for _, u := range updates {
		for _, sym := range u.Modified {
			flexSet[lg.Key(sym)] = sym
		}
	}
	var flexConsts []*lg.Const
	for _, sym := range flexSet {
		flexConsts = append(flexConsts, sym)
	}

	// Python: inflex = set(sym for sym in all_syms if sym not in flex and not is_new and not is_skolem)
	inflexSet := make(map[lg.NodeKey]bool)
	var inflexConsts []*lg.Const
	if mod.Sig != nil {
		for name, entry := range mod.Sig.Symbols.All() {
			if actions.IsNew(name) || actions.IsSkolem(name) {
				continue
			}
			c := lg.NewConst(name, entry.Sort)
			k := lg.Key(c)
			if _, isFlex := flexSet[k]; isFlex {
				continue
			}
			inflexSet[k] = true
			inflexConsts = append(inflexConsts, c)
		}
	}

	// Python: updates = [tr.frame_update(u, flex, sig) for u in updates]
	for i, u := range updates {
		updates[i] = actions.FrameUpdate(u, flexConsts)
	}

	// Python: init = forward_clauses(state.clauses, inflex)
	initClauses := mod.InitCond
	if initClauses == nil {
		initClauses = module.TrueClauses(nil)
	}
	initClauses = forwardClausesIvy(initClauses, inflexSet)

	// Python: error = forward_clauses(error, inflex)
	errorClauses = forwardClausesIvy(errorClauses, inflexSet)

	// Python: axioms += forward_clauses(axioms, inflex)
	fwdAxioms := forwardClausesIvy(axioms, inflexSet)
	axioms = module.AndClausesTyped(axioms, fwdAxioms)

	// Convert everything to Z3
	solver := z3bridge.NewSolver(mod, nil)

	// Python: init_z3 = sv.clauses_to_z3(init)
	initZ3, err := solver.ClausesToZ3(initClauses)
	if err != nil {
		return nil, fmt.Errorf("updr: init clauses_to_z3: %w", err)
	}

	// Python: rho_z3 = z3.Or(*[sv.clauses_to_z3(lu.simplify_clauses(a[1])) for a in updates])
	var rhoTerms []z3bridge.Expr
	for _, u := range updates {
		simplified := module.SimplifyClauses(u.TR)
		z3tr, zerr := solver.ClausesToZ3(simplified)
		if zerr != nil {
			return nil, fmt.Errorf("updr: trans clauses_to_z3: %w", zerr)
		}
		rhoTerms = append(rhoTerms, z3tr)
	}
	var rhoZ3 z3bridge.Expr
	if len(rhoTerms) == 1 {
		rhoZ3 = rhoTerms[0]
	} else {
		rhoZ3 = solver.Context().Or(rhoTerms...)
	}

	// Python: bad_z3 = sv.clauses_to_z3(error)
	badZ3, err := solver.ClausesToZ3(errorClauses)
	if err != nil {
		return nil, fmt.Errorf("updr: bad clauses_to_z3: %w", err)
	}

	// Python: background_z3 = sv.clauses_to_z3(axioms)
	backgroundZ3, err := solver.ClausesToZ3(axioms)
	if err != nil {
		return nil, fmt.Errorf("updr: background clauses_to_z3: %w", err)
	}

	// Python: lsyms = [(ns(sym), ns(tr.new(sym))) for sym in flex]
	var lsyms [][2]z3bridge.Expr
	for _, sym := range flexConsts {
		z3Cur, cerr := solver.FormulaToZ3(sym)
		if cerr != nil {
			continue
		}
		newSym := lg.NewConst(actions.New(sym.Name), sym.CSort)
		z3Next, nerr := solver.FormulaToZ3(newSym)
		if nerr != nil {
			continue
		}
		lsyms = append(lsyms, [2]z3bridge.Expr{z3Cur, z3Next})
	}

	// Python: gsyms = [ns(sym) for sym in inflex]
	var gsyms []z3bridge.Expr
	for _, sym := range inflexConsts {
		z3Sym, serr := solver.FormulaToZ3(sym)
		if serr != nil {
			continue
		}
		gsyms = append(gsyms, z3Sym)
	}

	// Python: relations_z3 = [x for x in gsyms if bool-sorted] + [x for _,x in lsyms if bool-sorted]
	var relationsZ3 []z3bridge.Expr
	for _, g := range gsyms {
		relationsZ3 = append(relationsZ3, g)
	}
	for _, pair := range lsyms {
		relationsZ3 = append(relationsZ3, pair[1])
	}

	// Python: pdr = pd.PDR(init_z3, rho_z3, bad_z3, background_z3, gsyms, lsyms, relations_z3, True)
	pdr := NewPDR(solver.Context(), initZ3, rhoZ3, badZ3,
		&backgroundZ3, gsyms, lsyms, relationsZ3, true)

	result := pdr.Run()

	stats := UPDRStats{
		NumFrames:     pdr.N,
		NumIterations: pdr.IterationCount,
		NumSATQueries: pdr.SATQueryCount,
	}

	if result.Valid {
		invStr := result.Invariant.String()
		stats.NumClauses = NumClauses(result.Invariant)
		stats.NumUnivClauses = NumUnivClauses(result.Invariant)
		return &UPDRResult{
			Valid:     true,
			Invariant: invStr,
			Stats:     stats,
		}, nil
	}

	return &UPDRResult{
		Valid: false,
		Error: fmt.Sprintf("counterexample trace with %d steps", len(result.Trace)),
		Stats: stats,
	}, nil
}

// makeUpdateContext creates an UpdateContext suitable for computing action
// updates from a module. Corresponds to Python's domain/in_scope pairing.
func makeUpdateContext(mod *module.Module) *actions.UpdateContext {
	pvars := make(map[string]bool)
	if mod.Sig != nil {
		for name := range mod.Sig.Symbols.All() {
			pvars[name] = true
		}
	}
	return &actions.UpdateContext{
		Domain: mod,
		PVars:  pvars,
		GetAction: func(name string) actions.Action {
			if raw, ok := mod.Actions.Get2(name); ok {
				if a, aok := raw.(actions.Action); aok {
					return a
				}
			}
			return nil
		},
	}
}

// ForwardClauses renames clauses from current-state vocabulary to
// next-state vocabulary, skipping inflexible (constant) symbols.
//
// This is used when propagating clauses forward in the frame sequence.
// Clauses about inflexible symbols don't need renaming since those
// symbols don't change between states.
func ForwardClauses(ctx *z3bridge.Z3Context, clauses z3bridge.Expr,
	x0, xn []z3bridge.Expr) z3bridge.Expr {
	if len(x0) == 0 {
		return clauses
	}
	return ctx.Substitute(clauses, x0, xn)
}

// NumClauses counts the number of top-level conjuncts in an expression.
// If the expression is a conjunction (And), it returns the number of children.
// Otherwise returns 1.
func NumClauses(e z3bridge.Expr) int {
	s := e.String()
	if s == "true" {
		return 0
	}
	if s == "false" {
		return 1
	}
	// Count top-level "and" conjuncts by looking at the string representation.
	// This is a heuristic; a proper implementation would inspect the AST.
	if strings.HasPrefix(s, "(and") {
		count := 0
		depth := 0
		for _, ch := range s {
			switch ch {
			case '(':
				depth++
			case ')':
				depth--
			case ' ':
				if depth == 1 {
					count++
				}
			}
		}
		return count
	}
	return 1
}

// NumUnivClauses counts the number of universally quantified clauses
// in an expression. A clause is "universal" if it is a ForAll or
// if it is a ground clause (implicitly universally quantified).
func NumUnivClauses(e z3bridge.Expr) int {
	s := e.String()
	return strings.Count(s, "forall") + NumClauses(e)
}
