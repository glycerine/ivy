// Additional solver utility functions for compatibility checking
// and type system integration.
// Ported from Python's ivy_solver.py.
package solver

import (
	"fmt"

	"github.com/glycerine/goivy/clauseops"
	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
)

// CheckNativeCompatSym checks if a symbol's sort is compatible with native Z3 types.
// Returns an error if there's a compatibility issue.
func CheckNativeCompatSym(sig *il.Sig, sym *lg.Symbol) error {
	if !il.IsFunctionSort(sym.CSort) {
		return nil // non-function sorts are always compatible
	}
	fs, ok := sym.CSort.(*lg.FunctionSort)
	if !ok {
		return nil
	}
	for _, d := range fs.Domain() {
		if err := checkSortCompat(sig, d); err != nil {
			return fmt.Errorf("symbol %s: domain sort %s: %w", sym.Name, d, err)
		}
	}
	if err := checkSortCompat(sig, fs.Range()); err != nil {
		return fmt.Errorf("symbol %s: range sort %s: %w", sym.Name, fs.Range(), err)
	}
	return nil
}

func checkSortCompat(sig *il.Sig, s lg.Sort) error {
	// Check that the sort either has no interpretation or a compatible one
	name := il.SortName(s)
	interp, ok := sig.Interp[name]
	if !ok {
		return nil // no interpretation = OK
	}
	switch v := interp.(type) {
	case string:
		if IsSolverSort(v) || v == "int" || v == "nat" {
			return nil
		}
		_, _, ok := ParseArrayTheory(v)
		if ok {
			return nil
		}
		return fmt.Errorf("unsupported native type: %s", v)
	}
	return nil
}

