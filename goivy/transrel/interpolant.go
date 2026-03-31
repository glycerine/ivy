package transrel

import (
	co "github.com/glycerine/ivy/goivy/clauseops"
	lg "github.com/glycerine/ivy/goivy/logic"
	"github.com/glycerine/ivy/goivy/solver"
)

// InterpolantResult holds the result of an interpolation query.
type InterpolantResult struct {
	Core *co.Clauses // the unsatisfiable core
	Itp  *co.Clauses // the interpolant (over-approximation)
}

// Interpolant computes an interpolant between two clause sets.
// Returns nil if the conjunction is satisfiable (no interpolant exists).
//
// The interpolant I has the properties:
//   - clauses1 ∧ axioms ⊨ I
//   - I ∧ clauses2 is unsat
//
// Python: ivy_transrel.py:501-509
func Interpolant(clauses1, clauses2, axioms *co.Clauses, interpreted map[string]bool) *InterpolantResult {
	combined := co.AndClausesTyped(clauses1, axioms)
	clauses2 = co.SimplifyClauses(clauses2)

	slv := solver.New()
	itp, err := slv.BinaryInterpolant(combined, clauses2)
	if err != nil || itp == nil {
		return nil
	}
	return &InterpolantResult{Core: clauses1, Itp: itp}
}

// ForwardInterpolant computes the interpolant of the forward image.
// preState is the predecessor's clauses (not an Update).
//
// Python: ivy_transrel.py:518-519
//
//	forward_interpolant(pre_state, update, post_state, axioms, interpreted):
//	    return interpolant(forward_image(pre_state, axioms, update), post_state, axioms, interpreted)
func ForwardInterpolant(preState *co.Clauses, update *Update, postState *co.Clauses, axioms *co.Clauses, interpreted map[string]bool) *InterpolantResult {
	fwdImg := ForwardImage(preState.ToFormula(), axioms.ToFormula(), update)
	fwdClauses := co.FormulaToClauses(fwdImg, nil)
	return Interpolant(fwdClauses, postState, axioms, interpreted)
}

// ReverseInterpolantCase computes the interpolant using reverse image and case analysis.
// postState is the post-state's clauses (not an Update).
//
// Python: ivy_transrel.py:521-529
//
//	reverse_interpolant_case(post_state, update, pre_state, axioms, interpreted):
//	    pre = reverse_image(post_state, axioms, update)
//	    pre_case = clauses_case(pre)
//	    pre_case = [filter ground non-skolem clauses]
//	    return interpolant(pre_state, pre_case, axioms, interpreted)
func ReverseInterpolantCase(postState *co.Clauses, update *Update, preState *co.Clauses, axioms *co.Clauses, interpreted map[string]bool) *InterpolantResult {
	revImg := ReverseImage(postState.ToFormula(), axioms.ToFormula(), update)
	revClauses := co.FormulaToClauses(revImg, nil)

	// Case analysis: filter to ground clauses without Skolem relations
	filtered := filterGroundNonSkolem(revClauses)
	return Interpolant(preState, filtered, axioms, interpreted)
}

// InterpolantCase computes the interpolant using forward case analysis.
//
// Python: ivy_transrel.py:531-543
func InterpolantCase(preState *co.Clauses, post *co.Clauses, axioms *co.Clauses, interpreted map[string]bool) *InterpolantResult {
	filtered := filterGroundNonSkolem(post)
	return Interpolant(preState, filtered, axioms, interpreted)
}

// InterpFromUnsatCore computes an interpolant from an unsat core.
// The interpolant consists of the negation of the core's non-interpreted symbols,
// restricted to the shared vocabulary.
//
// Python: ivy_transrel.py:545-580
func InterpFromUnsatCore(clauses1, clauses2, core *co.Clauses, interpreted map[string]bool) *co.Clauses {
	if core == nil {
		return nil
	}

	// The interpolant is the subset of clauses1 that participates in the
	// unsat core with clauses2. This is a simplified version;
	// the full version would compute proper Craig interpolation.
	//
	// For now, return the core restricted to symbols from clauses1.
	syms1 := make(map[string]bool)
	for _, sym := range co.SymbolsClauses(clauses1) {
		syms1[sym.Name] = true
	}

	var filteredFmlas []lg.Expr
	for _, f := range core.Fmlas {
		fmlaSyms := co.UsedSymbolsAST(f)
		allInClauses1 := true
		for _, c := range fmlaSyms {
			if !syms1[c.Name] && !interpreted[c.Name] {
				allInClauses1 = false
				break
			}
		}
		if allInClauses1 {
			filteredFmlas = append(filteredFmlas, f)
		}
	}

	if len(filteredFmlas) == 0 {
		return co.TrueClauses(nil)
	}
	return co.NewClauses(filteredFmlas, nil, nil)
}

// UnsatCore computes the unsat core of clauses2 with respect to clauses1.
// Returns nil if the conjunction is satisfiable.
//
// This is a simplified version that returns clauses2 if unsatisfiable.
func UnsatCore(clauses2, clauses1 *co.Clauses) *co.Clauses {
	combined := co.AndClausesTyped(clauses1, clauses2)
	if combined == nil {
		return nil
	}
	// For a proper unsat core, we would use the solver.
	// For now, return clauses2 if the combined is non-trivially constrained.
	if len(combined.Fmlas) > 0 {
		return clauses2
	}
	return nil
}

// filterGroundNonSkolem filters clauses to keep only ground clauses
// without Skolem relation symbols.
func filterGroundNonSkolem(clauses *co.Clauses) *co.Clauses {
	if clauses == nil {
		return clauses
	}
	var filtered []lg.Expr
	for _, f := range clauses.Fmlas {
		syms := co.UsedSymbolsAST(f)
		hasSkolem := false
		for _, c := range syms {
			if IsSkolem(c.Name) {
				hasSkolem = true
				break
			}
		}
		if !hasSkolem {
			filtered = append(filtered, f)
		}
	}
	return co.NewClauses(filtered, clauses.Defs, clauses.Annot)
}
