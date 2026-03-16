// Additional solver utility functions for compatibility checking
// and type system integration.
// Ported from Python's ivy_solver.py.
package solver

import (
	"fmt"

	il "github.com/glycerine/goivy/ivylogic"
	lg "github.com/glycerine/goivy/logic"
)

// CheckNativeCompatSym checks if a symbol's sort is compatible with native Z3 types.
// Returns an error if there's a compatibility issue.
func CheckNativeCompatSym(sig *il.Sig, sym *lg.Const) error {
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
		sym := lg.NewConst(entry.Name, entry.Sort)
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
func (s *Solver) GetArgRange(model *HerbrandModel, x *lg.Const) []lg.Node {
	sort := il.SortRange(x.CSort)
	universe := model.SortUniverse(sort)
	result := make([]lg.Node, len(universe))
	for i, c := range universe {
		result[i] = c
	}
	return result
}

// ModelIfNone returns the provided model, or creates one from clauses if nil.
func (s *Solver) ModelIfNone(clauses interface{}, implied interface{}, model *HerbrandModel) *HerbrandModel {
	if model != nil {
		return model
	}
	// In a full implementation, this would check satisfiability and return a model.
	// For now, return nil.
	return nil
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
func (s *Solver) CollectModelValues(sort lg.Sort, model *HerbrandModel, sym *lg.Const) []*lg.Const {
	if model == nil {
		return nil
	}
	return model.SortUniverse(sort)
}

// NumeralAssign assigns numerals from a model to universe elements.
func NumeralAssign(model *HerbrandModel) map[string]string {
	result := make(map[string]string)
	for _, sort := range model.Sorts() {
		elems := model.SortUniverse(sort)
		for i, elem := range elems {
			result[elem.Name] = fmt.Sprintf("%d", i)
		}
	}
	return result
}

// MineInterpretedConstants extracts interpreted constants from a Z3 model.
// This is called during HerbrandModel construction but also available separately.
func (s *Solver) MineInterpretedConstants(vocab []*lg.Const) map[string][]*lg.Const {
	result := make(map[string][]*lg.Const)
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
				Func:  lg.NewConst("<", il.RelationSort([]lg.Sort{args[0].NodeSort(), args[1].NodeSort()})),
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
				Func:  lg.NewConst("<", il.RelationSort([]lg.Sort{args[1].NodeSort(), args[0].NodeSort()})),
				Terms: []lg.Node{args[1], args[0]},
			}
		}
	case ">=":
		return func(args []lg.Node) lg.Node {
			if len(args) != 2 {
				return nil
			}
			lt := &lg.Apply{
				Func:  lg.NewConst("<", il.RelationSort([]lg.Sort{args[1].NodeSort(), args[0].NodeSort()})),
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
func QuantConstraints(vs []*lg.Var, z3Vs interface{}) lg.Node {
	var constraints []lg.Node
	for _, v := range vs {
		if es, ok := v.VSort.(*lg.EnumeratedSort); ok {
			// Generate: v = e0 | v = e1 | ...
			eqs := make([]lg.Node, len(es.Extension))
			for i, name := range es.Extension {
				eqs[i] = &lg.Eq{T1: v, T2: lg.NewConst(name, v.VSort)}
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
func TypeConstraints(syms []*lg.Const) lg.Node {
	// For each symbol with an enumerated sort, generate range constraints
	var constraints []lg.Node
	for _, sym := range syms {
		sort := il.SortRange(sym.CSort)
		if es, ok := sort.(*lg.EnumeratedSort); ok {
			eqs := make([]lg.Node, len(es.Extension))
			for i, name := range es.Extension {
				eqs[i] = &lg.Eq{T1: sym, T2: lg.NewConst(name, sort)}
			}
			constraints = append(constraints, &lg.Or{Terms: eqs})
		}
	}
	if len(constraints) == 0 {
		return &lg.And{} // true
	}
	return &lg.And{Terms: constraints}
}