// CheckCompat checks all symbols in the signature for native compatibility.
func CheckCompat(sig *il.Sig) []error {
	var errs []error
	for _, entry := range sig.Symbols {
		sym := lg.NewSymbol(entry.Name, entry.Sort)
		if err := CheckNativeCompatSym(sig, sym); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// TermsMatch checks if two term lists structurally match.
func TermsMatch(tl1, tl2 []lg.Node) bool {
	if len(tl1) != len(tl2) {
		return false
	}
	for i := range tl1 {
		if !tl1[i].Equal(tl2[i]) {
			return false
		}
	}
	return true
}

// GetArgRange returns the range of argument values from a model for a function symbol.
func (s *Solver) GetArgRange(model *HerbrandModel, x *lg.Symbol) []lg.Node {
	sort := il.SortRange(x.CSort)
	universe := model.SortUniverse(sort)
	result := make([]lg.Node, len(universe))
	for i, c := range universe {
		result[i] = c
	}
	return result
}

// ModelIfNone returns the provided model, or creates one from clauses if nil.
// If model is nil, it performs the incremental sort-size search to find a
// small model of the clauses (optionally conjoined with implied).
// Corresponds to Python's model_if_none.
func (s *Solver) ModelIfNone(clauses *clauseops.Clauses, implied *clauseops.Clauses, model *HerbrandModel) *HerbrandModel {
	if model != nil {
		return model
	}
	// Build clauses to check: clauses AND implied
	var combined *clauseops.Clauses
	if implied != nil {
		combined = clauseops.AndClausesTyped(clauses, implied)
	} else {
		combined = clauses
	}

	// Try to find a model using GetSmallModel
	mr, err := s.GetSmallModel(combined, nil, nil)
	if err != nil || mr == nil {
		return nil // UNSAT or error
	}

	// Collect vocabulary from clauses
	symSet := clauses.Symbols()
	if implied != nil {
		for sKey, sNode := range implied.Symbols() {
			symSet[sKey] = sNode
		}
	}
	vocab := make([]*lg.Symbol, 0, len(symSet))
	for _, sym := range symSet {
		if c, ok := sym.(*lg.Symbol); ok {
			vocab = append(vocab, c)
		}
	}

	return NewHerbrandModel(s, mr.Solver, mr.Model, vocab)
}

// ClauseModelSimp simplifies a clause using a model.
// Removes literals that are false in the model.
func (s *Solver) ClauseModelSimp(model *HerbrandModel, clause lg.Node) lg.Node {
	if model == nil {
		return clause
	}
	or, ok := clause.(*lg.Or)
	if !ok {
		return clause
	}
	var kept []lg.Node
	for _, term := range or.Terms {
		if model.Eval(term) {
			kept = append(kept, term)
		}
	}
	if len(kept) == 0 {
		return &lg.Or{} // false
	}
	if len(kept) == 1 {
		return kept[0]
	}
	return &lg.Or{Terms: kept}
}

// CollectModelValues collects all model values for a symbol of a given sort.
func (s *Solver) CollectModelValues(sort lg.Sort, model *HerbrandModel, sym *lg.Symbol) []*lg.Symbol {
	if model == nil {
		return nil
	}
	return model.SortUniverse(sort)
}

// NumeralAssign assigns numeral names to universe elements, respecting existing
// numerals in the clause set.
//
// For each sort:
//  1. First assigns existing numerals from the clause set to their model values.
//  2. Then assigns fresh numeral names (0, 1, 2, ...) to remaining elements,
//     skipping names already used by existing numerals.
//
// Returns a map from model element names to numeral names.
//
// Corresponds to Python numeral_assign (lines 1338-1371).
func NumeralAssign(model *HerbrandModel) map[string]string {
	return NumeralAssignWithClauses(model, nil)
}

// NumeralAssignWithClauses is the full version that respects existing numerals.
func NumeralAssignWithClauses(model *HerbrandModel, clauses *clauseops.Clauses) map[string]string {
	result := make(map[string]string)

	// Collect existing numerals from clauses, grouped by sort
	numBySort := make(map[string][]*lg.Symbol)
	if clauses != nil {
		usedConsts := clauseops.ConstantsClauses(clauses)
		for _, c := range usedConsts {
			if il.IsNumeral(c) {
				sortName := il.SortName(il.SortRange(c.CSort))
				numBySort[sortName] = append(numBySort[sortName], c)
			}
		}
	}

	for _, sort := range model.Sorts() {
		sortName := il.SortName(sort)

		// Skip interpreted sorts
		if il.IsInterpretedSort(model.sig, sort) {
			continue
		}

		// First pass: assign existing numerals to their model values
		usedNumerals := make(map[string]bool)
		foom := make(map[string]*lg.Symbol) // model element → numeral

		for _, num := range numBySort[sortName] {
			modelVal := model.EvalConstant(num)
			if modelVal != nil {
				if _, already := foom[modelVal.Name]; already {
					// Two numerals assigned same value — warn but continue
					continue
				}
				foom[modelVal.Name] = num
				usedNumerals[num.Name] = true
			}
		}

		// Second pass: assign fresh numerals to remaining elements
		elems := model.SortedSortUniverse(sort)
		i := 0
		for _, c := range elems {
			if _, assigned := foom[c.Name]; !assigned {
				// Find next unused numeral name
				for {
					name := fmt.Sprintf("%d", i)
					i++
					numConst := lg.NewSymbol(name, sort)
					if !usedNumerals[numConst.Name] {
						foom[c.Name] = numConst
						break
					}
				}
			}
		}

		// Copy into result
		for elemName, numConst := range foom {
			result[elemName] = numConst.Name
		}
	}

	return result
}

// MineInterpretedConstants extracts interpreted constants from a Z3 model.
// This is called during HerbrandModel construction but also available separately.
func (s *Solver) MineInterpretedConstants(vocab []*lg.Symbol) map[string][]*lg.Symbol {
	result := make(map[string][]*lg.Symbol)
	for _, c := range vocab {
		sortName := il.SortName(il.SortRange(c.CSort))
		if !il.IsInterpretedSort(s.sig, c.CSort) {
			continue
		}
		result[sortName] = append(result[sortName], c)
	}
	return result
}

// GetPolymacs returns polymorphic macros for an operator.
func GetPolymacs(op string) func([]lg.Node) lg.Node {
	switch op {
	case "<=":
		return func(args []lg.Node) lg.Node {
			if len(args) != 2 {
				return nil
			}
			lt := &lg.Apply{
				Func:  lg.NewSymbol("<", il.RelationSort([]lg.Sort{args[0].NodeSort(), args[1].NodeSort()})),
				Terms: args,
			}
			eq := &lg.Eq{T1: args[0], T2: args[1]}
			return &lg.Or{Terms: []lg.Node{lt, eq}}
		}
	case ">":
		return func(args []lg.Node) lg.Node {
			if len(args) != 2 {
				return nil
			}
			return &lg.Apply{
				Func:  lg.NewSymbol("<", il.RelationSort([]lg.Sort{args[1].NodeSort(), args[0].NodeSort()})),
				Terms: []lg.Node{args[1], args[0]},
			}
		}
	case ">=":
		return func(args []lg.Node) lg.Node {
			if len(args) != 2 {
				return nil
			}
			lt := &lg.Apply{
				Func:  lg.NewSymbol("<", il.RelationSort([]lg.Sort{args[1].NodeSort(), args[0].NodeSort()})),
				Terms: []lg.Node{args[1], args[0]},
			}
			eq := &lg.Eq{T1: args[0], T2: args[1]}
			return &lg.Or{Terms: []lg.Node{lt, eq}}
		}
	}
	return nil
}

// QuantConstraints generates sort constraints for quantifier variables.
// For finite/enumerated sorts, generates membership constraints.
func QuantConstraints(vs []*lg.Variable, z3Vs interface{}) lg.Node {
	var constraints []lg.Node
	for _, v := range vs {
		if es, ok := v.VSort.(*lg.EnumeratedSort); ok {
			// Generate: v = e0 | v = e1 | ...
			eqs := make([]lg.Node, len(es.Extension))
			for i, name := range es.Extension {
				eqs[i] = &lg.Eq{T1: v, T2: lg.NewSymbol(name, v.VSort)}
			}
			constraints = append(constraints, &lg.Or{Terms: eqs})
		}
	}
	if len(constraints) == 0 {
		return &lg.And{} // true
	}
	return &lg.And{Terms: constraints}
}

// TypeConstraints generates type constraints for a set of symbols.
func TypeConstraints(syms []*lg.Symbol) lg.Node {
	// For each symbol with an enumerated sort, generate range constraints
	var constraints []lg.Node
	for _, sym := range syms {
		sort := il.SortRange(sym.CSort)
		if es, ok := sort.(*lg.EnumeratedSort); ok {
			eqs := make([]lg.Node, len(es.Extension))
			for i, name := range es.Extension {
				eqs[i] = &lg.Eq{T1: sym, T2: lg.NewSymbol(name, sort)}
			}
			constraints = append(constraints, &lg.Or{Terms: eqs})
		}
	}
	if len(constraints) == 0 {
		return &lg.And{} // true
	}
	return &lg.And{Terms: constraints}
}
